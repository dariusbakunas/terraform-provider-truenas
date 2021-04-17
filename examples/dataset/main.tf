terraform {
  required_providers {
    truenas = {
      version = "0.1"
      source  = "dariusbakunas/providers/truenas"
    }
  }
}

provider "truenas" {}

resource "truenas_dataset" "test" {
  pool = var.dataset_pool
  name = var.dataset_name
  parent = var.dataset_parent
  comments = "Terraform created dataset"
  compression = "off"
  sync = "standard"
  atime = "off"
  copies = 2
}