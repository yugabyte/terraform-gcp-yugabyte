
data "google_compute_image" "YugaByte_DB_Image" {
  family  = "centos-6"
  project = "centos-cloud"
}
data "google_compute_zones" "available" {
    region = "${var.region_name}"
}

resource "google_compute_firewall" "YugaByte-Firewall" {
  name = "${var.vpc_firewall}-${var.prefix}${var.cluster_name}-firewall"
  network = "${var.vpc_network}"
  allow {
      protocol = "tcp"
      ports = ["9000","7000","6379","9042","5433","22"]
  }
  target_tags = ["${var.prefix}${var.cluster_name}"]
}
resource "google_compute_firewall" "YugaByte-Intra-Firewall" {
  name = "${var.vpc_firewall}-${var.prefix}${var.cluster_name}-intra-firewall"
  network = "${var.vpc_network}"
  allow {
      protocol = "tcp"
      ports = ["7100", "9100"]
  }
  target_tags = ["${var.prefix}${var.cluster_name}"]
}

resource "google_compute_instance" "yugabyte_node" {
    count = "${var.node_count}"
    name = "${var.prefix}${var.cluster_name}-n${format("%d", count.index + 1)}"
    machine_type = "${var.node_type}"
    zone = "${element(data.google_compute_zones.available.names, count.index)}"
    tags=["${var.prefix}${var.cluster_name}"]
    
    boot_disk{
        initialize_params {
            image = "${data.google_compute_image.YugaByte_DB_Image.self_link}"
            size = "${var.disk_size}"
        }
    }
    metadata = { 
        sshKeys = "${var.ssh_user}:${file(var.ssh_public_key)}"
    }

    network_interface{
        network = "${var.vpc_network}"
        access_config {
            // external ip to instance
        }
    }

    provisioner "file" {
        source = "${path.module}/utilities/scripts/install_software.sh"
        destination = "/home/${var.ssh_user}/install_software.sh"
        connection {
	    host = "${self.network_interface.0.access_config.0.nat_ip}" 
            type = "ssh"
            user = "${var.ssh_user}"
            private_key = "${file(var.ssh_private_key)}"
        }
    }

    provisioner "file" {
        source = "${path.module}/utilities/scripts/create_universe.sh"
        destination ="/home/${var.ssh_user}/create_universe.sh"
        connection {
	    host = "${self.network_interface.0.access_config.0.nat_ip}" 
            type = "ssh"
            user = "${var.ssh_user}"
            private_key = "${file(var.ssh_private_key)}"
        }
    }
    provisioner "file" {
        source = "${path.module}/utilities/scripts/start_master.sh"
        destination ="/home/${var.ssh_user}/start_master.sh"
        connection {
	    host = "${self.network_interface.0.access_config.0.nat_ip}" 
            type = "ssh"
            user = "${var.ssh_user}"
            private_key = "${file(var.ssh_private_key)}"
        }
    }
    provisioner "file" {
        source = "${path.module}/utilities/scripts/start_tserver.sh"
        destination ="/home/${var.ssh_user}/start_tserver.sh"
        connection {
	    host = "${self.network_interface.0.access_config.0.nat_ip}" 
            type = "ssh"
            user = "${var.ssh_user}"
            private_key = "${file(var.ssh_private_key)}"
        }
    }
    provisioner "remote-exec" {
        inline = [
            "chmod +x /home/${var.ssh_user}/install_software.sh",
            "chmod +x /home/${var.ssh_user}/create_universe.sh",
            "chmod +x /home/${var.ssh_user}/start_tserver.sh",
            "chmod +x /home/${var.ssh_user}/start_master.sh",
            "/home/${var.ssh_user}/install_software.sh '${var.yb_version}'"
        ]
        connection {
	    host = "${self.network_interface.0.access_config.0.nat_ip}" 
            type = "ssh"
            user = "${var.ssh_user}"
            private_key = "${file(var.ssh_private_key)}"
        }
    }
}

locals {
    depends_on = ["google_compute_instance.yugabyte_node"]
    ssh_ip_list = "${var.use_public_ip_for_ssh == "true" ? join(" ",google_compute_instance.yugabyte_node.*.network_interface.0.access_config.0.nat_ip) : join(" ",google_compute_instance.yugabyte_node.*.network_interface.0.network_ip)}"
    config_ip_list = "${join(" ",google_compute_instance.yugabyte_node.*.network_interface.0.network_ip)}"
    zone = "${join(" ", google_compute_instance.yugabyte_node.*.zone)}"
}

resource "null_resource" "create_yugabyte_universe" {
  depends_on = ["google_compute_instance.yugabyte_node"]

  provisioner "local-exec" {
      command = "${path.module}/utilities/scripts/create_universe.sh 'GCP' '${var.region_name}' ${var.replication_factor} '${local.config_ip_list}' '${local.ssh_ip_list}' '${local.zone}' '${var.ssh_user}' ${var.ssh_private_key}"
  }
}

