# Oracle Cloud Infrastructure Configuration
terraform {
  required_providers {
    oci = {
      source  = "oracle/oci"
      version = "~> 5.0"
    }
  }
}

provider "oci" {
  region = var.region
}

# Variables
variable "region" {
  description = "OCI region"
  type        = string
  default     = "us-ashburn-1"
}

variable "compartment_id" {
  description = "OCI compartment ID"
  type        = string
}

variable "availability_domain" {
  description = "Availability domain"
  type        = string
  default     = "AD-1"
}

# VCN and Networking
resource "oci_core_vcn" "secure_email_vcn" {
  compartment_id = var.compartment_id
  display_name   = "secure-email-vcn"
  cidr_blocks    = ["10.0.0.0/16"]
  dns_label      = "secureemail"
}

resource "oci_core_internet_gateway" "secure_email_ig" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.secure_email_vcn.id
  display_name   = "secure-email-ig"
}

resource "oci_core_route_table" "secure_email_rt" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.secure_email_vcn.id
  display_name   = "secure-email-rt"

  route_rules {
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
    network_entity_id = oci_core_internet_gateway.secure_email_ig.id
  }
}

resource "oci_core_security_list" "secure_email_sl" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.secure_email_vcn.id
  display_name   = "secure-email-sl"

  # Allow HTTP/HTTPS from anywhere (Cloudflare)
  ingress_security_rules {
    protocol  = "6"
    source    = "0.0.0.0/0"
    stateless = false

    tcp_options {
      min = 80
      max = 80
    }
  }

  ingress_security_rules {
    protocol  = "6"
    source    = "0.0.0.0/0"
    stateless = false

    tcp_options {
      min = 443
      max = 443
    }
  }

  # Allow SSH from anywhere (for management)
  ingress_security_rules {
    protocol  = "6"
    source    = "0.0.0.0/0"
    stateless = false

    tcp_options {
      min = 22
      max = 22
    }
  }

  # Allow internal communication
  ingress_security_rules {
    protocol  = "6"
    source    = "10.0.0.0/16"
    stateless = false
  }

  egress_security_rules {
    protocol    = "all"
    destination = "0.0.0.0/0"
    stateless   = false
  }
}

resource "oci_core_subnet" "secure_email_subnet" {
  compartment_id      = var.compartment_id
  vcn_id              = oci_core_vcn.secure_email_vcn.id
  display_name        = "secure-email-subnet"
  cidr_block          = "10.0.1.0/24"
  availability_domain = var.availability_domain
  route_table_id      = oci_core_route_table.secure_email_rt.id
  security_list_ids   = [oci_core_security_list.secure_email_sl.id]
  dns_label           = "secureemailsubnet"
}

# MySQL Database System
resource "oci_mysql_mysql_db_system" "secure_email_db" {
  compartment_id      = var.compartment_id
  availability_domain = var.availability_domain
  display_name        = "secure-email-db"
  shape_name          = "MySQL.VM.Standard.E3.1.8GB"
  subnet_id           = oci_core_subnet.secure_email_subnet.id
  admin_username      = "admin"
  admin_password      = var.db_admin_password
  hostname_label      = "secureemaildb"
  data_storage_size_in_gb = 50

  mysql_version = "8.0.35"

  configuration_id = oci_mysql_mysql_configuration.secure_email_config.id

  backup_policy {
    is_enabled        = true
    retention_in_days = 7
    window_start_time = "02:00"
  }
}

resource "oci_mysql_mysql_configuration" "secure_email_config" {
  compartment_id = var.compartment_id
  display_name   = "secure-email-config"
  shape_name     = "MySQL.VM.Standard.E3.1.8GB"

  variables {
    innodb_buffer_pool_size = "536870912"
    max_connections         = "200"
    binlog_expire_logs_seconds = "2592000"
  }
}

# Compute Instance for Backend
resource "oci_core_instance" "secure_email_backend" {
  compartment_id      = var.compartment_id
  availability_domain = var.availability_domain
  display_name        = "secure-email-backend"
  shape               = "VM.Standard.E2.1.Micro"

  source_details {
    source_type = "image"
    source_id   = data.oci_core_images.ubuntu_images.images[0].id
  }

  create_vnic_details {
    subnet_id        = oci_core_subnet.secure_email_subnet.id
    display_name     = "secure-email-backend-vnic"
    assign_public_ip = true
    hostname_label   = "secureemailbackend"
  }

  metadata = {
    ssh_authorized_keys = var.ssh_public_key
    user_data = base64encode(templatefile("${path.module}/user_data.sh", {
      db_host = oci_mysql_mysql_db_system.secure_email_db.mysql_endpoint[0].ip_address
      db_port = oci_mysql_mysql_db_system.secure_email_db.mysql_endpoint[0].port
    }))
  }
}

data "oci_core_images" "ubuntu_images" {
  compartment_id   = var.compartment_id
  operating_system = "Canonical Ubuntu"
  sort_by          = "TIMECREATED"
  sort_order       = "DESC"
}

# Outputs
output "backend_public_ip" {
  value = oci_core_instance.secure_email_backend.public_ip
}

output "db_endpoint" {
  value = oci_mysql_mysql_db_system.secure_email_db.mysql_endpoint[0].ip_address
}

output "vcn_id" {
  value = oci_core_vcn.secure_email_vcn.id
}
