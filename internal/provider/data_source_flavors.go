package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &flavorsDataSource{}
var _ datasource.DataSourceWithConfigure = &flavorsDataSource{}

type flavorsDataSource struct {
	client *apiClient
}

type flavorsDataSourceModel struct {
	Region  types.String `tfsdk:"region"`
	Flavors types.List   `tfsdk:"flavors"`
}

func NewFlavorsDataSource() datasource.DataSource {
	return &flavorsDataSource{}
}

func (d *flavorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flavors"
}

func (d *flavorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"flavors": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true},
						"name":           schema.StringAttribute{Computed: true},
						"vcpus":          schema.Int64Attribute{Computed: true},
						"ram":            schema.Int64Attribute{Computed: true},
						"disk":           schema.Int64Attribute{Computed: true},
						"price_per_hour": schema.Float64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *flavorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*apiClient)
}

func (d *flavorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state flavorsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(d.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	var out flavorsEnvelope
	if err := d.client.get(ctx, "/flavors", queryWithRegion(region), &out); err != nil {
		resp.Diagnostics.AddError("Read flavors failed", describeAPIError(err))
		return
	}

	attrTypes := map[string]attr.Type{
		"id":             types.StringType,
		"name":           types.StringType,
		"vcpus":          types.Int64Type,
		"ram":            types.Int64Type,
		"disk":           types.Int64Type,
		"price_per_hour": types.Float64Type,
	}
	rows := make([]types.Object, 0, len(out.Data.Flavors))
	for _, f := range out.Data.Flavors {
		obj, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id":             types.StringValue(f.ID),
			"name":           types.StringValue(f.Name),
			"vcpus":          types.Int64Value(f.VCPUs),
			"ram":            types.Int64Value(f.RAM),
			"disk":           types.Int64Value(f.Disk),
			"price_per_hour": types.Float64Value(f.PricePerHour),
		})
		resp.Diagnostics.Append(diags...)
		rows = append(rows, obj)
	}

	state.Region = types.StringValue(region)
	state.Flavors, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: attrTypes}, rows)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
