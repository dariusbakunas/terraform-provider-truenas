terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/jdella/truenas"
    }
  }
}

variable TRUENAS_API_KEY {
  type        = string
}

variable TRUENAS_HOST_NAME {
  type        = string
}

provider "truenas" {
  api_key = var.TRUENAS_API_KEY
  base_url = "http://${var.TRUENAS_HOST_NAME}/api/v2.0"
}

data "truenas_dataset" "dataset" {
  dataset_id = "della-pool-rust/backup"
} 

output test {
  value       = data.truenas_dataset.dataset
}
