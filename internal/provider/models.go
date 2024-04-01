package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type PasswordResourceModel struct {
	VaultId     types.String   `tfsdk:"vault_id"`
	FolderId    types.String   `tfsdk:"folder_id"`
	Id          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Login       types.String   `tfsdk:"login"`
	Password    types.String   `tfsdk:"password"`
	Description types.String   `tfsdk:"description"`
	Url         types.String   `tfsdk:"url"`
	Color       types.Int64    `tfsdk:"color"`
	Tags        []types.String `tfsdk:"tags"`
	Access      types.String   `tfsdk:"access"`
	AccessCode  types.Int64    `tfsdk:"access_code"`
}

type passwordDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Id          types.String `tfsdk:"id"`
	VaultId     types.String `tfsdk:"vault_id"`
	Password    types.String `tfsdk:"password"`
	Description types.String `tfsdk:"description"`
	Login       types.String `tfsdk:"login"`
	Url         types.String `tfsdk:"url"`
	Tags        types.List   `tfsdk:"tags"`
	Access      types.String `tfsdk:"access"`
	AccessCode  types.Int64  `tfsdk:"access_code"`
}

type VaultResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	MasterPassword types.String `tfsdk:"master_password"`
	Access         types.String `tfsdk:"access"`
	Scope          types.String `tfsdk:"scope"`
	IsPrivate      types.Bool   `tfsdk:"is_private"`
}

type FolderResourceModel struct {
	Name     types.String `tfsdk:"name"`
	VaultId  types.String `tfsdk:"vault_id"`
	Id       types.String `tfsdk:"id"`
	ParentId types.String `tfsdk:"parent_id"`
	Access   types.String `tfsdk:"access"`
}
