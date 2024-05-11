package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lupa95/passwork-client-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &FolderResource{}
var _ resource.ResourceWithImportState = &FolderResource{}

func NewFolderResource() resource.Resource {
	return &FolderResource{}
}

// ExampleResource defines the resource implementation.
type FolderResource struct {
	client *passwork.Client
}

func (r *FolderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folder"
}

func (r *FolderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this resource to create a folder. Folders can be used to organize password entries. Folders need to be create inside a vault.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the folder entry.",
				Required:    true,
			},
			"vault_id": schema.StringAttribute{
				Description: "The Id of the vault, which the folder should be created in.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The Id of the folder.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parent_id": schema.StringAttribute{
				Description: "The Id of the parent folder of the folder. Omit if this should be a top level folder.",
				Optional:    true,
			},
		},
	}
}

func (r *FolderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FolderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan     FolderResourceModel
		newState FolderResourceModel
		request  passwork.FolderRequest
		response passwork.FolderResponse
		err      error
	)

	// Retrieve values from plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request
	request.Name = plan.Name.ValueString()
	if !plan.VaultId.IsNull() {
		request.VaultId = plan.VaultId.ValueString()
	}
	if !plan.ParentId.IsNull() {
		request.ParentId = plan.ParentId.ValueString()
	}

	// Send request
	response, err = r.client.AddFolder(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating folder",
			"Could not create folder, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert response to model
	newState = FolderResponseToModel(response)

	// Set refreshed state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *FolderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    FolderResourceModel
		newState FolderResourceModel
		response passwork.FolderResponse
		err      error
	)

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err = r.client.GetFolder(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting folder",
			"Could not delete folder, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert response to model
	newState = FolderResponseToModel(response)

	// Set refreshed state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *FolderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan     FolderResourceModel
		newState FolderResourceModel
		request  passwork.FolderRequest
		response passwork.FolderResponse
		err      error
	)

	// Retrieve values from plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request
	request.Name = plan.Name.ValueString()

	// Send request
	response, err = r.client.EditFolder(plan.Id.ValueString(), request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating folder",
			"Could not update folder, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert response to model
	newState = FolderResponseToModel(response)

	// Set refreshed state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *FolderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var (
		plan FolderResourceModel
		err  error
	)

	// Retrieve values from plan
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Send request
	_, err = r.client.DeleteFolder(plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating folder",
			"Could not update folder, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *FolderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func FolderResponseToModel(response passwork.FolderResponse) FolderResourceModel {
	folder := FolderResourceModel{
		Name:    types.StringValue(response.Data.Name),
		Id:      types.StringValue(response.Data.Id),
		VaultId: types.StringValue(response.Data.VaultId),
	}

	// Client SDK returns empty string if there is no parent ID. Set to null in Terraform
	if response.Data.ParentId != "" {
		folder.ParentId = types.StringValue(response.Data.ParentId)
	}

	return folder
}
