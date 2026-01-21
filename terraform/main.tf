provider "oci" {
  tenancy_ocid     = var.tenancy_ocid
  user_ocid        = var.user_ocid
  fingerprint      = var.fingerprint
  private_key_path = var.private_key_path
  region           = var.region
}

resource "oci_core_vcn" "main_vcn" {
  compartment_id = var.compartment_ocid
  cidr_block     = "10.0.0.0/16"
  display_name   = "main-vcn"
}

resource "oci_core_internet_gateway" "ig" {
  compartment_id = var.compartment_ocid
  vcn_id         = oci_core_vcn.main_vcn.id
}

resource "oci_core_default_route_table" "rt" {
  manage_default_resource_id = oci_core_vcn.main_vcn.default_route_table_id
  route_rules {
    network_entity_id = oci_core_internet_gateway.ig.id
    destination       = "0.0.0.0/0"
  }
}

resource "oci_core_default_security_list" "sl" {
  manage_default_resource_id = oci_core_vcn.main_vcn.default_security_list_id

  # SSH
  ingress_security_rules { protocol = "6"; source = "0.0.0.0/0"; tcp_options { min = 22; max = 22 } }
  # HTTP/HTTPS
  ingress_security_rules { protocol = "6"; source = "0.0.0.0/0"; tcp_options { min = 80; max = 80 } }
  ingress_security_rules { protocol = "6"; source = "0.0.0.0/0"; tcp_options { min = 443; max = 443 } }
  # SMTP
  ingress_security_rules { protocol = "6"; source = "0.0.0.0/0"; tcp_options { min = 25; max = 25 } }
  
  # Pozwól na cały ruch wychodzący
  egress_security_rules { protocol = "all"; destination = "0.0.0.0/0" }
}

resource "oci_core_subnet" "main_subnet" {
  compartment_id = var.compartment_ocid
  vcn_id         = oci_core_vcn.main_vcn.id
  cidr_block     = "10.0.1.0/24"
}

resource "oci_core_instance" "devops_vps" {
  availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
  compartment_id      = var.compartment_ocid
  shape              = var.instance_shape

  shape_config {
    memory_in_gbs = 6 # Always Free OCI ARM to nawet 24GB
    ocpus         = 1
  }

  source_details {
    source_type = "image"
    source_id   = "ocid1.image.oc1.eu-frankfurt-1.xxxx" # SPRAWDŹ OCID Ubuntu 22.04 dla swojego regionu!
  }

  create_vnic_details {
    subnet_id = oci_core_subnet.main_subnet.id
    assign_public_ip = true
  }

  metadata = {
    ssh_authorized_keys = var.ssh_public_key
    user_data           = base64encode(file("userdata.sh"))
  }
}

data "oci_identity_availability_domains" "ads" {
  compartment_id = var.compartment_ocid
}