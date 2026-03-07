variable "instance_shape" { default = "VM.Standard.A1.Flex" } 
variable "compartment_id" {
  description = "The OCID of your OCI compartment/tenancy"
  type        = string
}

variable "region" {
  description = "The OCI region where resources will be created"
  type        = string
  default     = "eu-frankfurt-1" 
}

variable "ssh_public_key" {
  description = "The public SSH key to access instances"
  type        = string
}