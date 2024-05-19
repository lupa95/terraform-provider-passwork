// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lupa95/passwork-client-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PasswordResource{}
var _ resource.ResourceWithImportState = &PasswordResource{}

func NewPasswordResource() resource.Resource {
	return &PasswordResource{}
}

// ExampleResource defines the resource implementation.
type PasswordResource struct {
	client *passwork.Client
}

func (r *PasswordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password"
}

func (r *PasswordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this resource to create a password entry. Passwords need to be stored inside a vault.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the password entry.",
				Required:    true,
			},
			"vault_id": schema.StringAttribute{
				Description: "The Id of the vault, which the password entry should be stored in.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The Id of the password entry.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access": schema.StringAttribute{
				Description: "The type of access of the password entry.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_code": schema.Int64Attribute{
				Description: "The access code of the password entry.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"login": schema.StringAttribute{
				Description: "The Login of the password entry.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password value of the password entry.",
				Optional:    true,
			},
			"url": schema.StringAttribute{
				Description: "The URL of the password entry.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the password entry.",
				Optional:    true,
			},
			"color": schema.Int64Attribute{
				Description: "The color code of the password entry.",
				Optional:    true,
			},
			"tags": schema.ListAttribute{
				Description: "The list of tags, which are assigned to the password entry.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"folder_id": schema.StringAttribute{
				Description: "The Id of the folder, which the password entry should be stored in.",
				Optional:    true,
			},
		},
	}
}

func (r *PasswordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*passwork.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *passwork.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *PasswordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan     PasswordResourceModel
		newState PasswordResourceModel
		request  passwork.PasswordRequest
		response passwork.PasswordResponse
		err      error
	)

	// Retrieve values from plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create request from model
	request = PasswordModelToRequest(plan)

	// Send request
	response, err = r.client.AddPassword(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Password",
			"Could not update password, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert response to state
	newState, err = PasswordResponseToModel(response)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting Password response into state",
			"Could not update state with API response, unexpected error: "+err.Error(),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *PasswordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    PasswordResourceModel
		newState PasswordResourceModel
		response passwork.PasswordResponse
		err      error
	)

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed password value from Passwork
	response, err = r.client.GetPassword(state.Id.ValueString())

	// Check for errors
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Passwork Password",
			"Could not read Passwork Password ID "+state.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	// Remove resource from state if error is returned from Passwork
	if response.Status != "success" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert response to state
	newState, err = PasswordResponseToModel(response)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting Password response into state",
			"Could not update state with API response, unexpected error: "+err.Error(),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *PasswordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan     PasswordResourceModel
		newState PasswordResourceModel
		request  passwork.PasswordRequest
		response passwork.PasswordResponse
		err      error
	)

	// Retrieve values from plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create request from state
	request = PasswordModelToRequest(plan)

	// Send request
	response, err = r.client.EditPassword(plan.Id.ValueString(), request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Password",
			"Could not update password, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert response to state
	newState, err = PasswordResponseToModel(response)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting Password response into state",
			"Could not update state with API response, unexpected error: "+err.Error(),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *PasswordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var (
		plan PasswordResourceModel
		err  error
	)

	// Retrieve values from plan
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Send delete request
	_, err = r.client.DeletePassword(plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Password",
			"Could not delete password, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *PasswordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func PasswordModelToRequest(model PasswordResourceModel) passwork.PasswordRequest {
	// Encode base64 password
	cryptedPassword := base64.StdEncoding.EncodeToString([]byte(model.Password.ValueString()))

	// Generate API request body from model
	var request = passwork.PasswordRequest{
		Name:            model.Name.ValueString(),
		Login:           model.Login.ValueString(),
		CryptedPassword: cryptedPassword,
		Description:     model.Description.ValueString(),
		Url:             model.Url.ValueString(),
		Color:           int(model.Color.ValueInt64()),
		VaultId:         model.VaultId.ValueString(),
		FolderId:        model.FolderId.ValueString(),
	}

	for _, tag := range model.Tags {
		request.Tags = append(request.Tags, tag.ValueString())
	}

	return request
}

func PasswordResponseToModel(response passwork.PasswordResponse) (PasswordResourceModel, error) {
	var model PasswordResourceModel

	if response.Data.CryptedPassword == "" {
		model.Password = types.StringNull()
	} else {
		decryptedPassword, err := base64.StdEncoding.DecodeString(response.Data.CryptedPassword)
		if err != nil {
			return model, err
		}
		model.Password = types.StringValue(string(decryptedPassword))
	}

	model.VaultId = types.StringValue(response.Data.VaultId)
	if response.Data.FolderId != "" {
		model.FolderId = types.StringValue(response.Data.FolderId)
	}
	model.Id = types.StringValue(response.Data.Id)
	model.Name = types.StringValue(response.Data.Name)

	if response.Data.Login != "" {
		model.Login = types.StringValue(response.Data.Login)
	} else {
		model.Login = types.StringNull()
	}

	if response.Data.Description != "" {
		model.Description = types.StringValue(response.Data.Description)
	} else {
		model.Description = types.StringNull()
	}

	if response.Data.Url != "" {
		model.Url = types.StringValue(response.Data.Url)
	} else {
		model.Url = types.StringNull()
	}

	// No color chosen: API returns 0
	if response.Data.Color != 0 {
		model.Color = types.Int64Value(int64(response.Data.Color))
	}

	for _, tag := range response.Data.Tags {
		model.Tags = append(model.Tags, types.StringValue(tag))
	}

	model.Access = types.StringValue(response.Data.Access)
	model.AccessCode = types.Int64Value(int64(response.Data.AccessCode))

	return model, nil
}
