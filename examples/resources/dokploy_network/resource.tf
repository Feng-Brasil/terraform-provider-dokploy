# Local bridge network with custom IPAM
resource "dokploy_network" "bridge_network" {
  name        = "apps-bridge"
  driver      = "bridge"
  internal    = false
  attachable  = true
  enable_ipv4 = true
  enable_ipv6 = false
  mtu         = 1450

  ipam {
    driver = "default"

    config {
      subnet   = "172.30.0.0/16"
      gateway  = "172.30.0.1"
      ip_range = "172.30.10.0/24"
    }
  }
}

# Overlay network in a remote Dokploy server
resource "dokploy_network" "overlay_network" {
  name        = "apps-overlay"
  driver      = "overlay"
  internal    = false
  attachable  = true
  enable_ipv4 = true
  enable_ipv6 = false
  server_id   = var.server_id

  ipam {
    config {
      subnet = "10.55.0.0/16"
    }
  }
}
