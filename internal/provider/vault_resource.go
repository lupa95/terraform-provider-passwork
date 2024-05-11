// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lupa95/passwork-client-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VaultResource{}
var _ resource.ResourceWithImportState = &VaultResource{}

func NewVaultResource() resource.Resource {
	return &VaultResource{}
}

// ExampleResource defines the resource implementation.
type VaultResource struct {
	client *passwork.Client
}

func (r *VaultResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault"
}

func (r *VaultResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this resource to create a vault. Vaults are top level containers, that contain password entries.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the vault.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The Id of the vault.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access": schema.StringAttribute{
				Description: "The type of access of the vault.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"scope": schema.StringAttribute{
				Description: "The scope of the vault.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_private": schema.BoolAttribute{
				Description:   "Enable to create a private vault. A private vault is only visiable to the user, who created it.",
				Optional:      true,
				Computed:      true,
				Default:       booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"master_password": schema.StringAttribute{
				Description: "The master password of the vault.",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VaultResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VaultResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan         VaultResourceModel
		newState     VaultResourceModel
		request      passwork.VaultAddRequest
		response_add passwork.VaultOperationResponse
		response_get passwork.VaultResponse
		err          error
	)

	// Retrieve values from plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request
	request.Name = plan.Name.ValueString()
	request.IsPrivate = plan.IsPrivate.ValueBool()

	// Not returned by API, set and forget after creating. Not tracked in state
	request.Salt = randomString(12)
	request.PasswordHash = base64.StdEncoding.EncodeToString([]byte(randomString(12)))

	if plan.MasterPassword.IsUnknown() {
		mp := randomString(12)
		request.MpCrypted = base64.StdEncoding.EncodeToString([]byte(mp))
		newState.MasterPassword = types.StringValue(mp)
	} else {
		request.MpCrypted = base64.StdEncoding.EncodeToString([]byte(plan.MasterPassword.ValueString()))
		newState.MasterPassword = plan.MasterPassword
	}

	// Send create request
	response_add, err = r.client.AddVault(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Vault",
			"Could not create Vault, unexpected error: "+err.Error(),
		)
		return
	}

	// Send get request to get all fields
	response_get, err = r.client.GetVault(response_add.Data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Vault",
			"Could not create Vault, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert response to state
	newState, err = VaultResponseToModel(response_get)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting Vault API response to state.",
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

func (r *VaultResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    VaultResourceModel
		newState VaultResourceModel
		response passwork.VaultResponse
		err      error
	)

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed Vault value from Passwork
	response, err = r.client.GetVault(state.Id.ValueString())

	// Check for errors
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Passwork Vault",
			"Could not read Passwork Vault ID "+state.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	// Convert response to state
	newState, err = VaultResponseToModel(response)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting Vault API response to state.",
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

func (r *VaultResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan         VaultResourceModel
		newState     VaultResourceModel
		request      passwork.VaultEditRequest
		response     passwork.VaultOperationResponse
		response_get passwork.VaultResponse
		err          error
	)

	// Retrieve values from plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create request from state
	request.Name = plan.Name.ValueString()
	// Send request
	response, err = r.client.EditVault(plan.Id.ValueString(), request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Vault",
			"Could not update Vault, unexpected error: "+err.Error(),
		)
		return
	}

	// Send get request to get all fields
	response_get, err = r.client.GetVault(response.Data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Vault",
			"Could not create Vault, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert response to state
	newState, err = VaultResponseToModel(response_get)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting Vault API response to state.",
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

func (r *VaultResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var (
		plan VaultResourceModel
		err  error
	)

	// Retrieve values from plan
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Send delete request
	_, err = r.client.DeleteVault(plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Vault",
			"Could not delete Vault, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *VaultResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func randomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		random_number, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if random_number.IsInt64() {
			int64Value := random_number.Int64()
			result[i] = chars[int(int64Value)]
		}
	}
	return string(result)
}

func VaultResponseToModel(response passwork.VaultResponse) (VaultResourceModel, error) {
	var model VaultResourceModel

	model.Id = types.StringValue(response.Data.Id)
	model.Name = types.StringValue(response.Data.Name)
	model.IsPrivate = types.BoolValue(!response.Data.Visible)
	model.Access = types.StringValue(response.Data.Access)
	model.Scope = types.StringValue(response.Data.Scope)
	master_password, err := base64.StdEncoding.DecodeString(response.Data.VaultPasswordCrypted)
	if err != nil {
		return model, err
	}
	model.MasterPassword = types.StringValue(string(master_password))

	return model, nil
}
