package truenas

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strings"
	"testing"
)

func TestAccDataSourceTruenasVM_basic(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	// VM name must be alphanumeric
	name := fmt.Sprintf("%s%s", strings.Replace(testResourcePrefix, "-", "", -1), suffix)
	resourceName := "data.truenas_vm.vm"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceTruenasVMConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "Test VM"),
					resource.TestCheckResourceAttr(resourceName, "vcpus", "2"),
					resource.TestCheckResourceAttr(resourceName, "bootloader", "UEFI"),
					resource.TestCheckResourceAttr(resourceName, "autostart", "true"),
					resource.TestCheckResourceAttr(resourceName, "time", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "shutdown_timeout", "10"),
					resource.TestCheckResourceAttr(resourceName, "cores", "4"),
					resource.TestCheckResourceAttr(resourceName, "threads", "2"),
					resource.TestCheckResourceAttr(resourceName, "memory", "536870912"),
					resource.TestCheckResourceAttr(resourceName, "device.#", "3"),
					// TODO: not exactly sure how to test nested device attributes
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						"device.*",
						map[string]string{
							"type": "NIC",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						"device.*",
						map[string]string{
							"type": "DISPLAY",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						"device.*",
						map[string]string{
							"type": "DISK",
						},
					),
				),
			},
		},
	})
}

func testAccCheckDataSourceTruenasVMConfig(name string) string {
	return fmt.Sprintf(`
		resource "truenas_vm" "vm" {
		  name = "%s"
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
			  mac = "00:a1:98:39:5b:76"
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
			  port = 9799
			  resolution = "1024x768"
			  bind = "0.0.0.0"
			  password = ""
			  web = true
			  type = "VNC"
			}
		  }
		}

		data "truenas_vm" "vm" {
			vm_id = truenas_vm.vm.vm_id
		}
	`, name)
}