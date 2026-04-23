package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &imagesDataSource{}
var _ datasource.DataSourceWithConfigure = &imagesDataSource{}

type imagesDataSource struct {
	client *apiClient
}

type imagesDataSourceModel struct {
	Region      types.String `tfsdk:"region"`
	ImageGroups types.List   `tfsdk:"image_groups"`
}

func NewImagesDataSource() datasource.DataSource {
	return &imagesDataSource{}
}

func (d *imagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

func (d *imagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all OS images available in a region, grouped by Linux distribution.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Region to query for available images. Defaults to the provider-level region.",
			},
			"image_groups": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "OS images grouped by Linux distribution.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"distro": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Linux distribution name (e.g. `ubuntu`, `debian`, `centos`).",
						},
						"versions": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Available versions of this distribution.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Unique image ID to use in `image_id` module inputs.",
									},
									"version": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Human-readable version string (e.g. `22.04`).",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *imagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*apiClient)
}

func (d *imagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state imagesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(d.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	var out imagesEnvelope
	if err := d.client.get(ctx, "/images", queryWithRegion(region), &out); err != nil {
		resp.Diagnostics.AddError("Read images failed", describeAPIError(err))
		return
	}

	versionTypes := map[string]attr.Type{
		"id":      types.StringType,
		"version": types.StringType,
	}
	groupTypes := map[string]attr.Type{
		"distro":   types.StringType,
		"versions": types.ListType{ElemType: types.ObjectType{AttrTypes: versionTypes}},
	}

	groups := make([]types.Object, 0, len(out.ImageGroups))
	for _, g := range out.ImageGroups {
		versions := make([]types.Object, 0, len(g.Versions))
		for _, v := range g.Versions {
			obj, diags := types.ObjectValue(versionTypes, map[string]attr.Value{
				"id":      types.StringValue(v.ID),
				"version": types.StringValue(v.Version),
			})
			resp.Diagnostics.Append(diags...)
			versions = append(versions, obj)
		}
		versionList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: versionTypes}, versions)
		resp.Diagnostics.Append(diags...)

		groupObj, diags := types.ObjectValue(groupTypes, map[string]attr.Value{
			"distro":   types.StringValue(g.Distro),
			"versions": versionList,
		})
		resp.Diagnostics.Append(diags...)
		groups = append(groups, groupObj)
	}

	state.Region = types.StringValue(region)
	state.ImageGroups, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: groupTypes}, groups)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
