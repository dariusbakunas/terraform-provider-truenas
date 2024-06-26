terraform {
  required_providers {
    # TrueNas API
    truenas = {
      #source = "jdella/truenas"  # <-- online version?
      source = "jdella/providers/truenas"  # <-- local installed version
      version = "0.77.1"
    }
  }
  backend "local" {
    path = "tests.tfstate"
  }
}

provider "truenas" {
  api_key = "<<<API_KEY>>>"
  base_url = "http://127.0.0.1:<<<API_PORT>>>/api/v2.0"
}
