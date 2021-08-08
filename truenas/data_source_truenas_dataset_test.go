package truenas

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceTruenasDataset_basic(t *testing.T) {
	pool := "Tank"
	suffix := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	name := fmt.Sprintf("%s-%s", testResourcePrefix, suffix)
	resourceName := "data.truenas_dataset.ds"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceTruenasDatasetConfig(pool, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "pool", pool),
					resource.TestCheckResourceAttr(resourceName, "comments", "Test dataset"),
					resource.TestCheckResourceAttr(resourceName, "sync", "standard"),
					resource.TestCheckResourceAttr(resourceName, "atime", "off"),
					resource.TestCheckResourceAttr(resourceName, "copies", "2"),
					resource.TestCheckResourceAttr(resourceName, "quota_bytes", "2147483648"),
					resource.TestCheckResourceAttr(resourceName, "quota_critical", "90"),
					resource.TestCheckResourceAttr(resourceName, "quota_warning", "70"),
					resource.TestCheckResourceAttr(resourceName, "ref_quota_bytes", "1073741824"),
					resource.TestCheckResourceAttr(resourceName, "ref_quota_critical", "90"),
					resource.TestCheckResourceAttr(resourceName, "ref_quota_warning", "70"),
					resource.TestCheckResourceAttr(resourceName, "deduplication", "off"),
					resource.TestCheckResourceAttr(resourceName, "exec", "on"),
					resource.TestCheckResourceAttr(resourceName, "snap_dir", "hidden"),
					resource.TestCheckResourceAttr(resourceName, "readonly", "off"),
					resource.TestCheckResourceAttr(resourceName, "record_size", "256K"),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "mixed"),
				),
			},
		},
	})
}

func testAccCheckDataSourceTruenasDatasetConfig(pool string, name string) string {
	return fmt.Sprintf(`
		resource "truenas_dataset" "test" {
			name = "%s"
			pool = "%s"
			comments = "Test dataset"
			compression = "gzip"
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
		}

		data "truenas_dataset" "ds" {
			id = truenas_dataset.test.id
		}
	`, name, pool)
}
