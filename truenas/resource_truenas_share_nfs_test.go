package truenas

import (
	"context"
	"fmt"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"testing"
)

func TestAccResourceTruenasShareNFS_basic(t *testing.T) {
	resourceName := "truenas_share_nfs.nfs"

	var share api.ShareNFS

	suffix := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	datasetName := fmt.Sprintf("%s-%s", testResourcePrefix, suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckResourceTruenasShareNFSDestroy,
			testAccCheckResourceTruenasShareNFSDatasetDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceTruenasShareNFSConfig(testPoolName, datasetName),
				Check: resource.ComposeTestCheckFunc(
					// make sure the API reports the expected values
					testAccCheckTruenasShareNFSResourceExists(resourceName, &share),
					testAccCheckTruenasShareNFSResourceAttributes(t, resourceName, &share, testPoolName, datasetName),
					// make sure the Terraform resource matches
					resource.TestCheckTypeSetElemAttr(resourceName, "paths.*", fmt.Sprintf("/mnt/%s/%s", testPoolName, datasetName)),
					resource.TestCheckResourceAttr(resourceName, "comment", "Testing NFS share"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hosts.*", "8.8.8.8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hosts.*", "google.com"),
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
					resource.TestCheckTypeSetElemAttr(resourceName, "networks.*", "10.128.0.0/9"),
				),
			},
		},
	})
}

func testAccCheckResourceTruenasShareNFSConfig(pool string, datasetName string) string {
	return fmt.Sprintf(`
	resource "truenas_dataset" "test" {
		name = "%s"
		pool = "%s"
	}

	resource "truenas_share_nfs" "nfs" {
		paths = [
			resource.truenas_dataset.test.mount_point,
		]
		comment = "Testing NFS share"
		hosts = [
			"8.8.8.8",
			"google.com",
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
			"10.128.0.0/9",
		]
	}
	`, datasetName, pool)
}

func testAccCheckTruenasShareNFSResourceExists(n string, share *api.ShareNFS) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("nfs share resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no nfs share ID is set")
		}

		client := testAccProvider.Meta().(*api.APIClient)

		id, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not convert ID of nfs share: %s", n)
		}

		resp, _, err := client.SharingApi.GetShareNFS(context.Background(), int32(id)).Execute()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(resp.Id)) != rs.Primary.ID {
			return fmt.Errorf("nfs share not found")
		}

		*share = *resp
		return nil
	}
}

func testAccCheckTruenasShareNFSResourceAttributes(t *testing.T, n string, share *api.ShareNFS, pool string, dataset string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// order does not matter
		if !assert.ElementsMatch(t, share.Paths, []string{fmt.Sprintf("/mnt/%s/%s", pool, dataset)}) {
			return fmt.Errorf("remote paths for nfs share do not match expected")
		}

		if *share.Comment != "Testing NFS share" {
			return fmt.Errorf("remote comment for nfs share does not match expected")
		}

		if !assert.ElementsMatch(t, share.Hosts, []string{"8.8.8.8", "google.com"}) {
			return fmt.Errorf("remote hosts for nfs share do not match expected")
		}

		if *share.Alldirs != false {
			return fmt.Errorf("remote alldirs for nfs share does not match expected")
		}

		if *share.Ro != true {
			return fmt.Errorf("remote ro for nfs share does not match expected")
		}

		if *share.Quiet != false {
			return fmt.Errorf("remote quiet for nfs share does not match expected")
		}

		if share.MaprootUser != nil {
			return fmt.Errorf("remote maproot_user for nfs share not empty as expected")
		}

		if share.MaprootGroup != nil {
			return fmt.Errorf("remote maproot_group for nfs share not empty as expected")
		}

		if share.MapallUser != nil {
			return fmt.Errorf("remote mapall_user for nfs share not empty as expected")
		}

		if share.MapallGroup != nil {
			return fmt.Errorf("remote mapall_group for nfs share not empty as expected")
		}

		// for security, order would matter
		if !reflect.DeepEqual(share.Security, []string{}) {
			return fmt.Errorf("remote security for nfs share not empty as expected: %#v", share.Security)
		}

		if *share.Enabled != true {
			return fmt.Errorf("remote enabled for nfs share does not match expected")
		}

		if !assert.ElementsMatch(t, share.Networks, []string{"10.128.0.0/9"}) {
			return fmt.Errorf("remote networks for nfs share do not match expected")
		}

		return nil
	}
}

func testAccCheckResourceTruenasShareNFSDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*api.APIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "truenas_share_nfs" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not convert ID of share: %s", rs.Primary.ID)
		}

		// Try to find the nfs share on the node
		_, r, err := client.SharingApi.GetShareNFS(context.Background(), int32(id)).Execute()

		if err == nil {
			return fmt.Errorf("nfs share (%s) still exists", rs.Primary.ID)
		}

		// check if error is in fact 404 (not found)
		if r.StatusCode != 404 {
			return fmt.Errorf("Error occured while checking for absence of nfs share (%s)", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckResourceTruenasShareNFSDatasetDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*api.APIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "truenas_dataset" {
			continue
		}

		// Try to find the dataset on the node
		_, r, err := client.DatasetApi.GetDataset(context.Background(), rs.Primary.ID).Execute()

		if err == nil {
			return fmt.Errorf("dataset (%s) created for nfs share still exists", rs.Primary.ID)
		}

		// check if error is in fact 404 (not found)
		if r.StatusCode != 404 {
			return fmt.Errorf("Error occured while checking for absence of dataset (%s) created for nfs share", rs.Primary.ID)
		}
	}

	return nil
}
