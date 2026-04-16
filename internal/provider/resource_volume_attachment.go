package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &volumeAttachmentResource{}
var _ resource.ResourceWithConfigure = &volumeAttachmentResource{}
var _ resource.ResourceWithImportState = &volumeAttachmentResource{}

type volumeAttachmentResource struct {
	client *apiClient
}

type volumeAttachmentResourceModel struct {
	ID         types.String `tfsdk:"id"`
	VolumeID   types.String `tfsdk:"volume_id"`
	InstanceID types.String `tfsdk:"instance_id"`
	Region     types.String `tfsdk:"region"`
	Device     types.String `tfsdk:"device"`
}

func NewVolumeAttachmentResource() resource.Resource {
	return &volumeAttachmentResource{}
}

func (r *volumeAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_attachment"
}

func (r *volumeAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"volume_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"instance_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"device": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *volumeAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*apiClient)
}

func (r *volumeAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeAttachmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, plan.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	vol, err := r.readVolumeDetail(ctx, plan.VolumeID.ValueString(), region)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Read volume before attach failed", describeAPIError(err))
		return
	}
	if vol != nil {
		if matched := findAttachmentByInstance(vol.Attachments, plan.InstanceID.ValueString()); matched != nil {
			plan.ID = types.StringValue(volumeAttachmentStateID(plan.VolumeID.ValueString(), plan.InstanceID.ValueString()))
			plan.Region = types.StringValue(region)
			plan.Device = types.StringValue(matched.Device)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}
	}

	body := map[string]any{"instance_id": plan.InstanceID.ValueString()}
	path := "/volumes/" + url.PathEscape(plan.VolumeID.ValueString()) + "/attach"
	if err := r.client.post(ctx, path, queryWithRegion(region), body, generateIdempotencyKey(), nil); err != nil {
		resp.Diagnostics.AddError("Attach volume failed", describeAPIError(err))
		return
	}

	if err := waitForCondition(ctx, 3*time.Minute, 2*time.Second, func(c context.Context) (bool, error) {
		refreshed, readErr := r.readVolumeDetail(c, plan.VolumeID.ValueString(), region)
		if readErr != nil {
			return false, readErr
		}
		return findAttachmentByInstance(refreshed.Attachments, plan.InstanceID.ValueString()) != nil, nil
	}); err != nil {
		resp.Diagnostics.AddError("Attach volume timed out", err.Error())
		return
	}

	refreshed, err := r.readVolumeDetail(ctx, plan.VolumeID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Read volume after attach failed", describeAPIError(err))
		return
	}

	matched := findAttachmentByInstance(refreshed.Attachments, plan.InstanceID.ValueString())
	if matched == nil {
		resp.Diagnostics.AddError("Attach volume failed", "volume is attached but attachment details were not returned by API")
		return
	}

	plan.ID = types.StringValue(volumeAttachmentStateID(plan.VolumeID.ValueString(), plan.InstanceID.ValueString()))
	plan.Region = types.StringValue(region)
	plan.Device = types.StringValue(matched.Device)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeAttachmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	volume, err := r.readVolumeDetail(ctx, state.VolumeID.ValueString(), region)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read volume attachment failed", describeAPIError(err))
		return
	}

	matched := findAttachmentByInstance(volume.Attachments, state.InstanceID.ValueString())
	if matched == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(volumeAttachmentStateID(state.VolumeID.ValueString(), state.InstanceID.ValueString()))
	state.Region = types.StringValue(region)
	state.Device = types.StringValue(matched.Device)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported update", "Volume attachment updates are not supported by API. Recreate the resource.")
}

func (r *volumeAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeAttachmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	volume, err := r.readVolumeDetail(ctx, state.VolumeID.ValueString(), region)
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Read volume before detach failed", describeAPIError(err))
		return
	}
	if findAttachmentByInstance(volume.Attachments, state.InstanceID.ValueString()) == nil {
		return
	}

	body := map[string]any{"instance_id": state.InstanceID.ValueString()}
	path := "/volumes/" + url.PathEscape(state.VolumeID.ValueString()) + "/detach"
	if err := r.client.post(ctx, path, queryWithRegion(region), body, generateIdempotencyKey(), nil); err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Detach volume failed", describeAPIError(err))
		return
	}
}

func (r *volumeAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	volumeID, instanceID := parseVolumeAttachmentImportID(req.ID)
	if volumeID == "" || instanceID == "" {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected <volume_id>/<instance_id>, got %q", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), volumeAttachmentStateID(volumeID, instanceID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("volume_id"), volumeID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_id"), instanceID)...)
	if r.client.defaultRegion != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), r.client.defaultRegion)...)
	}
}

func (r *volumeAttachmentResource) readVolumeDetail(ctx context.Context, volumeID, region string) (*volumePayload, error) {
	var out volumeDetailEnvelope
	if err := r.client.get(ctx, "/volumes/"+url.PathEscape(volumeID), queryWithRegion(region), &out); err != nil {
		return nil, err
	}
	return &out.Volume, nil
}

func findAttachmentByInstance(attachments []volumeAttachmentPayload, instanceID string) *volumeAttachmentPayload {
	for i := range attachments {
		if strings.EqualFold(attachments[i].ServerID, instanceID) {
			return &attachments[i]
		}
	}
	return nil
}

func volumeAttachmentStateID(volumeID, instanceID string) string {
	return volumeID + "/" + instanceID
}

func parseVolumeAttachmentImportID(v string) (string, string) {
	parts := strings.SplitN(v, "/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	if strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", ""
	}
	return parts[0], parts[1]
}
