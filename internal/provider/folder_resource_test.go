package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestFolderResource(t *testing.T) {
	folderName := acctest.RandomWithPrefix("provider-test")
	folderNameRenamed := fmt.Sprintf("%s-renamed", folderName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccFoldertResourceConfig(folderName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("passwork_folder.test", "name", folderName),
					resource.TestCheckResourceAttrSet("passwork_folder.test", "id"),
					resource.TestCheckResourceAttrPair("passwork_folder.test_nested", "parent_id", "passwork_folder.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "passwork_folder.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccFoldertResourceConfig(folderNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("passwork_folder.test", "name", folderNameRenamed),
					resource.TestCheckResourceAttrSet("passwork_folder.test", "id"),
				),
			},
		},
	})
}

func testAccFoldertResourceConfig(folderName string) string {
	return fmt.Sprintf(`

resource "passwork_vault" "test" {
	name       = %[1]q
	is_private = true
}

resource "passwork_folder" "test" {
	name     = %[1]q
	vault_id = passwork_vault.test.id
}

resource "passwork_folder" "test_nested" {
	name      = "provider-test-folder-nested"
	vault_id  = passwork_folder.test.vault_id
	parent_id = passwork_folder.test.id
}
`, folderName)
}
