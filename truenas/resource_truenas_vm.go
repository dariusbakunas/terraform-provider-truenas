package truenas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceTrueNASVM() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
}
