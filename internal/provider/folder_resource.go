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
	resp.TypeName = req.ProviderTypeName + "_password"
}

func (r *FolderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Folder resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the folder",
				Required:            true,
			},
			"vault_id": schema.StringAttribute{
				MarkdownDescription: "Vault ID of the folder",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the folder",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parent_id": schema.StringAttribute{
				MarkdownDescription: "Parent folder ID.",
				Optional:            true,
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
			"Error reading folder",
			"Could not read folder, unexpected error: "+err.Error(),
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
}

func (r *FolderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *FolderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func FolderResponseToModel(response passwork.FolderResponse) FolderResourceModel {
	return FolderResourceModel{
		Name:     types.StringValue(response.Data.Name),
		Id:       types.StringValue(response.Data.Id),
		VaultId:  types.StringValue(response.Data.VaultId),
		ParentId: types.StringValue(response.Data.ParentId),
	}
}
