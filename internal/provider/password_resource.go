// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/base64"
	"fmt"

	"example.com/passwork-client-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
		MarkdownDescription: "Password resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Password entry",
				Required:            true,
			},
			"vault_id": schema.StringAttribute{
				MarkdownDescription: "Name of the Password entry",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Password entry",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access": schema.StringAttribute{
				MarkdownDescription: "Access of the Password entry",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_code": schema.Int64Attribute{
				MarkdownDescription: "Access of the Password entry",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"login": schema.StringAttribute{
				MarkdownDescription: "Login of the Password entry",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password value of the Password entry",
				Optional:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Url value of the Password entry",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description value of the Password entry",
				Optional:            true,
			},
			"color": schema.Int64Attribute{
				MarkdownDescription: "Color number of the Password entry",
				Optional:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags of the Password entry",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"folder_id": schema.StringAttribute{
				MarkdownDescription: "Folder ID of the of the Password entry",
				Optional:            true,
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
	request = ModelToRequest(plan)

	// Send request
	response, err = r.client.AddPassword(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Password",
			"Could not update password, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "DEBUG MESSAGE CREATE")
	tflog.Debug(ctx, response.Data.Id)
	tflog.Debug(ctx, response.Data.Name)

	// Convert response to state
	newState, err = ResponseToModel(response)
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
	newState, err = ResponseToModel(response)
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
	tflog.Debug(ctx, "DEBUG MESSAGE UPDATE 0")
	tflog.Debug(ctx, plan.Name.String())

	// Create request from state
	request = ModelToRequest(plan)
	tflog.Debug(ctx, "DEBUG MESSAGE UPDATE 1")
	tflog.Debug(ctx, request.Name)

	// Send request
	response, err = r.client.EditPassword(plan.Id.ValueString(), request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Password",
			"Could not update password, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "DEBUG MESSAGE UPDATE")
	tflog.Debug(ctx, response.Data.Id)
	tflog.Debug(ctx, response.Data.Name)

	// Convert response to state
	newState, err = ResponseToModel(response)
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

func ModelToRequest(model PasswordResourceModel) passwork.PasswordRequest {
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

func ResponseToModel(response passwork.PasswordResponse) (PasswordResourceModel, error) {
	var model PasswordResourceModel

	// Decode base64 password
	decryptedPassword, err := base64.StdEncoding.DecodeString(response.Data.CryptedPassword)
	if err != nil {
		return model, err
	}

	model.VaultId = types.StringValue(response.Data.VaultId)
	model.FolderId = types.StringValue(response.Data.FolderId)
	model.Id = types.StringValue(response.Data.Id)
	model.Name = types.StringValue(response.Data.Name)
	model.Login = types.StringValue(response.Data.Login)
	model.Password = types.StringValue(string(decryptedPassword))
	model.Description = types.StringValue(response.Data.Description)
	model.Url = types.StringValue(response.Data.Url)
	model.Color = types.Int64Value(int64(response.Data.Color))
	for _, tag := range response.Data.Tags {
		model.Tags = append(model.Tags, types.StringValue(tag))
	}
	model.Access = types.StringValue(response.Data.Access)
	model.AccessCode = types.Int64Value(int64(response.Data.AccessCode))

	return model, nil
}