package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
	"github.com/trueform/terraform-provider-trueform/internal/datasources"
	"github.com/trueform/terraform-provider-trueform/internal/resources"
)

// Ensure TrueformProvider satisfies various provider interfaces.
var _ provider.Provider = &TrueformProvider{}

// TrueformProvider defines the provider implementation.
type TrueformProvider struct {
	version string
}

// TrueformProviderModel describes the provider data model.
type TrueformProviderModel struct {
	Host      types.String `tfsdk:"host"`
	APIKey    types.String `tfsdk:"api_key"`
	VerifySSL types.Bool   `tfsdk:"verify_ssl"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TrueformProvider{
			version: version,
		}
	}
}

func (p *TrueformProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "trueform"
	resp.Version = p.version
}

func (p *TrueformProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing TrueNAS Scale 25.04+ resources via the WebSocket JSON-RPC API.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The hostname or IP address of the TrueNAS server. Can also be set via the TRUENAS_HOST environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key for authenticating with TrueNAS. Can also be set via the TRUENAS_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"verify_ssl": schema.BoolAttribute{
				Description: "Whether to verify SSL certificates. Defaults to true. Can also be set via the TRUENAS_VERIFY_SSL environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *TrueformProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Trueform provider")

	var config TrueformProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get configuration values, with environment variable fallbacks
	host := os.Getenv("TRUENAS_HOST")
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	apiKey := os.Getenv("TRUENAS_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	verifySSL := true
	if envVal := os.Getenv("TRUENAS_VERIFY_SSL"); envVal == "false" {
		verifySSL = false
	}
	if !config.VerifySSL.IsNull() {
		verifySSL = config.VerifySSL.ValueBool()
	}

	// Validate required configuration
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing TrueNAS Host",
			"The provider cannot create the TrueNAS API client without a host. "+
				"Set the host value in the configuration or use the TRUENAS_HOST environment variable.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing TrueNAS API Key",
			"The provider cannot create the TrueNAS API client without an API key. "+
				"Set the api_key value in the configuration or use the TRUENAS_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API client
	tflog.Debug(ctx, "Creating TrueNAS API client", map[string]interface{}{
		"host":       host,
		"verify_ssl": verifySSL,
	})

	apiClient := client.NewClient(&client.Config{
		Host:      host,
		APIKey:    apiKey,
		VerifySSL: verifySSL,
	})

	// Test connection
	if err := apiClient.Connect(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Connect to TrueNAS",
			"An unexpected error occurred when creating the TrueNAS API client. "+
				"Error: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Successfully connected to TrueNAS")

	// Make the client available to resources and data sources
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *TrueformProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewPoolResource,
		resources.NewDatasetResource,
		resources.NewSnapshotResource,
		resources.NewShareSMBResource,
		resources.NewShareNFSResource,
		resources.NewUserResource,
		resources.NewVMResource,
		resources.NewVMDeviceResource,
		resources.NewAppResource,
		resources.NewCronjobResource,
		resources.NewISCSIPortalResource,
		resources.NewISCSITargetResource,
		resources.NewISCSIExtentResource,
		resources.NewISCSIInitiatorResource,
		resources.NewISCSITargetExtentResource,
		resources.NewCertificateResource,
		resources.NewStaticRouteResource,
	}
}

func (p *TrueformProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewPoolDataSource,
		datasources.NewDatasetDataSource,
		datasources.NewUserDataSource,
		datasources.NewVMDataSource,
	}
}
