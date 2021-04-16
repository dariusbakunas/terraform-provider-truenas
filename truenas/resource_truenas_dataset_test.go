package truenas

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccTruenasDataset_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckTruenasDatasetDestroy,
		Steps: []resource.TestStep{},
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
		_, err := client.Datasets.Get(context.Background(), rs.Primary.ID)

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
