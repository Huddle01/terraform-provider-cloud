package provider

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &instanceResource{}
var _ resource.ResourceWithConfigure = &instanceResource{}
var _ resource.ResourceWithImportState = &instanceResource{}

type instanceResource struct {
	client *apiClient
}

type instanceResourceModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FlavorID           types.String  `tfsdk:"flavor_id"`
	ImageID            types.String  `tfsdk:"image_id"`
	BootDiskSize       types.Int64   `tfsdk:"boot_disk_size"`
	KeyNames           types.List    `tfsdk:"key_names"`
	SecurityGroupNames types.List    `tfsdk:"security_group_names"`
	Tags               types.List    `tfsdk:"tags"`
	AssignPublicIP     types.Bool    `tfsdk:"assign_public_ip"`
	NetworkID          types.String  `tfsdk:"network_id"`
	PowerState         types.String  `tfsdk:"power_state"`
	Region             types.String  `tfsdk:"region"`
	Status             types.String  `tfsdk:"status"`
	VCPUs              types.Float64 `tfsdk:"vcpus"`
	RAM                types.Float64 `tfsdk:"ram"`
	CreatedAt          types.String  `tfsdk:"created_at"`
	PrivateIPv4        types.String  `tfsdk:"private_ipv4"`
	PublicIPv4         types.String  `tfsdk:"public_ipv4"`
}

func NewInstanceResource() resource.Resource {
	return &instanceResource{}
}

func (r *instanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *instanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Huddle01 Cloud virtual machine instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the instance.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable name for the instance.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"flavor_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the flavor (hardware profile) to use. Use the `huddle_cloud_flavors` data source to list available flavors.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"image_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the OS image to boot from. Use the `huddle_cloud_images` data source to list available images.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"boot_disk_size": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Size of the boot disk in GB. Defaults to `30`.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"key_names": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of keypair names to inject into the instance for SSH access.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"security_group_names": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of security group names to attach to the instance.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Map of key/value tags to apply to the instance.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"assign_public_ip": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to assign a public IPv4 address. Defaults to `true`.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"network_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the network to attach the instance to. If omitted, the workspace default network is used.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"power_state": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Desired power state of the instance. One of `active`, `stopped`, `paused`, `suspended`. Defaults to `active`.",
				Validators: []validator.String{
					stringvalidator.OneOf("active", "stopped", "paused", "suspended"),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Region in which to create the instance. Defaults to the provider-level region.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current status of the instance as reported by the API.",
			},
			"vcpus": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of virtual CPUs allocated to the instance.",
			},
			"ram": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Amount of RAM allocated to the instance in MB.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the instance was created (RFC 3339).",
			},
			"private_ipv4": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Private IPv4 address of the instance.",
			},
			"public_ipv4": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Public IPv4 address of the instance, if assigned.",
			},
		},
	}
}

func (r *instanceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*apiClient)
}

func (r *instanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan instanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	originalPlan := plan

	region, err := effectiveRegion(r.client, plan.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	body := map[string]any{
		"name":             plan.Name.ValueString(),
		"flavor_id":        plan.FlavorID.ValueString(),
		"image_id":         plan.ImageID.ValueString(),
		"boot_disk_size":   plan.BootDiskSize.ValueInt64(),
		"key_name":         listStringToSlice(plan.KeyNames),
		"sg_names":         listStringToSlice(plan.SecurityGroupNames),
		"assign_public_ip": boolOrDefault(plan.AssignPublicIP, true),
	}
	if tags := listStringToSlice(plan.Tags); len(tags) > 0 {
		body["tags"] = tags
	}

	opKey := generateIdempotencyKey()
	var created createInstanceResponse
	if err := r.client.post(ctx, "/instances", queryWithRegion(region), body, opKey, &created); err != nil {
		resp.Diagnostics.AddError("Create instance failed", describeAPIError(err))
		return
	}

	instanceID := created.ID
	if instanceID == "" {
		resp.Diagnostics.AddError("Create instance failed", "API response missing instance id")
		return
	}

	if networkID := stringOrEmpty(plan.NetworkID); networkID != "" {
		attachBody := map[string]any{"network_id": networkID}
		path := "/instances/" + url.PathEscape(instanceID) + "/networks"
		if err := r.client.post(ctx, path, queryWithRegion(region), attachBody, generateIdempotencyKey(), nil); err != nil {
			resp.Diagnostics.AddError("Attach network failed", describeAPIError(err))
			return
		}
	}

	if err := waitForInstanceStatus(ctx, r.client, instanceID, region, 15*time.Minute, "ACTIVE"); err != nil {
		resp.Diagnostics.AddError("Instance provisioning timed out", err.Error())
		return
	}

	// Floating IP assignment is asynchronous — the instance reaches ACTIVE
	// before the association completes. Poll until public_ipv4 is populated so
	// Terraform state reflects the correct value immediately after create.
	if boolOrDefault(originalPlan.AssignPublicIP, true) {
		if _, err := waitForFloatingIP(ctx, r.client, instanceID, region); err != nil {
			resp.Diagnostics.AddWarning(
				"Floating IP not assigned in time",
				"Instance is ACTIVE but public_ipv4 was not populated within 3 minutes. "+
					"It may appear in state after the next refresh.",
			)
		}
	}

	finalState, err := r.readInstance(ctx, instanceID, region)
	if err != nil {
		resp.Diagnostics.AddError("Read instance after create failed", describeAPIError(err))
		return
	}

	plan = *finalState
	plan.FlavorID = originalPlan.FlavorID
	plan.ImageID = originalPlan.ImageID
	plan.BootDiskSize = originalPlan.BootDiskSize
	plan.KeyNames = originalPlan.KeyNames
	plan.SecurityGroupNames = originalPlan.SecurityGroupNames
	plan.Tags = originalPlan.Tags
	plan.AssignPublicIP = originalPlan.AssignPublicIP
	plan.NetworkID = originalPlan.NetworkID
	plan.PowerState = originalPlan.PowerState
	plan.Region = types.StringValue(region)
	if plan.PowerState.IsNull() || plan.PowerState.IsUnknown() {
		if ps := stringOrEmpty(plan.PowerState); ps == "" {
			plan.PowerState = types.StringValue(powerStateFromStatus(plan.Status.ValueString()))
		}
	}

	desired := stringOrEmpty(plan.PowerState)
	if desired != "" {
		if err := r.ensurePowerState(ctx, instanceID, region, plan.Status.ValueString(), desired); err != nil {
			resp.Diagnostics.AddError("Set instance power_state failed", err.Error())
			return
		}
		finalState, err = r.readInstance(ctx, instanceID, region)
		if err != nil {
			resp.Diagnostics.AddError("Read instance after power_state update failed", describeAPIError(err))
			return
		}
		plan = *finalState
		plan.FlavorID = originalPlan.FlavorID
		plan.ImageID = originalPlan.ImageID
		plan.BootDiskSize = originalPlan.BootDiskSize
		plan.KeyNames = originalPlan.KeyNames
		plan.SecurityGroupNames = originalPlan.SecurityGroupNames
		plan.Tags = originalPlan.Tags
		plan.AssignPublicIP = originalPlan.AssignPublicIP
		plan.NetworkID = originalPlan.NetworkID
		plan.Region = types.StringValue(region)
		plan.PowerState = types.StringValue(desired)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *instanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	next, err := r.readInstance(ctx, state.ID.ValueString(), region)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read instance failed", describeAPIError(err))
		return
	}

	next.Region = types.StringValue(region)
	if state.PowerState.IsNull() || state.PowerState.IsUnknown() {
		next.PowerState = types.StringValue(powerStateFromStatus(next.Status.ValueString()))
	} else {
		next.PowerState = state.PowerState
	}
	next.NetworkID = state.NetworkID
	// AssignPublicIP, BootDiskSize, KeyNames, SecurityGroupNames are write-only inputs
	// not returned by the API — preserve from prior state.
	next.AssignPublicIP = state.AssignPublicIP
	next.KeyNames = state.KeyNames
	next.SecurityGroupNames = state.SecurityGroupNames
	next.Tags = state.Tags
	next.BootDiskSize = state.BootDiskSize
	// FlavorID and ImageID come from the API response; fall back to state only if
	// the API returned an empty string (e.g. older server versions).
	if next.FlavorID.ValueString() == "" {
		next.FlavorID = state.FlavorID
	}
	if next.ImageID.ValueString() == "" {
		next.ImageID = state.ImageID
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, next)...)
}

func (r *instanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan instanceResourceModel
	var state instanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, plan.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	current, err := r.readInstance(ctx, state.ID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Read instance before update failed", describeAPIError(err))
		return
	}

	desiredPower := stringOrEmpty(plan.PowerState)
	if desiredPower == "" {
		desiredPower = powerStateFromStatus(current.Status.ValueString())
	}

	if err := r.ensurePowerState(ctx, state.ID.ValueString(), region, current.Status.ValueString(), desiredPower); err != nil {
		resp.Diagnostics.AddError("Update power_state failed", err.Error())
		return
	}

	updated, err := r.readInstance(ctx, state.ID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Read instance after update failed", describeAPIError(err))
		return
	}

	updated.Region = types.StringValue(region)
	updated.PowerState = types.StringValue(desiredPower)
	updated.NetworkID = state.NetworkID
	updated.AssignPublicIP = state.AssignPublicIP
	updated.KeyNames = state.KeyNames
	updated.SecurityGroupNames = state.SecurityGroupNames
	updated.Tags = state.Tags
	updated.FlavorID = state.FlavorID
	updated.ImageID = state.ImageID
	updated.BootDiskSize = state.BootDiskSize

	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

func (r *instanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state instanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	path := "/instances/" + url.PathEscape(state.ID.ValueString())
	opKey := generateIdempotencyKey()
	err = r.client.delete(ctx, path, queryWithRegion(region), opKey, nil)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete instance failed", describeAPIError(err))
		return
	}

	waitErr := waitForCondition(ctx, 15*time.Minute, 5*time.Second, func(c context.Context) (bool, error) {
		var payload instanceResponseEnvelope
		err := r.client.get(c, path, queryWithRegion(region), &payload)
		if err != nil {
			if isNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
	if waitErr != nil {
		resp.Diagnostics.AddError("Delete instance timed out", waitErr.Error())
	}
}

func (r *instanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	if r.client.defaultRegion != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), r.client.defaultRegion)...)
	}
}

func (r *instanceResource) readInstance(ctx context.Context, id, region string) (*instanceResourceModel, error) {
	var out instanceResponseEnvelope
	err := r.client.get(ctx, "/instances/"+url.PathEscape(id), queryWithRegion(region), &out)
	if err != nil {
		return nil, err
	}

	privateIP, publicIP := extractIPv4(out.Instance.Networks)
	model := &instanceResourceModel{
		ID:          types.StringValue(out.Instance.ID),
		Name:        types.StringValue(out.Instance.Name),
		FlavorID:    types.StringValue(out.Instance.FlavorID),
		ImageID:     types.StringValue(out.Instance.Image.ID),
		Status:      types.StringValue(out.Instance.Status),
		VCPUs:       types.Float64Value(out.Instance.VCPUs),
		RAM:         types.Float64Value(out.Instance.RAM),
		CreatedAt:   types.StringValue(out.Instance.CreatedAt),
		PrivateIPv4: types.StringValue(privateIP),
		PublicIPv4:  types.StringValue(publicIP),
	}
	return model, nil
}

func extractIPv4(n instanceNetworks) (private string, public string) {
	for _, item := range n.V4 {
		switch strings.ToLower(item.Type) {
		case "floating":
			public = item.IPAddress
		default:
			if private == "" {
				private = item.IPAddress
			}
		}
	}
	return private, public
}

func powerStateFromStatus(status string) string {
	switch strings.ToUpper(status) {
	case "SHUTOFF":
		return "stopped"
	case "PAUSED":
		return "paused"
	case "SUSPENDED":
		return "suspended"
	default:
		return "active"
	}
}

func (r *instanceResource) ensurePowerState(ctx context.Context, id string, region string, currentStatus string, desired string) error {
	current := strings.ToUpper(currentStatus)
	desired = strings.ToLower(desired)

	var action string
	switch desired {
	case "active":
		switch current {
		case "ACTIVE":
			return nil
		case "SHUTOFF":
			action = "start"
		case "PAUSED":
			action = "unpause"
		case "SUSPENDED":
			action = "resume"
		default:
			action = "start"
		}
	case "stopped":
		if current == "SHUTOFF" {
			return nil
		}
		action = "stop"
	case "paused":
		if current == "PAUSED" {
			return nil
		}
		action = "pause"
	case "suspended":
		if current == "SUSPENDED" {
			return nil
		}
		action = "suspend"
	default:
		return nil
	}

	body := map[string]any{"action": action}
	path := "/instances/" + url.PathEscape(id) + "/action"
	if err := r.client.post(ctx, path, queryWithRegion(region), body, generateIdempotencyKey(), nil); err != nil {
		return err
	}

	target := "ACTIVE"
	switch desired {
	case "stopped":
		target = "SHUTOFF"
	case "paused":
		target = "PAUSED"
	case "suspended":
		target = "SUSPENDED"
	}

	return waitForCondition(ctx, 10*time.Minute, 5*time.Second, func(c context.Context) (bool, error) {
		var payload instanceResponseEnvelope
		err := r.client.get(c, "/instances/"+url.PathEscape(id), queryWithRegion(region), &payload)
		if err != nil {
			return false, err
		}
		return strings.EqualFold(payload.Instance.Status, target), nil
	})
}
