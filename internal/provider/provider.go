package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &huddleProvider{}

type huddleProvider struct {
	version string
}

type providerModel struct {
	APIKey                types.String `tfsdk:"api_key"`
	Region                types.String `tfsdk:"region"`
	BaseURL               types.String `tfsdk:"base_url"`
	RequestTimeoutSeconds types.Int64  `tfsdk:"request_timeout_seconds"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &huddleProvider{version: version}
	}
}

func (p *huddleProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "huddle_cloud"
	resp.Version = p.version
}

func (p *huddleProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Provider for Huddle01 Cloud IaaS APIs.",
		MarkdownDescription: "The Huddle01 Cloud provider manages cloud infrastructure resources — virtual machines, networks, security groups, keypairs, and volumes — via the Huddle01 Cloud API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description:         "Huddle01 Cloud API key. Defaults to HUDDLE_API_KEY.",
				MarkdownDescription: "Huddle01 Cloud API key used to authenticate all requests. Can also be set via the `HUDDLE_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"region": schema.StringAttribute{
				Description:         "Default region for region-scoped operations. Defaults to HUDDLE_REGION.",
				MarkdownDescription: "Default region for all resource operations (e.g. `eu2`). Can also be set via the `HUDDLE_REGION` environment variable. Individual resources can override this.",
				Optional:            true,
			},
			"base_url": schema.StringAttribute{
				Description:         "Base API URL. Defaults to https://cloud.huddleapis.com/api/v1.",
				MarkdownDescription: "Base URL of the Huddle01 Cloud API. Defaults to `https://cloud.huddleapis.com/api/v1`. Can also be set via the `HUDDLE_BASE_URL` environment variable.",
				Optional:            true,
			},
			"request_timeout_seconds": schema.Int64Attribute{
				Description:         "HTTP request timeout in seconds. Defaults to 60.",
				MarkdownDescription: "HTTP request timeout in seconds. Defaults to `60`.",
				Optional:            true,
			},
		},
	}
}

func (p *huddleProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := stringOrEnv(config.APIKey, "HUDDLE_API_KEY")
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing API key",
			"Set provider.api_key or HUDDLE_API_KEY.",
		)
	}

	region := stringOrEnv(config.Region, "HUDDLE_REGION")
	if region == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("region"),
			"Missing default region",
			"Set provider.region or HUDDLE_REGION.",
		)
	}

	baseURL := stringOrDefault(config.BaseURL, "https://cloud.huddleapis.com/api/v1")
	timeoutSec := int64OrDefault(config.RequestTimeoutSeconds, 60)
	if timeoutSec <= 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("request_timeout_seconds"),
			"Invalid timeout",
			"request_timeout_seconds must be greater than zero.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := newAPIClient(apiClientConfig{
		APIKey:        apiKey,
		BaseURL:       baseURL,
		DefaultRegion: region,
		Timeout:       time.Duration(timeoutSec) * time.Second,
	})

	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *huddleProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRegionsDataSource,
		NewFlavorsDataSource,
		NewImagesDataSource,
		NewNetworksDataSource,
		NewInstanceDataSource,
	}
}

func (p *huddleProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNetworkResource,
		NewSecurityGroupResource,
		NewSecurityGroupRuleResource,
		NewKeyPairResource,
		NewVolumeResource,
		NewVolumeAttachmentResource,
		NewInstanceResource,
	}
}

func (p *huddleProvider) ValidateConfig(context.Context, provider.ValidateConfigRequest, *provider.ValidateConfigResponse) {
}

func stringOrEnv(value types.String, key string) string {
	if !value.IsNull() && !value.IsUnknown() {
		return value.ValueString()
	}
	return os.Getenv(key)
}

func stringOrDefault(value types.String, fallback string) string {
	if !value.IsNull() && !value.IsUnknown() {
		return value.ValueString()
	}
	return fallback
}

func int64OrDefault(value types.Int64, fallback int64) int64 {
	if !value.IsNull() && !value.IsUnknown() {
		return value.ValueInt64()
	}
	return fallback
}

func effectiveRegion(client *apiClient, region types.String) (string, error) {
	if !region.IsNull() && !region.IsUnknown() {
		return region.ValueString(), nil
	}
	if client.defaultRegion != "" {
		return client.defaultRegion, nil
	}
	return "", fmt.Errorf("region must be set on resource or provider")
}
