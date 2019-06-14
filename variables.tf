variable "use_public_ip_for_ssh" {
  description = "Flag to control use of public or private ips for ssh."
  default = "true"
  type = "string"
}
variable "replication_factor" {
  description = "The replication factor for the universe."
  default     = 3
  type        = "string"
}
variable "node_count" {
  description = "The number of nodes to create YugaByte Db Cluter"
  default     = 3
  type        = "string"  
}
variable "vpc_network" {
  description = "VPC network to deploy YugaByte DB"
  default     = "default"
  type        = "string"
}
variable "vpc_firewall" {
  description = "Firewall used by the YugaByte Node"
  default     = "default"
  type        = "string"
}
variable "ssh_key_path" {
  description = "The public key to use when connecting to the instances."
  type        = "string"
}
variable "ssh_pub_key" {
  description = "SSH public key to be used by nodes"
  type        = "string"
}
variable "ssh_user" {
  description = "User name to ssh YugaByte Node to configure cluster"
  type        = "string"
}
variable "node_type" {
  description = "Type of Node to be used for YugaByte DB node "
  default     = "n1-standard-4"
  type        = "string"
}
variable "yb_edition" {
  description = "The edition of YugaByteDB to install"
  default     = "ce"
  type        = "string"
}

variable "yb_download_url" {
  description = "The download location of the YugaByteDB edition"
  default     = "https://downloads.yugabyte.com"
  type        = "string"
}

variable "yb_version" {
  description = "The version number of YugaByteDB to install"
  default     = "1.2.8.0"
  type        = "string"
}

variable "region_name" {
  description = "Region name for GCP"
  default     = "us-west1"
  type        = "string"
}
variable "disk_size" {
  description = "Disk size for YugaByte DB nodes"
  default     = "50"
  type        = "string"
}
