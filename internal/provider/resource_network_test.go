package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworkResource(t *testing.T) {
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
				Config: testAccNetworkResourceConfig(
					"tfacc-network",
					false,
					1400,
					"172.31.0.0/16",
					"",
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_network.test", "name", "tfacc-network"),
					resource.TestCheckResourceAttr("dokploy_network.test", "driver", "bridge"),
					resource.TestCheckResourceAttr("dokploy_network.test", "internal", "false"),
					resource.TestCheckResourceAttr("dokploy_network.test", "attachable", "false"),
					resource.TestCheckResourceAttr("dokploy_network.test", "enable_ipv4", "true"),
					resource.TestCheckResourceAttr("dokploy_network.test", "enable_ipv6", "false"),
					resource.TestCheckResourceAttr("dokploy_network.test", "mtu", "1400"),
					resource.TestCheckResourceAttr("dokploy_network.test", "ipam.config.0.subnet", "172.31.0.0/16"),
					resource.TestCheckResourceAttrSet("dokploy_network.test", "id"),
				),
			},
			{
				Config: testAccNetworkResourceConfig(
					"tfacc-network-updated",
					true,
					1450,
					"172.32.0.0/16",
					"172.32.0.1",
					"172.32.10.0/24",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_network.test", "name", "tfacc-network-updated"),
					resource.TestCheckResourceAttr("dokploy_network.test", "attachable", "true"),
					resource.TestCheckResourceAttr("dokploy_network.test", "mtu", "1450"),
					resource.TestCheckResourceAttr("dokploy_network.test", "ipam.config.0.subnet", "172.32.0.0/16"),
					resource.TestCheckResourceAttr("dokploy_network.test", "ipam.config.0.gateway", "172.32.0.1"),
					resource.TestCheckResourceAttr("dokploy_network.test", "ipam.config.0.ip_range", "172.32.10.0/24"),
				),
			},
			{
				ResourceName:      "dokploy_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNetworkResourceConfig(name string, attachable bool, mtu int, subnet, gateway, ipRange string) string {
	ipamOptional := ""
	if gateway != "" {
		ipamOptional += fmt.Sprintf("\n      gateway  = %q", gateway)
	}
	if ipRange != "" {
		ipamOptional += fmt.Sprintf("\n      ip_range = %q", ipRange)
	}

	return fmt.Sprintf(`
provider "dokploy" {
  host    = %q
  api_key = %q
}

resource "dokploy_network" "test" {
  name        = %q
  driver      = "bridge"
  internal    = false
  attachable  = %t
  enable_ipv4 = true
  enable_ipv6 = false
  mtu         = %d

  ipam {
    config {
      subnet = %q%s
    }
  }
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, attachable, mtu, subnet, ipamOptional)
}
