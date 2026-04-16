package provider

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &networkResource{}
var _ resource.ResourceWithConfigure = &networkResource{}
var _ resource.ResourceWithImportState = &networkResource{}

type networkResource struct {
	client *apiClient
}

type networkResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	PoolCIDR          types.String `tfsdk:"pool_cidr"`
	PrimarySubnetCIDR types.String `tfsdk:"primary_subnet_cidr"`
	PrimarySubnetSize types.Int64  `tfsdk:"primary_subnet_size"`
	NoGateway         types.Bool   `tfsdk:"no_gateway"`
	EnableDHCP        types.Bool   `tfsdk:"enable_dhcp"`
	Region            types.String `tfsdk:"region"`
	Status            types.String `tfsdk:"status"`
	Subnets           types.List   `tfsdk:"subnets"`
	AdminStateUp      types.Bool   `tfsdk:"admin_state_up"`
}

func NewNetworkResource() resource.Resource {
	return &networkResource{}
}

func (r *networkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *networkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a private network and its primary subnet.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the network.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable name for the network.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional description for the network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pool_cidr": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "CIDR block of the allocation pool for the subnet (e.g. `10.0.0.0/8`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"primary_subnet_cidr": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "CIDR block for the primary subnet (e.g. `192.168.1.0/24`). Takes precedence over `primary_subnet_size`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"primary_subnet_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Prefix length used to auto-allocate the primary subnet from `pool_cidr` (e.g. `24` allocates a `/24`).",
			},
			"no_gateway": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "When `true`, no gateway is created for the subnet. Defaults to `false`.",
			},
			"enable_dhcp": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether DHCP is enabled on the primary subnet. Defaults to `true`.",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Region in which to create the network. Defaults to the provider-level region.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current status of the network (`ACTIVE`, `DOWN`, etc.).",
			},
			"subnets": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of subnet IDs associated with this network.",
			},
			"admin_state_up": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the network is administratively up.",
			},
		},
	}
}

func (r *networkResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*apiClient)
}

func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan networkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, plan.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	body := map[string]any{
		"name": plan.Name.ValueString(),
	}
	if v := stringOrEmpty(plan.Description); v != "" {
		body["description"] = v
	}
	if v := stringOrEmpty(plan.PoolCIDR); v != "" {
		body["pool_cidr"] = v
	}
	if v := stringOrEmpty(plan.PrimarySubnetCIDR); v != "" {
		body["primary_subnet_cidr"] = v
	}
	if !plan.PrimarySubnetSize.IsNull() && !plan.PrimarySubnetSize.IsUnknown() {
		body["primary_subnet_size"] = plan.PrimarySubnetSize.ValueInt64()
	}
	if !plan.NoGateway.IsNull() && !plan.NoGateway.IsUnknown() {
		body["no_gateway"] = plan.NoGateway.ValueBool()
	}
	if !plan.EnableDHCP.IsNull() && !plan.EnableDHCP.IsUnknown() {
		body["enable_dhcp"] = plan.EnableDHCP.ValueBool()
	}

	var out networkEnvelope
	if err := r.client.post(ctx, "/networks", queryWithRegion(region), body, generateIdempotencyKey(), &out); err != nil {
		resp.Diagnostics.AddError("Create network failed", describeAPIError(err))
		return
	}

	plan.ID = types.StringValue(out.Network.ID)
	plan.Region = types.StringValue(region)
	plan.Status = types.StringValue(out.Network.Status)
	plan.AdminStateUp = types.BoolValue(out.Network.AdminStateUp)
	plan.Subnets, _ = types.ListValueFrom(ctx, types.StringType, out.Network.Subnets)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	q := queryWithRegion(region)
	q.Set("owned", "true")

	var out networkListEnvelope
	if err := r.client.get(ctx, "/networks", q, &out); err != nil {
		resp.Diagnostics.AddError("Read networks failed", describeAPIError(err))
		return
	}

	var found *networkPayload
	for i := range out.Data.Networks {
		if out.Data.Networks[i].ID == state.ID.ValueString() {
			found = &out.Data.Networks[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(found.Name)
	state.Status = types.StringValue(found.Status)
	state.AdminStateUp = types.BoolValue(found.AdminStateUp)
	state.Subnets, _ = types.ListValueFrom(ctx, types.StringType, found.Subnets)
	state.Region = types.StringValue(region)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported update", "Network updates are not supported by API. Recreate the resource.")
}

func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state networkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	err = r.client.delete(ctx, "/networks/"+url.PathEscape(state.ID.ValueString()), queryWithRegion(region), generateIdempotencyKey(), nil)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete network failed", describeAPIError(err))
	}
}

func (r *networkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	if r.client.defaultRegion != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), r.client.defaultRegion)...)
	}
}
