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
  quota_bytes = 2147483648
  quota_critical = 90
  quota_warning = 70
  ref_quota_bytes = 1073741824
  ref_quota_critical = 90
  ref_quota_warning = 70
  deduplication = "off"
  exec = "on"
  snap_dir = "hidden"
  readonly = "off"
  record_size = "256K"
  case_sensitivity = "mixed"

  inherit_encryption = false
  encrypted = true

  generate_key = false
  encryption_algorithm = "AES-128-CCM"
  encryption_key = "3e10193aa02f4167edc46c9f4b8ba723eed474deede646fded99628de1878d51"
}