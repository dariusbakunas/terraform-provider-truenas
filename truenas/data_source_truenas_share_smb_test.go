package truenas

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceTruenasShareSMB_basic(t *testing.T) {
	resourceName := "data.truenas_share_smb.smb"

	suffix := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	dataset_name := fmt.Sprintf("%s-%s", testResourcePrefix, suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			// those work for the data source test as well
			testAccCheckResourceTruenasShareSMBDestroy,
			testAccCheckResourceTruenasShareSMBDatasetDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceTruenasShareSMBConfig(testPoolName, dataset_name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "path", fmt.Sprintf("/mnt/%s/%s", testPoolName, dataset_name)),
					resource.TestCheckResourceAttr(resourceName, "path_suffix", "%U"),
					resource.TestCheckResourceAttr(resourceName, "purpose", "NO_PRESET"),
					resource.TestCheckResourceAttr(resourceName, "comment", "Testing SMB share"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hostsallow.*", "10.1.0.1/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hostsallow.*", "foo.bar.baz"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hostsdeny.*", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "home", "false"),
					resource.TestCheckResourceAttr(resourceName, "timemachine", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", dataset_name),
					resource.TestCheckResourceAttr(resourceName, "ro", "true"),
					resource.TestCheckResourceAttr(resourceName, "browsable", "false"),
					resource.TestCheckResourceAttr(resourceName, "recyclebin", "false"),
					resource.TestCheckResourceAttr(resourceName, "shadowcopy", "true"),
					resource.TestCheckResourceAttr(resourceName, "guestok", "false"),
					resource.TestCheckResourceAttr(resourceName, "aapl_name_mangling", "false"),
					resource.TestCheckResourceAttr(resourceName, "abe", "false"),
					resource.TestCheckResourceAttr(resourceName, "acl", "true"),
					resource.TestCheckResourceAttr(resourceName, "durablehandle", "false"),
					resource.TestCheckResourceAttr(resourceName, "fsrvp", "false"),
					resource.TestCheckResourceAttr(resourceName, "streams", "false"),
					resource.TestCheckResourceAttr(resourceName, "auxsmbconf", "vfs objects=zfs_space zfsacl streams_xattr"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckDataSourceTruenasShareSMBConfig(pool string, dataset_name string) string {
	return fmt.Sprintf(`
	resource "truenas_dataset" "test" {
		name = "%s"
		pool = "%s"
	}

	resource "truenas_share_smb" "smbtest" {
		path = resource.truenas_dataset.test.mount_point
		path_suffix = "%%U"
		purpose = "NO_PRESET"
		comment = "Testing SMB share"
		hostsallow = [
		  "10.1.0.1/24",
		  "foo.bar.baz",
		]
		hostsdeny = [
		  "ALL",
		]
		home = false
		timemachine = false
		name = "%s"
		ro = true
		browsable = false
		recyclebin = false
		shadowcopy = true
		guestok = false
		aapl_name_mangling = false
		abe = false
		acl = true
		durablehandle = false
		fsrvp = false
		streams = false
		auxsmbconf = "vfs objects=zfs_space zfsacl streams_xattr"
		enabled = true
	}

	data "truenas_share_smb" "smb" {
		sharesmb_id = resource.truenas_share_smb.smbtest.sharesmb_id
	}
	`, dataset_name, pool, dataset_name)
}
