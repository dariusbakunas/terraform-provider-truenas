package truenas

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceTruenasShareNFS_basic(t *testing.T) {
	resourceName := "data.truenas_share_nfs.nfs"

	suffix := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	dataset_name := fmt.Sprintf("%s-%s", testResourcePrefix, suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			// those work for the data source test as well
			testAccCheckResourceTruenasShareNFSDestroy,
			testAccCheckResourceTruenasShareNFSDatasetDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceTruenasShareNFSConfig(testPoolName, dataset_name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(resourceName, "paths.*", fmt.Sprintf("/mnt/%s/%s", testPoolName, dataset_name)),
					resource.TestCheckResourceAttr(resourceName, "comment", "Testing NFS share"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hosts.*", "10.1.0.1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hosts.*", "foo.bar.baz"),
					resource.TestCheckResourceAttr(resourceName, "alldirs", "false"),
					resource.TestCheckResourceAttr(resourceName, "ro", "true"),
					resource.TestCheckResourceAttr(resourceName, "quiet", "false"),
					// Testing the following reliably would require knowledge
					// about an existing user and activation state of NFSv4 support.
					// resource.TestCheckResourceAttr(resourceName, "maproot_user", "maproot_user"),
					// resource.TestCheckResourceAttr(resourceName, "maproot_group", "maproot_group"),
					// resource.TestCheckResourceAttr(resourceName, "mapall_user", "mapall_user"),
					// resource.TestCheckResourceAttr(resourceName, "mapall_group", "mapall_group"),
					// resource.TestCheckResourceAttr(resourceName, "security.0", "krb5i"),
					// resource.TestCheckResourceAttr(resourceName, "security.1", "sys"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckTypeSetElemAttr(resourceName, "networks.*", "10.1.1.0/24"),
				),
			},
		},
	})
}

func testAccCheckDataSourceTruenasShareNFSConfig(pool string, dataset_name string) string {
	return fmt.Sprintf(`
	resource "truenas_dataset" "test" {
		name = "%s"
		pool = "%s"
	}

	resource "truenas_share_nfs" "nfstest" {
		depends_on = [
			truenas_dataset.test,
		]
		paths = [
			resource.truenas_dataset.test.mount_point,
		]
		comment = "Testing NFS share"
		hosts = [
			"10.1.0.1",
			"foo.bar.baz",
		]
		alldirs = false
		ro = true
		quiet = false
		# maproot_user = "maproot_user"
		# maproot_group = "maproot_group"
		# mapall_user = "mapall_user"
		# mapall_group = "mapall_group"
		# security = [
		# 	"krb5i",
		# 	"sys",
		# ]
		enabled = true
		networks = [
			"10.1.1.0/24",
		]
	}

	data "truenas_share_nfs" "nfs" {
		sharenfs_id = resource.truenas_share_nfs.nfstest.sharenfs_id
	}
	`, dataset_name, pool)
}
