package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworkDataSource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_network.test", "id"),
					resource.TestCheckResourceAttrSet("data.dokploy_network.test", "name"),
					resource.TestCheckResourceAttrSet("data.dokploy_network.test", "driver"),
					resource.TestCheckResourceAttrSet("data.dokploy_network.test", "internal"),
					resource.TestCheckResourceAttrSet("data.dokploy_network.test", "attachable"),
					resource.TestCheckResourceAttrSet("data.dokploy_network.test", "enable_ipv4"),
					resource.TestCheckResourceAttrSet("data.dokploy_network.test", "enable_ipv6"),
				),
			},
		},
	})
}

func TestAccNetworksDataSource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworksDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_networks.test", "networks.#"),
				),
			},
		},
	})
}

func testAccNetworkDataSourceConfig() string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = %q
  api_key = %q
}

resource "dokploy_network" "test" {
  name        = "tfacc-ds-network"
  driver      = "bridge"
  internal    = false
  attachable  = true
  enable_ipv4 = true
  enable_ipv6 = false
}

data "dokploy_network" "test" {
  id = dokploy_network.test.id
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"))
}

func testAccNetworksDataSourceConfig() string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = %q
  api_key = %q
}

resource "dokploy_network" "test" {
  name        = "tfacc-ds-networks"
  driver      = "bridge"
  internal    = false
  attachable  = true
  enable_ipv4 = true
  enable_ipv6 = false
}

data "dokploy_networks" "test" {
  depends_on = [dokploy_network.test]
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"))
}
