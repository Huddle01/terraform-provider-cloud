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

var _ resource.Resource = &keyPairResource{}
var _ resource.ResourceWithConfigure = &keyPairResource{}
var _ resource.ResourceWithImportState = &keyPairResource{}

type keyPairResource struct {
	client *apiClient
}

type keyPairResourceModel struct {
	ID          types.String `tfsdk:"id"`
	APIID       types.String `tfsdk:"api_id"`
	Name        types.String `tfsdk:"name"`
	PublicKey   types.String `tfsdk:"public_key"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewKeyPairResource() resource.Resource {
	return &keyPairResource{}
}

func (r *keyPairResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keypair"
}

func (r *keyPairResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"api_id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fingerprint": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *keyPairResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*apiClient)
}

func (r *keyPairResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan keyPairResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]any{
		"name":       plan.Name.ValueString(),
		"public_key": plan.PublicKey.ValueString(),
	}

	var out createKeyPairEnvelope
	if err := r.client.post(ctx, "/keypairs", nil, body, generateIdempotencyKey(), &out); err != nil {
		resp.Diagnostics.AddError("Create key pair failed", describeAPIError(err))
		return
	}

	plan.ID = types.StringValue(out.Name)
	plan.APIID = types.StringValue(out.ID)
	plan.Fingerprint = types.StringValue(out.Fingerprint)
	plan.CreatedAt = types.StringValue(out.CreatedAt)
	plan.UpdatedAt = types.StringNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *keyPairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state keyPairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var out keyPairDetailEnvelope
	err := r.client.get(ctx, "/keypairs/"+url.PathEscape(state.Name.ValueString()), nil, &out)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read key pair failed", describeAPIError(err))
		return
	}

	state.ID = types.StringValue(out.KeyPair.Name)
	state.APIID = types.StringValue(out.KeyPair.ID)
	state.Name = types.StringValue(out.KeyPair.Name)
	state.PublicKey = types.StringValue(out.KeyPair.PublicKey)
	state.Fingerprint = types.StringValue(out.KeyPair.Fingerprint)
	state.CreatedAt = types.StringValue(out.KeyPair.CreatedAt)
	state.UpdatedAt = types.StringValue(out.KeyPair.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *keyPairResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported update", "Key pair updates are not supported by API. Recreate the resource.")
}

func (r *keyPairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state keyPairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.delete(ctx, "/keypairs/"+url.PathEscape(state.Name.ValueString()), nil, generateIdempotencyKey(), nil)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete key pair failed", describeAPIError(err))
	}
}

func (r *keyPairResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
