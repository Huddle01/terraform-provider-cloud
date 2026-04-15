package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &instanceDataSource{}
var _ datasource.DataSourceWithConfigure = &instanceDataSource{}

type instanceDataSource struct {
	client *apiClient
}

type instanceDataSourceModel struct {
	ID          types.String  `tfsdk:"id"`
	Region      types.String  `tfsdk:"region"`
	Name        types.String  `tfsdk:"name"`
	Status      types.String  `tfsdk:"status"`
	VCPUs       types.Float64 `tfsdk:"vcpus"`
	RAM         types.Float64 `tfsdk:"ram"`
	CreatedAt   types.String  `tfsdk:"created_at"`
	PrivateIPv4 types.String  `tfsdk:"private_ipv4"`
	PublicIPv4  types.String  `tfsdk:"public_ipv4"`
}

func NewInstanceDataSource() datasource.DataSource {
	return &instanceDataSource{}
}

func (d *instanceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (d *instanceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"name":         schema.StringAttribute{Computed: true},
			"status":       schema.StringAttribute{Computed: true},
			"vcpus":        schema.Float64Attribute{Computed: true},
			"ram":          schema.Float64Attribute{Computed: true},
			"created_at":   schema.StringAttribute{Computed: true},
			"private_ipv4": schema.StringAttribute{Computed: true},
			"public_ipv4":  schema.StringAttribute{Computed: true},
		},
	}
}

func (d *instanceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*apiClient)
}

func (d *instanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state instanceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(d.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	var out instanceResponseEnvelope
	if err := d.client.get(ctx, "/instances/"+state.ID.ValueString(), queryWithRegion(region), &out); err != nil {
		resp.Diagnostics.AddError("Read instance failed", describeAPIError(err))
		return
	}

	privateIP, publicIP := extractIPv4(out.Instance.Networks)
	state.Region = types.StringValue(region)
	state.Name = types.StringValue(out.Instance.Name)
	state.Status = types.StringValue(out.Instance.Status)
	state.VCPUs = types.Float64Value(out.Instance.VCPUs)
	state.RAM = types.Float64Value(out.Instance.RAM)
	state.CreatedAt = types.StringValue(out.Instance.CreatedAt)
	state.PrivateIPv4 = types.StringValue(privateIP)
	state.PublicIPv4 = types.StringValue(publicIP)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
