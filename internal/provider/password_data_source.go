package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/lupa95/passwork-client-go"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &passwordDataSource{}
	_ datasource.DataSourceWithConfigure = &passwordDataSource{}
)

// NewPasswordDataSource is a helper function to simplify the provider implementation.
func NewPasswordDataSource() datasource.DataSource {
	return &passwordDataSource{}
}

// passwordDataSource is the data source implementation.
type passwordDataSource struct {
	client *passwork.Client
}

// Metadata returns the data source type name.
func (d *passwordDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password"
}

// Schema defines the schema for the data source.
func (d *passwordDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"id": schema.StringAttribute{
				Optional: true,
			},
			"vault_id": schema.StringAttribute{
				Required: true,
			},
			"password": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"login": schema.StringAttribute{
				Computed: true,
			},
			"url": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"access": schema.StringAttribute{
				Computed: true,
			},
			"access_code": schema.Int64Attribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *passwordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from plan
	var plan passwordDataSourceModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Setup passwork data models
	var getResponse passwork.PasswordResponse
	var searchRequest = passwork.PasswordSearchRequest{
		VaultId: plan.VaultId.ValueString(),
	}

	ctx = tflog.SetField(ctx, "state id", plan.Id)
	tflog.Debug(ctx, "Printing ID")

	// If id is missing, search by name
	if !plan.Id.IsNull() {
		tflog.Debug(ctx, "here1")
		getResponse, _ = d.client.GetPassword(plan.Id.ValueString())
	} else if plan.Id.IsNull() && !plan.Name.IsNull() {
		tflog.Debug(ctx, "here2")
		searchRequest.Query = plan.Name.ValueString()
		searchResponse, _ := d.client.SearchPassword(searchRequest)
		getResponse, _ = d.client.GetPassword(searchResponse.Data[0].Id)
	}
	ctx = tflog.SetField(ctx, "getResponse", getResponse)

	// Decode base64 password
	decryptedPassword, err := base64.StdEncoding.DecodeString(getResponse.Data.CryptedPassword)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	ctx = tflog.SetField(ctx, "password", string(decryptedPassword))
	tflog.Debug(ctx, "Printing password")

	// Update State
	plan.Password = types.StringValue(string(decryptedPassword))
	plan.Id = types.StringValue(getResponse.Data.Id)
	plan.VaultId = types.StringValue(getResponse.Data.VaultId)
	plan.Name = types.StringValue(getResponse.Data.Name)
	plan.Login = types.StringValue(getResponse.Data.Login)
	plan.Url = types.StringValue(getResponse.Data.Url)
	plan.Description = types.StringValue(getResponse.Data.Description)
	plan.Access = types.StringValue(getResponse.Data.Access)
	plan.AccessCode = types.Int64Value(int64(getResponse.Data.AccessCode))
	plan.Tags, _ = types.ListValueFrom(ctx, types.StringType, getResponse.Data.Tags)

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *passwordDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*passwork.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *passwork.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}
