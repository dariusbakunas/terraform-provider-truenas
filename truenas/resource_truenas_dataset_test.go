package truenas

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccTruenasDataset_basic(t *testing.T) {
	var dataset DatasetResponse
	pool := "Tank"
	suffix := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	name := fmt.Sprintf("%s-%s", testResourcePrefix, suffix)
	resourceName := "truenas_dataset.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTruenasDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckTruenasDatasetResource(pool, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
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
					testAccCheckTruenasDatasetResourceExists(resourceName, &dataset),
				),
			},
		},
	})
}

func testAccCheckTruenasDatasetDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	// loop through the resources in state, verifying each widget
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "truenas_dataset" {
			continue
		}

		// Try to find the dataset
		_, err := client.DatasetAPI.Get(context.Background(), rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("dataset (%s) still exists", rs.Primary.ID)
		}

		// TODO: check if error is in fact 404
	}

	return nil
}

func Test_datasetPathString(t *testing.T) {
	testcases := []struct {
		path     datasetPath
		expected string
	}{
		{path: datasetPath{Pool: "Tank", Parent: "home", Name: "Test"}, expected: "Tank/home/Test"},
		{path: datasetPath{Pool: "Tank", Parent: "", Name: "Test"}, expected: "Tank/Test"},
		{path: datasetPath{Pool: "Tank", Parent: "//home/sub//", Name: "Test"}, expected: "Tank/home/sub/Test"},
		{path: datasetPath{Pool: "TankV2", Parent: "/home/", Name: "Test"}, expected: "TankV2/home/Test"},
	}

	for _, c := range testcases {
		actual := c.path.String()
		assert.Equal(t, actual, c.expected)
	}
}

func Test_newDatasetPath(t *testing.T) {
	testcases := []struct {
		path     string
		expected datasetPath
	}{
		{expected: datasetPath{Pool: "Tank", Parent: "home", Name: "Test"}, path: "Tank/home/Test"},
		{expected: datasetPath{Pool: "Tank", Parent: "", Name: "Test"}, path: "Tank/Test"},
		{expected: datasetPath{Pool: "Tank", Parent: "home/sub", Name: "Test"}, path: "Tank/home/sub/Test"},
		{expected: datasetPath{Pool: "TankV2", Parent: "home", Name: "Test"}, path: "TankV2/home/Test"},
	}

	for _, c := range testcases {
		actual := newDatasetPath(c.path)
		assert.Equal(t, actual, c.expected)
	}
}

func testAccCheckTruenasDatasetResource(pool string, name string) string {
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
	`, name, pool)
}

func testAccCheckTruenasDatasetResourceExists(n string, dataset *DatasetResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("dataset resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no dataset ID is set")
		}

		client := testAccProvider.Meta().(*Client)

		resp, err := client.DatasetAPI.Get(context.Background(), rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp.ID != rs.Primary.ID {
			return fmt.Errorf("dataset not found")
		}

		*dataset = *resp
		return nil
	}
}
