package provider

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &securityGroupResource{}
var _ resource.ResourceWithConfigure = &securityGroupResource{}
var _ resource.ResourceWithImportState = &securityGroupResource{}

type securityGroupResource struct {
	client *apiClient
}

type securityGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Region      types.String `tfsdk:"region"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	Rules       types.List   `tfsdk:"rules"`
}

type securityRuleStateModel struct {
	ID             types.String `tfsdk:"id"`
	Direction      types.String `tfsdk:"direction"`
	EtherType      types.String `tfsdk:"ether_type"`
	Protocol       types.String `tfsdk:"protocol"`
	PortRangeMin   types.Int64  `tfsdk:"port_range_min"`
	PortRangeMax   types.Int64  `tfsdk:"port_range_max"`
	RemoteIPPrefix types.String `tfsdk:"remote_ip_prefix"`
	RemoteGroupID  types.String `tfsdk:"remote_group_id"`
}

func NewSecurityGroupResource() resource.Resource {
	return &securityGroupResource{}
}

func (r *securityGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *securityGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
			"rules": schema.ListNestedAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":               schema.StringAttribute{Computed: true},
						"direction":        schema.StringAttribute{Computed: true},
						"ether_type":       schema.StringAttribute{Computed: true},
						"protocol":         schema.StringAttribute{Computed: true},
						"port_range_min":   schema.Int64Attribute{Computed: true},
						"port_range_max":   schema.Int64Attribute{Computed: true},
						"remote_ip_prefix": schema.StringAttribute{Computed: true},
						"remote_group_id":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (r *securityGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*apiClient)
}

func (r *securityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityGroupResourceModel
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
	if d := stringOrEmpty(plan.Description); d != "" {
		body["description"] = d
	}

	var out createSecurityGroupEnvelope
	if err := r.client.post(ctx, "/security-groups", queryWithRegion(region), body, generateIdempotencyKey(), &out); err != nil {
		resp.Diagnostics.AddError("Create security group failed", describeAPIError(err))
		return
	}

	// The create endpoint does not return the default rules that the backend
	// provisions alongside the security group. Fetch the full record so state
	// matches what a subsequent Read (or import) would return — otherwise
	// ImportStateVerify diffs the rules list against the empty post-create state.
	var detail securityGroupDetailEnvelope
	if err := r.client.get(ctx, "/security-groups/"+url.PathEscape(out.ID), queryWithRegion(region), &detail); err != nil {
		resp.Diagnostics.AddError("Read security group after create failed", describeAPIError(err))
		return
	}

	plan.ID = types.StringValue(out.ID)
	plan.Region = types.StringValue(region)
	plan.CreatedAt = types.StringValue(detail.SecurityGroup.CreatedAt)
	if detail.SecurityGroup.UpdatedAt == "" {
		plan.UpdatedAt = types.StringNull()
	} else {
		plan.UpdatedAt = types.StringValue(detail.SecurityGroup.UpdatedAt)
	}

	rules, diags := flattenSecurityGroupRules(detail.SecurityGroup.Rules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Rules = rules

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	var out securityGroupDetailEnvelope
	err = r.client.get(ctx, "/security-groups/"+url.PathEscape(state.ID.ValueString()), queryWithRegion(region), &out)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read security group failed", describeAPIError(err))
		return
	}

	state.Name = types.StringValue(out.SecurityGroup.Name)
	// description is Optional (not Computed). Storing "" when the user omitted
	// the field causes a plan diff that triggers RequiresReplace. Normalize to
	// null so state stays in sync with an unset config attribute.
	if out.SecurityGroup.Description == "" {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(out.SecurityGroup.Description)
	}
	state.Region = types.StringValue(region)
	state.CreatedAt = types.StringValue(out.SecurityGroup.CreatedAt)
	state.UpdatedAt = types.StringValue(out.SecurityGroup.UpdatedAt)

	rules, diags := flattenSecurityGroupRules(out.SecurityGroup.Rules)
	resp.Diagnostics.Append(diags...)
	state.Rules = rules

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported update", "Security group updates are not supported by API. Recreate the resource.")
}

func (r *securityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	err = r.client.delete(ctx, "/security-groups/"+url.PathEscape(state.ID.ValueString()), queryWithRegion(region), generateIdempotencyKey(), nil)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete security group failed", describeAPIError(err))
	}
}

func (r *securityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	if r.client.defaultRegion != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), r.client.defaultRegion)...)
	}
}

func flattenSecurityGroupRules(in []securityRuleItem) (types.List, diag.Diagnostics) {
	attrTypes := map[string]attr.Type{
		"id":               types.StringType,
		"direction":        types.StringType,
		"ether_type":       types.StringType,
		"protocol":         types.StringType,
		"port_range_min":   types.Int64Type,
		"port_range_max":   types.Int64Type,
		"remote_ip_prefix": types.StringType,
		"remote_group_id":  types.StringType,
	}

	rules := make([]types.Object, 0, len(in))
	for _, rule := range in {
		obj, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id":               types.StringValue(rule.ID),
			"direction":        types.StringValue(rule.Direction),
			"ether_type":       types.StringValue(rule.EffectiveEtherType()),
			"protocol":         types.StringValue(rule.Protocol),
			"port_range_min":   ptrInt64ToTerraform(rule.PortRangeMin),
			"port_range_max":   ptrInt64ToTerraform(rule.PortRangeMax),
			"remote_ip_prefix": types.StringValue(rule.RemoteIPPrefix),
			"remote_group_id":  types.StringValue(rule.RemoteGroupID),
		})
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
		}
		rules = append(rules, obj)
	}

	return types.ListValueFrom(context.Background(), types.ObjectType{AttrTypes: attrTypes}, rules)
}
