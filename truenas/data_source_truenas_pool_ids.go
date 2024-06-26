package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)

func dataSourceTrueNASPoolIDs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTrueNASPoolsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func dataSourceTrueNASPoolsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pools, _, err := c.PoolApi.ListPools(ctx).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error getting pool ids: %s\n%s", err, body)
	}

	converted := flattenPoolsResponse(pools)

	if err := d.Set("ids", converted); err != nil {
		return diag.FromErr(err)
	}

	// always run
	// TODO: not sure yet what should go here, find out
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenPoolsResponse(pools []api.Pool) []int32 {
	result := make([]int32, len(pools))

	for i, pool := range pools {
		result[i] = pool.Id
	}

	return result
}
