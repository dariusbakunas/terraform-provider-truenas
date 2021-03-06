package truenas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"os"
	"testing"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

const testResourcePrefix = "tf-acc-test"

var testPoolName string

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"truenas": testAccProvider,
	}

	if v := os.Getenv("TRUENAS_POOL_NAME"); v != "" {
		testPoolName = v
	}

	if testPoolName == "" {
		log.Printf("[WARN] Env TRUENAS_POOL_NAME was not specified, using default 'Tank'")
		testPoolName = "Tank"
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
