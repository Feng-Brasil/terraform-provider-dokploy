package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserPermissionsResource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	// Note: This test uses the current user (owner) whose permissions cannot be modified.
	// The test verifies the resource creates/reads without error.
	// For non-owner users, permissions would actually be applied.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing - uses current user's member_id
			// Since owner permissions can't change, we just verify the resource works
			{
				Config: testAccUserPermissionsResourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("dokploy_user_permissions.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_user_permissions.test", "member_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_user_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserPermissionsResourceConfig() string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

# Get the current user to obtain member_id
data "dokploy_user" "current" {}

# Note: For owner users, permissions cannot be modified via the API.
# The resource will be created but permissions will remain at their current values.
resource "dokploy_user_permissions" "test" {
  member_id = data.dokploy_user.current.member_id

  # These are left at defaults (false) which matches owner's actual permissions
  can_create_projects       = false
  can_create_services       = false
  can_delete_projects       = false
  can_delete_services       = false
  can_access_to_docker      = false
  can_access_to_api         = false
  can_access_to_ssh_keys    = false
  can_access_to_git_providers = false
  can_access_to_traefik_files = false
  can_create_environments   = false
  can_delete_environments   = false

  accessed_projects     = []
  accessed_environments = []
  accessed_services     = []
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"))
}
