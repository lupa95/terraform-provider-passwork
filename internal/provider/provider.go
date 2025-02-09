// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lupa95/passwork-client-go"
)

// Ensure PassworkProvider satisfies various provider interfaces.
var _ provider.Provider = &PassworkProvider{}

// PassworkProvider defines the provider implementation.
type PassworkProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PassworkProviderModel describes the provider data model.
type passworkProviderModel struct {
	Host    types.String `tfsdk:"host"`
	Api_key types.String `tfsdk:"api_key"`
	Timeout types.Int32  `tfsdk:"timeout"`
}

func (p *PassworkProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "passwork"
	resp.Version = p.version
}

func (p *PassworkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Passwork provider provides resources for managing password resources on the password manager Passwork.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The Passwork instance's API URL (i.e. https://my-passwork-instance.com). This can alternatively be sourced from the `PASSWORK_HOST` environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The Passwork API key which should be used for authentication. This can alternatively be sourced from the `PASSWORK_API_KEY` environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"timeout": schema.Int32Attribute{
				Description: "The timeout in seconds used against the Passwork API. Defaults to `30` seconds.",
				Optional:    true,
			},
		},
	}
}

func (p *PassworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config passworkProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Passwork API Host",
			"The provider cannot create the Passwork API client as there is an unknown configuration value for the Passwork API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PASSWORK_HOST environment variable.",
		)
	}

	if config.Api_key.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Passwork API Key",
			"The provider cannot create the Passwork API client as there is an unknown configuration value for the Passwork API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PASSWORK_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("PASSWORK_HOST")
	apiKey := os.Getenv("PASSWORK_API_KEY")
	timeout := 30

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Api_key.IsNull() {
		apiKey = config.Api_key.ValueString()
	}

	if !config.Timeout.IsNull() {
		timeout = int(config.Timeout.ValueInt32())
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Passwork API Host",
			"The provider cannot create the Passwork API client as there is a missing or empty value for the Passwork API host. "+
				"Set the host value in the configuration or use the PASSWORK_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Passwork API Key",
			"The provider cannot create the Passwork API client as there is a missing or empty value for the Passwork API key. "+
				"Set the api_key value in the configuration or use the PASSWORK_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Passwork client using the configuration values
	timeout_duration := time.Duration(timeout) * time.Second
	url := host + "/api/v4"
	client := passwork.NewClient(url, apiKey, timeout_duration)
	err := client.Login()
	if err != nil {
		resp.Diagnostics.AddError(
			"Passwork API Login failed",
			"Client was unable to login to Passwork. Please check if the api_key and host value on the configuration are correct."+
				"If you use environment variables, please check if PASSWORK_API_KEY and PASSWORK_HOST are correct.",
		)
	}

	// Make the Passwork client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PassworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFolderResource,
		NewPasswordResource,
		NewVaultResource,
	}
}

func (p *PassworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPasswordDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PassworkProvider{
			version: version,
		}
	}
}
