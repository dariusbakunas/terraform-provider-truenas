package truenas

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)

func dataSourceTrueNASPools() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTrueNASPoolsRead,
		Schema: map[string]*schema.Schema{
			"pools": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTrueNASPoolsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	pools, _, err := c.Pools.List(ctx, &ListOptions{})

	if err != nil {
		return diag.FromErr(err)
	}

	converted := convertPools(pools)

	if err := d.Set("pools", converted); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func convertPools(pools []*Pool) []interface{} {
	if pools != nil {
		result := make([]interface{}, len(pools), len(pools))

		for i, pool := range pools {
			p := make(map[string]interface{})

			p["id"] = *pool.ID
			p["name"] = *pool.Name
			p["path"] = *pool.Path

			result[i] = p
		}

		return result
	}

	return make([]interface{}, 0)
}
