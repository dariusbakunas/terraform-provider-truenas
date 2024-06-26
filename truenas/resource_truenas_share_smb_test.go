package truenas

import (
	"context"
	"fmt"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestAccResourceTruenasShareSMB_basic(t *testing.T) {
	resourceName := "truenas_share_smb.smb"

	var share api.ShareSMB

	suffix := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	datasetName := fmt.Sprintf("%s-%s", testResourcePrefix, suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckResourceTruenasShareSMBDestroy,
			testAccCheckResourceTruenasShareSMBDatasetDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceTruenasShareSMBConfig(testPoolName, datasetName),
				Check: resource.ComposeTestCheckFunc(
					// make sure the API reports the expected values
					testAccCheckTruenasShareSMBResourceExists(resourceName, &share),
					testAccCheckTruenasShareSMBResourceAttributes(t, resourceName, &share, testPoolName, datasetName),
					// make sure the Terraform resource matches
					resource.TestCheckResourceAttr(resourceName, "path", fmt.Sprintf("/mnt/%s/%s", testPoolName, datasetName)),
					resource.TestCheckResourceAttr(resourceName, "path_suffix", "%U"),
					resource.TestCheckResourceAttr(resourceName, "purpose", "NO_PRESET"),
					resource.TestCheckResourceAttr(resourceName, "comment", "Testing SMB share"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hostsallow.*", "10.1.0.1/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hostsallow.*", "foo.bar.baz"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hostsdeny.*", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "home", "false"),
					resource.TestCheckResourceAttr(resourceName, "timemachine", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", datasetName),
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

func Test_lockedSMBProfileParamIsRejected(t *testing.T) {
	locked_map := map[string][]string{
		"NO_PRESET":            []string{},
		"DEFAULT_SHARE":        []string{"Acl", "Ro", "Browsable", "Abe", "Hostsallow", "Hostsdeny", "Home", "Timemachine", "Shadowcopy", "Recyclebin", "AaplNameMangling", "Streams", "Durablehandle", "Fsrvp", "PathSuffix"},
		"ENHANCED_TIMEMACHINE": []string{"Timemachine", "PathSuffix"},
		"MULTI_PROTOCOL_NFS":   []string{"Acl", "Streams", "Durablehandle"},
		"PRIVATE_DATASETS":     []string{"PathSuffix"},
		"WORM_DROPBOX":         []string{"PathSuffix"},
	}

	share := api.CreateShareSMBParams{
		Path: "/mnt/test/path",
	}

	for preset, locked_params := range locked_map {
		share.Purpose = &preset
		for _, locked := range locked_params {
			err := isParamLocked(locked, &share)
			if assert.NotNil(t, err) {
				assert.Error(t, err)
			}
		}
	}
}

func testAccCheckResourceTruenasShareSMBConfig(pool string, datasetName string) string {
	return fmt.Sprintf(`
	resource "truenas_dataset" "test" {
		name = "%s"
		pool = "%s"
	}

	resource "truenas_share_smb" "smb" {
		path = truenas_dataset.test.mount_point
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
	`, datasetName, pool, datasetName)
}

func testAccCheckTruenasShareSMBResourceExists(n string, share *api.ShareSMB) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("smb share resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no smb share ID is set")
		}

		client := testAccProvider.Meta().(*api.APIClient)

		id, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not convert ID of share: %s", n)
		}

		resp, _, err := client.SharingApi.GetShareSMB(context.Background(), int32(id)).Execute()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(resp.Id)) != rs.Primary.ID {
			return fmt.Errorf("smb share not found")
		}

		share = resp
		return nil
	}
}

func testAccCheckTruenasShareSMBResourceAttributes(t *testing.T, n string, share *api.ShareSMB, pool string, dataset string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if share.Path != fmt.Sprintf("/mnt/%s/%s", testPoolName, dataset) {
			return fmt.Errorf("remote path for smb share does not match expected")
		}

		if *share.PathSuffix != "%U" {
			return fmt.Errorf("remote path_suffix for smb share does not match expected")
		}

		if *share.Purpose != "NO_PRESET" {
			return fmt.Errorf("remote purpose for smb share does not match expected")
		}

		if *share.Comment != "Testing SMB share" {
			return fmt.Errorf("remote comment for smb share does not match expected")
		}

		// order does not matter
		if !assert.ElementsMatch(t, share.Hostsallow, []string{"10.1.0.1/24", "foo.bar.baz"}) {
			return fmt.Errorf("remote hostsallow list for smb share does not match expected")
		}

		if !assert.ElementsMatch(t, share.Hostsdeny, []string{"ALL"}) {
			return fmt.Errorf("remote hostsdeny list for smb share does not match expected")
		}

		if *share.Home != false {
			return fmt.Errorf("remote home for smb share does not match expected")
		}

		if *share.Timemachine != false {
			return fmt.Errorf("remote timemachine for smb share does not match expected")
		}

		if *share.Name != dataset {
			return fmt.Errorf("remote name for smb share does not match expected")
		}

		if *share.Ro != true {
			return fmt.Errorf("remote ro for smb share does not match expected")
		}

		if *share.Browsable != false {
			return fmt.Errorf("remote browsable for smb share does not match expected")
		}

		if *share.Recyclebin != false {
			return fmt.Errorf("remote recyclebin for smb share does not match expected")
		}

		if *share.Shadowcopy != true {
			return fmt.Errorf("remote shadowcopy for smb share does not match expected")
		}

		if *share.Guestok != false {
			return fmt.Errorf("remote guestok for smb share does not match expected")
		}

		if *share.AaplNameMangling != false {
			return fmt.Errorf("remote aapl_name_mangling for smb share does not match expected")
		}

		if *share.Abe != false {
			return fmt.Errorf("remote abe for smb share does not match expected")
		}

		if *share.Acl != true {
			return fmt.Errorf("remote acl for smb share does not match expected")
		}

		if *share.Durablehandle != false {
			return fmt.Errorf("remote durablehandle for smb share does not match expected")
		}

		if *share.Fsrvp != false {
			return fmt.Errorf("remote fsrvp for smb share does not match expected")
		}

		if *share.Streams != false {
			return fmt.Errorf("remote streams for smb share does not match expected")
		}

		if *share.Auxsmbconf != "vfs objects=zfs_space zfsacl streams_xattr" {
			return fmt.Errorf("remote auxsmbconf for smb share does not match expected")
		}

		if *share.Enabled != true {
			return fmt.Errorf("remote enabled for smb share does not match expected")
		}

		return nil
	}
}

func testAccCheckResourceTruenasShareSMBDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*api.APIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "truenas_share_smb" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not convert ID of smb share: %s", rs.Primary.ID)
		}

		// Try to find the smb share on the node
		_, r, err := client.SharingApi.GetShareSMB(context.Background(), int32(id)).Execute()

		if err == nil {
			return fmt.Errorf("smb share (%s) still exists", rs.Primary.ID)
		}

		// check if error is in fact 404 (not found)
		if r.StatusCode != 404 {
			return fmt.Errorf("Error occured while checking for absence of smb share (%s)", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckResourceTruenasShareSMBDatasetDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*api.APIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "truenas_dataset" {
			continue
		}

		// Try to find the dataset on the node
		_, r, err := client.DatasetApi.GetDataset(context.Background(), rs.Primary.ID).Execute()

		if err == nil {
			return fmt.Errorf("dataset (%s) created for smb share still exists", rs.Primary.ID)
		}

		// check if error is in fact 404 (not found)
		if r.StatusCode != 404 {
			return fmt.Errorf("Error occured while checking for absence of dataset (%s) created for smb share", rs.Primary.ID)
		}
	}

	return nil
}
