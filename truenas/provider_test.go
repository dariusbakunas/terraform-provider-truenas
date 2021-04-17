package truenas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"testing"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

const testResourcePrefix = "tf-acc-test-"

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"truenas": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("TRUENAS_API_KEY"); v == "" {
		t.Fatal("TRUENAS_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("TRUENAS_BASE_URL"); v == "" {
		t.Fatal("TRUENAS_BASE_URL must be set for acceptance tests")
	}
}
