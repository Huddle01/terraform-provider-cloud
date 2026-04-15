package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &regionsDataSource{}
var _ datasource.DataSourceWithConfigure = &regionsDataSource{}

type regionsDataSource struct {
	client *apiClient
}

type regionsDataSourceModel struct {
	Regions types.Map `tfsdk:"regions"`
}

func NewRegionsDataSource() datasource.DataSource {
	return &regionsDataSource{}
}

func (d *regionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *regionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"regions": schema.MapAttribute{
				Computed:    true,
				ElementType: types.BoolType,
			},
		},
	}
}

func (d *regionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*apiClient)
}

func (d *regionsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var out regionsEnvelope
	if err := d.client.get(ctx, "/regions", nil, &out); err != nil {
		resp.Diagnostics.AddError("Read regions failed", describeAPIError(err))
		return
	}

	values := map[string]bool{}
	for region, enabled := range out {
		values[region] = enabled
	}

	regionMap, diags := types.MapValueFrom(ctx, types.BoolType, values)
	resp.Diagnostics.Append(diags...)

	state := regionsDataSourceModel{
		Regions: regionMap,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
