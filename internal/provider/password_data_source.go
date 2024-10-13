package provider

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/lupa95/passwork-client-go"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
		Description: "Use this data source to get information about a password entry. Passwords entries can either be selected by Id or searched for by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Id of the password entry. Either `id` or `name` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The name of the password entry. If `id` is not supplied, password will be searched by name (best effort). Either `id` or `name` must be set.",
			},
			"password": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The password value of the password entry.",
			},
			"vault_id": schema.StringAttribute{
				Optional:    true,
				Description: "The Id of the vault, which the password entry should be searched in. Only applicable if `name` is supplied and `id` is not supplied.",
			},
			"login": schema.StringAttribute{
				Computed:    true,
				Description: "The Login of the password entry.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the password entry.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the password entry.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The list of tags, which are assigned to the password entry.",
			},
			"access": schema.StringAttribute{
				Computed:    true,
				Description: "The type of access of the password entry.",
			},
			"access_code": schema.Int32Attribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The access code of the password entry.",
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
	var searchResponse passwork.PasswordSearchResponse
	var searchRequest passwork.PasswordSearchRequest
	var err error

	// If id is missing, search by name
	if !plan.Id.IsNull() {
		getResponse, err = d.client.GetPassword(plan.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(ParsePasswordResponseError(err))
			return
		}
	} else if plan.Id.IsNull() && !plan.Name.IsNull() {
		searchRequest.Query = plan.Name.ValueString()
		if !plan.VaultId.IsUnknown() {
			searchRequest.VaultId = plan.VaultId.ValueString()
		}
		searchResponse, err = d.client.SearchPassword(searchRequest)
		if err != nil {
			resp.Diagnostics.AddError(ParsePasswordResponseError(err))
			return
		}
		getResponse, err = d.client.GetPassword(searchResponse.Data[0].Id)
		if err != nil {
			resp.Diagnostics.AddError(ParsePasswordResponseError(err))
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Pasword search error.",
			"Please either provide id or name argument.",
		)
		return
	}

	// Decode base64 password
	decryptedPassword, err := base64.StdEncoding.DecodeString(getResponse.Data.CryptedPassword)
	if err != nil {
		resp.Diagnostics.AddError(
			"Pasword search error.",
			"Could not decode password "+err.Error(),
		)
		return
	}

	// Update State
	plan.Password = types.StringValue(string(decryptedPassword))
	plan.Id = types.StringValue(getResponse.Data.Id)
	plan.VaultId = types.StringValue(getResponse.Data.VaultId)
	plan.Name = types.StringValue(getResponse.Data.Name)
	plan.Login = types.StringValue(getResponse.Data.Login)
	plan.Url = types.StringValue(getResponse.Data.Url)
	plan.Description = types.StringValue(getResponse.Data.Description)
	plan.Access = types.StringValue(getResponse.Data.Access)
	plan.AccessCode = types.Int32Value(int32(getResponse.Data.AccessCode))
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
