package truenas

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceTruenasCronjob_basic(t *testing.T) {
	resourceName := "data.truenas_cronjob.cj"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceTruenasCronjobConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user", "root"),
					resource.TestCheckResourceAttr(resourceName, "command", "ls"),
					resource.TestCheckResourceAttr(resourceName, "description", "tf cron job"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.dom", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.dow", "2"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.hour", "3"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.minute", "5"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.month", "4"),
				),
			},
		},
	})
}

func testAccCheckDataSourceTruenasCronjobConfig() string {
	return fmt.Sprintf(`
		resource "truenas_cronjob" "cj" {
			user = "root"
			command = "ls"
			description = "tf cron job"
			schedule {
				dom = "1"
				dow = "2"
				hour = "3"
				minute = "5"
				month = "4"
			}
		}

		data "truenas_cronjob" "cj" {
		  cronjob_id = truenas_cronjob.cj.cronjob_id
		}
	`)
}
