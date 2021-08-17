package truenas

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceTruenasZvol_basic(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	name := fmt.Sprintf("%s-%s", testResourcePrefix, suffix)
	resourceName := "data.truenas_zvol.zv"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceTruenasZvolConfig(testPoolName, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "blocksize", "64K"),
					resource.TestCheckResourceAttr(resourceName, "pool", testPoolName),
					resource.TestCheckResourceAttr(resourceName, "comments", "Test dataset"),
					resource.TestCheckResourceAttr(resourceName, "compression", "gzip"),
					resource.TestCheckResourceAttr(resourceName, "sync", "standard"),
					resource.TestCheckResourceAttr(resourceName, "copies", "1"),
					resource.TestCheckResourceAttr(resourceName, "deduplication", "off"),
					resource.TestCheckResourceAttr(resourceName, "readonly", "off"),
					resource.TestCheckResourceAttr(resourceName, "volsize", "1073741824"),
				),
			},
		},
	})
}

func testAccCheckDataSourceTruenasZvolConfig(pool string, name string) string {
	return fmt.Sprintf(`
		resource "truenas_zvol" "test" {
			name = "%s"
			pool = "%s"
			blocksize = "64K"
			comments = "Test dataset"
			compression = "gzip"
			sync = "standard"
			deduplication = "off"		
			readonly = "off"
			volsize = 1024*1024*1024
		}

		data "truenas_zvol" "zv" {
			id = truenas_zvol.test.id
		}
	`, name, pool)
}
