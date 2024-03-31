package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestVaultResource(t *testing.T) {
	vaultName := acctest.RandomWithPrefix("provider-test")
	vaultNameRenamed := fmt.Sprintf("%s-renamed", vaultName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccVaultResourceConfig(vaultName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("passwork_vault.test", "name", vaultName),
					resource.TestCheckResourceAttr("passwork_vault.test", "is_private", "true"),
					resource.TestCheckResourceAttrSet("passwork_vault.test", "id"),
					resource.TestCheckResourceAttrSet("passwork_vault.test", "master_password"),
					resource.TestCheckResourceAttrSet("passwork_vault.test", "access"),
					resource.TestCheckResourceAttrSet("passwork_vault.test", "scope"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "passwork_vault.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccVaultResourceConfig(vaultNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("passwork_vault.test", "name", vaultNameRenamed),
					resource.TestCheckResourceAttr("passwork_vault.test", "is_private", "true"),
					resource.TestCheckResourceAttrSet("passwork_vault.test", "id"),
					resource.TestCheckResourceAttrSet("passwork_vault.test", "master_password"),
					resource.TestCheckResourceAttrSet("passwork_vault.test", "access"),
					resource.TestCheckResourceAttrSet("passwork_vault.test", "scope"),
				),
			},
		},
	})
}

func testAccVaultResourceConfig(vaultName string) string {
	return fmt.Sprintf(`
resource "passwork_vault" "test" {
	name       = %[1]q
	is_private = true
}
`, vaultName)
}
