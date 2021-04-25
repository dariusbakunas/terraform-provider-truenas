package truenas

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
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
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pools, err := c.PoolAPI.List(ctx, &ListOptions{})

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Received TrueNAS pools: %+q", pools)

	converted := flattenPoolsResponse(pools)

	if err := d.Set("ids", converted); err != nil {
		return diag.FromErr(err)
	}

	// always run
	// TODO: not sure yet what should go here, find out
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenPoolsResponse(pools []Pool) []int64 {
	result := make([]int64, len(pools))

	for i, pool := range pools {
		result[i] = pool.ID
	}

	return result
}
