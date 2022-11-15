resource "truenas_vm" "vm" {
  name = "TestVM"
  description = "Test VM"
  vcpus = 2
  bootloader = "UEFI"
  autostart = true
  time = "UTC"
  shutdown_timeout = "10"
  cores = 4
  threads = 2
  memory = 1024*1024*512 // 512MB

  device {
    type = "NIC"
    attributes = {
      type = "VIRTIO"
      mac = "00:a0:98:39:5b:78"
      nic_attach = "br4"
    }
  }

  device {
    type = "DISK"
    attributes = {
      path = "/dev/zvol/Tank/dev-3qsqd"
      type = "AHCI"
      physical_sectorsize = null
      logical_sectorsize = null
    }
  }

  device {
    type = "DISPLAY"
    attributes = {
      wait = false
      port = 9736
      resolution = "1024x768"
      bind = "0.0.0.0"
      password = ""
      web = true
      type = "VNC"
    }
  }
}
