package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPasswordResource(t *testing.T) {
	passwordName := acctest.RandomWithPrefix("provider-test")
	vaultId := os.Getenv("PASSWORK_VAULT_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccPasswordResourceConfig(passwordName, vaultId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("passwork_password.test", "name", passwordName),
					resource.TestCheckResourceAttr("passwork_password.test", "vault_id", vaultId),
					resource.TestCheckResourceAttr("passwork_password.test", "login", "provider-test-user"),
					resource.TestCheckResourceAttr("passwork_password.test", "password", "provider-test-password"),
					resource.TestCheckResourceAttr("passwork_password.test", "url", "https://login.com"),
					resource.TestCheckResourceAttr("passwork_password.test", "description", "provider-test-description"),
					resource.TestCheckResourceAttr("passwork_password.test", "color", "1"),
					resource.TestCheckResourceAttr("passwork_password.test", "tags.#", "3"),
					resource.TestCheckResourceAttrSet("passwork_password.test", "id"),
					resource.TestCheckResourceAttrSet("passwork_password.test", "access"),
					resource.TestCheckResourceAttrSet("passwork_password.test", "access_code"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "passwork_password.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccPasswordResourceChangedConfig(passwordName, vaultId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("passwork_password.test", "name", passwordName),
					resource.TestCheckResourceAttr("passwork_password.test", "vault_id", vaultId),
					resource.TestCheckResourceAttr("passwork_password.test", "login", "provider-test-user-changed"),
					resource.TestCheckResourceAttr("passwork_password.test", "password", "provider-test-password-changed"),
					resource.TestCheckResourceAttr("passwork_password.test", "url", "https://login-changed.com"),
					resource.TestCheckResourceAttr("passwork_password.test", "description", "provider-test-description-changed"),
					resource.TestCheckResourceAttr("passwork_password.test", "color", "2"),
					resource.TestCheckResourceAttr("passwork_password.test", "tags.#", "2"),
					resource.TestCheckResourceAttrSet("passwork_password.test", "id"),
					resource.TestCheckResourceAttrSet("passwork_password.test", "access"),
					resource.TestCheckResourceAttrSet("passwork_password.test", "access_code"),
				),
			},
		},
	})
}

func testAccPasswordResourceConfig(passwordName, vaultId string) string {
	return fmt.Sprintf(`
resource "passwork_password" "test" {
	name        = %[1]q
	vault_id    = %[2]q
	login       = "provider-test-user"
	password    = "provider-test-password"
	url         = "https://login.com"
	description = "provider-test-description"
	color       = 1
	tags        = ["provider", "test", "tag"]
}
`, passwordName, vaultId)
}

func testAccPasswordResourceChangedConfig(passwordName, vaultId string) string {
	return fmt.Sprintf(`
resource "passwork_password" "test" {
	name        = %[1]q
	vault_id    = %[2]q
	login       = "provider-test-user-changed"
	password    = "provider-test-password-changed"
	url         = "https://login-changed.com"
	description = "provider-test-description-changed"
	color       = 2
	tags        = ["provider", "changed"]
}
`, passwordName, vaultId)
}
