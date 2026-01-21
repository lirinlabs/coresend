variable "tenancy_ocid" {}
variable "user_ocid" {}
variable "fingerprint" {}
variable "private_key_path" {}
variable "region" { default = "eu-frankfurt-1" } 
variable "compartment_ocid" {}
variable "ssh_public_key" {}
variable "instance_shape" { default = "VM.Standard.A1.Flex" } # ARM (Always Free)