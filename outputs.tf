output "ui" {
  sensitive = false
  value     = "http://${google_compute_instance.yugabyte_node.0.network_interface.0.access_config.0.nat_ip}:7000"
}
output "ssh_user" {
  sensitive = false
  value = "${var.ssh_user}"
}
output "ssh_key" {
  sensitive = false
  value     = "${var.ssh_private_key}"
}

output "JDBC" {
  sensitive =false
  value     = "postgresql://postgres@${google_compute_instance.yugabyte_node.0.network_interface.0.access_config.0.nat_ip}:5433"
}

output "YSQL"{
  sensitive = false
  value     = "psql -U postgres -h ${google_compute_instance.yugabyte_node.0.network_interface.0.access_config.0.nat_ip} -p 5433"
}

output "YCQL"{
  sensitive = false
  value     = "cqlsh ${google_compute_instance.yugabyte_node.0.network_interface.0.access_config.0.nat_ip} 9042"
}

output "YEDIS"{
  sensitive = false
  value     = "redis-cli -h ${google_compute_instance.yugabyte_node.0.network_interface.0.access_config.0.nat_ip} -p 6379"
}
