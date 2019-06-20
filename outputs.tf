output "ui" {
  sensitive = false
  value     = "http://${google_compute_instance.yugabyte_node.0.network_interface.0.access_config.0.nat_ip}:7000"
}
output "ssh_key" {
  sensitive = false
  value     = "${var.ssh_private_key}"
}