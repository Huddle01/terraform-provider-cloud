package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &volumeResource{}
var _ resource.ResourceWithConfigure = &volumeResource{}
var _ resource.ResourceWithImportState = &volumeResource{}

type volumeResource struct {
	client *apiClient
}

type volumeResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Size            types.Int64  `tfsdk:"size"`
	VolumeType      types.String `tfsdk:"volume_type"`
	Region          types.String `tfsdk:"region"`
	Status          types.String `tfsdk:"status"`
	Bootable        types.Bool   `tfsdk:"bootable"`
	Attachments     types.List   `tfsdk:"attachments"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	DeleteOnDestroy types.Bool   `tfsdk:"delete_on_destroy"`
}

type volumeAttachmentStateModel struct {
	ServerID types.String `tfsdk:"server_id"`
	Device   types.String `tfsdk:"device"`
}

func NewVolumeResource() resource.Resource {
	return &volumeResource{}
}

func (r *volumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *volumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"size": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"volume_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("standard"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"status":   schema.StringAttribute{Computed: true},
			"bootable": schema.BoolAttribute{Computed: true},
			"attachments": schema.ListNestedAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"server_id": schema.StringAttribute{Computed: true},
						"device":    schema.StringAttribute{Computed: true},
					},
				},
			},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
			"delete_on_destroy": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				MarkdownDescription: "When `true`, the volume is deleted from the cloud when the resource is destroyed. " +
					"When `false` (default), `terraform destroy` removes the resource from state but **leaves the volume intact** in the cloud. " +
					"Set this to `true` only if you want Terraform to permanently delete the volume on destroy.",
			},
		},
	}
}

func (r *volumeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*apiClient)
}

func (r *volumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeResourceModel
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
		"size": plan.Size.ValueInt64(),
	}
	if v := stringOrEmpty(plan.Description); v != "" {
		body["description"] = v
	}
	if v := stringOrEmpty(plan.VolumeType); v != "" {
		body["volume_type"] = v
	}

	var created createVolumeEnvelope
	if err := r.client.post(ctx, "/volumes", queryWithRegion(region), body, generateIdempotencyKey(), &created); err != nil {
		resp.Diagnostics.AddError("Create volume failed", describeAPIError(err))
		return
	}

	if created.ID == "" {
		resp.Diagnostics.AddError("Create volume failed", "API response missing volume id")
		return
	}

	if err := waitForVolumeStatus(ctx, r.client, created.ID, region, 5*time.Minute, "available", "in-use"); err != nil {
		resp.Diagnostics.AddError("Volume provisioning timed out", err.Error())
		return
	}

	next, err := r.readVolume(ctx, created.ID, region)
	if err != nil {
		resp.Diagnostics.AddError("Read volume after create failed", describeAPIError(err))
		return
	}

	deleteOnDestroy := plan.DeleteOnDestroy
	plan = *next
	plan.Region = types.StringValue(region)
	plan.DeleteOnDestroy = deleteOnDestroy

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	next, err := r.readVolume(ctx, state.ID.ValueString(), region)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read volume failed", describeAPIError(err))
		return
	}

	next.Region = types.StringValue(region)
	next.DeleteOnDestroy = state.DeleteOnDestroy
	resp.Diagnostics.Append(resp.State.Set(ctx, next)...)
}

func (r *volumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported update", "Volume updates are not supported by API. Recreate the resource.")
}

func (r *volumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// When delete_on_destroy is false (the default), remove the resource from
	// Terraform state without deleting the actual volume. This prevents accidental
	// data loss and avoids race conditions during terraform destroy when a volume
	// attachment and instance are being destroyed in parallel.
	if !state.DeleteOnDestroy.ValueBool() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	// Wait for the volume to become available before attempting deletion.
	// This is necessary because terraform destroy runs resource deletions in
	// parallel: the volume_attachment detach may have just completed but
	// OpenStack processes the detach asynchronously, so the volume can still
	// report "in-use" for a few seconds afterward.
	if err := waitForVolumeStatus(ctx, r.client, state.ID.ValueString(), region, 5*time.Minute, "available"); err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Waiting for volume to become available failed", err.Error())
		return
	}

	if err := r.client.delete(ctx, "/volumes/"+url.PathEscape(state.ID.ValueString()), queryWithRegion(region), generateIdempotencyKey(), nil); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete volume failed", describeAPIError(err))
		return
	}

	// The delete endpoint returns as soon as the request is accepted; the volume
	// is removed asynchronously. Wait for the resource to disappear so that
	// CheckDestroy assertions (and any follow-up operations) observe a
	// consistent post-destroy state.
	if err := waitForVolumeDeleted(ctx, r.client, state.ID.ValueString(), region, 5*time.Minute); err != nil {
		resp.Diagnostics.AddError("Waiting for volume deletion failed", err.Error())
		return
	}
}

func (r *volumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	if r.client.defaultRegion != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), r.client.defaultRegion)...)
	}
}

func (r *volumeResource) readVolume(ctx context.Context, id, region string) (*volumeResourceModel, error) {
	var out volumeDetailEnvelope
	if err := r.client.get(ctx, "/volumes/"+url.PathEscape(id), queryWithRegion(region), &out); err != nil {
		return nil, err
	}

	attachments, diags := flattenVolumeAttachments(out.Volume.Attachments)
	if diags.HasError() {
		return nil, fmt.Errorf("flatten volume attachments")
	}

	// description and volume_type are Optional (not Computed). Storing the API's
	// default values ("" / "__DEFAULT__") when the user omitted the field would
	// cause a plan diff on the next refresh and trigger RequiresReplace. Normalize
	// both to null so state stays in sync with an unset config attribute.
	description := types.StringNull()
	if out.Volume.Description != "" {
		description = types.StringValue(out.Volume.Description)
	}

	// The API may return "" or "__DEFAULT__" when the volume was created without
	// an explicit type. Normalize both to "standard" to match the schema default.
	volumeType := types.StringValue("standard")
	if out.Volume.VolumeType != "" && out.Volume.VolumeType != "__DEFAULT__" {
		volumeType = types.StringValue(out.Volume.VolumeType)
	}

	return &volumeResourceModel{
		ID:              types.StringValue(out.Volume.ID),
		Name:            types.StringValue(out.Volume.Name),
		Description:     description,
		Size:            types.Int64Value(out.Volume.Size),
		VolumeType:      volumeType,
		Status:          types.StringValue(out.Volume.Status),
		Bootable:        types.BoolValue(out.Volume.Bootable),
		Attachments:     attachments,
		CreatedAt:       types.StringValue(out.Volume.CreatedAt),
		UpdatedAt:       types.StringValue(out.Volume.UpdatedAt),
		DeleteOnDestroy: types.BoolValue(false),
	}, nil
}

func flattenVolumeAttachments(in []volumeAttachmentPayload) (types.List, diag.Diagnostics) {
	attrTypes := map[string]attr.Type{
		"server_id": types.StringType,
		"device":    types.StringType,
	}

	items := make([]types.Object, 0, len(in))
	for _, a := range in {
		obj, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"server_id": types.StringValue(a.ServerID),
			"device":    types.StringValue(a.Device),
		})
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
		}
		items = append(items, obj)
	}

	return types.ListValueFrom(context.Background(), types.ObjectType{AttrTypes: attrTypes}, items)
}

func attachedServerIDs(attachments types.List) []string {
	if attachments.IsNull() || attachments.IsUnknown() {
		return nil
	}

	ids := make([]string, 0, len(attachments.Elements()))
	for _, el := range attachments.Elements() {
		obj, ok := el.(types.Object)
		if !ok {
			continue
		}
		attrVal, ok := obj.Attributes()["server_id"]
		if !ok {
			continue
		}
		s, ok := attrVal.(types.String)
		if !ok || s.IsNull() || s.IsUnknown() {
			continue
		}
		if strings.TrimSpace(s.ValueString()) == "" {
			continue
		}
		ids = append(ids, s.ValueString())
	}
	return ids
}
