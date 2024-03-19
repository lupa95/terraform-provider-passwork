package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPasswordDataSource(t *testing.T) {
	dataSourceName := "data.passwork_password.test"
	passwordResourceName := "passwork_password.test"

	passwordName := acctest.RandomWithPrefix("provider-test")
	vaultId := os.Getenv("PASSWORK_VAULT_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccPasswordDataSourceConfig(passwordName, vaultId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", passwordResourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", passwordResourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vault_id", passwordResourceName, "vault_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "login", passwordResourceName, "login"),
					resource.TestCheckResourceAttrPair(dataSourceName, "password", passwordResourceName, "password"),
					resource.TestCheckResourceAttrPair(dataSourceName, "url", passwordResourceName, "url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", passwordResourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", passwordResourceName, "tags"),
				),
			},
		},
	})
}

func testAccPasswordDataSourceConfig(passwordName, vaultId string) string {
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

data "passwork_password" "test" {
	name     = passwork_password.test.name
	vault_id = passwork_password.test.vault_id
}
`, passwordName, vaultId)
}
