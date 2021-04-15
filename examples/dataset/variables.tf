variable "dataset_name" {
  type = string
  description = "Name of TrueNAS dataset, not including pool or parent"
}

variable "dataset_pool" {
  type = string
  description = "Pool name where dataset will be created"
}

variable "dataset_parent" {
  type = string
  default = ""
  description = "Parent dataset"
}