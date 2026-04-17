package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &networksDataSource{}
var _ datasource.DataSourceWithConfigure = &networksDataSource{}

type networksDataSource struct {
	client *apiClient
}

type networksDataSourceModel struct {
	Region   types.String `tfsdk:"region"`
	Owned    types.Bool   `tfsdk:"owned"`
	Networks types.List   `tfsdk:"networks"`
}

func NewNetworksDataSource() datasource.DataSource {
	return &networksDataSource{}
}

func (d *networksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networks"
}

func (d *networksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists networks accessible to the workspace in a region.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Region to query. Defaults to the provider-level region.",
			},
			"owned": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "When `true`, only returns networks owned by the workspace. When `false` (default), shared networks are included.",
			},
			"networks": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of accessible networks.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique identifier of the network.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the network.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Current network status (`ACTIVE`, `DOWN`, etc.).",
						},
						"subnets": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "List of subnet IDs associated with the network.",
						},
						"admin_state_up": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the network is administratively up.",
						},
					},
				},
			},
		},
	}
}

func (d *networksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*apiClient)
}

func (d *networksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state networksDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(d.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}
	q := queryWithRegion(region)
	if boolOrDefault(state.Owned, true) {
		q.Set("owned", "true")
	}

	var out networkListEnvelope
	if err := d.client.get(ctx, "/networks", q, &out); err != nil {
		resp.Diagnostics.AddError("Read networks failed", describeAPIError(err))
		return
	}

	attrTypes := map[string]attr.Type{
		"id":             types.StringType,
		"name":           types.StringType,
		"status":         types.StringType,
		"subnets":        types.ListType{ElemType: types.StringType},
		"admin_state_up": types.BoolType,
	}
	items := make([]types.Object, 0, len(out.Data.Networks))
	for _, n := range out.Data.Networks {
		subnets, _ := types.ListValueFrom(ctx, types.StringType, stringSliceToTerraform(n.Subnets))
		obj, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id":             types.StringValue(n.ID),
			"name":           types.StringValue(n.Name),
			"status":         types.StringValue(n.Status),
			"subnets":        subnets,
			"admin_state_up": types.BoolValue(n.AdminStateUp),
		})
		resp.Diagnostics.Append(diags...)
		items = append(items, obj)
	}

	state.Region = types.StringValue(region)
	state.Owned = types.BoolValue(boolOrDefault(state.Owned, true))
	state.Networks, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: attrTypes}, items)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
