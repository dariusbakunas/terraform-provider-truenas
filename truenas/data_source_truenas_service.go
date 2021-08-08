package truenas

import (
	"context"
	api "github.com/dariusbakunas/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"strings"
)

func dataSourceTrueNASService() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about system service",
		ReadContext: dataSourceTrueNASServiceRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Description: "Service ID",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"name": &schema.Schema{
				Description: "Service name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"enabled": &schema.Schema{
				Description: "`true` if service is enabled",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"pids": &schema.Schema{
				Description: "List of pids that belong to service",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Computed: true,
			},
			"state": &schema.Schema{
				Description: "Current state: `stopped`, `running`",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceTrueNASServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id := d.Get("id").(int)

	resp, _, err := c.ServiceApi.GetService(ctx, int32(id)).Execute()

	if err != nil {
		return diag.Errorf("error getting service: %s", err)
	}

	d.Set("name", resp.Service)
	d.Set("enabled", *resp.Enable)

	if resp.Pids != nil {
		d.Set("pids", flattenInt32List(*resp.Pids))
	}

	if resp.State != nil {
		d.Set("state", strings.ToLower(*resp.State))
	}

	d.SetId(strconv.Itoa(int(resp.Id)))

	return diags
}
