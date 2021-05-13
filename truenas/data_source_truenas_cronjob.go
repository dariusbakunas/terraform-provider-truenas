package truenas

import (
	"context"
	"github.com/dariusbakunas/terraform-provider-truenas/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func dataSourceTrueNASCronjob() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about specific cronjob",
		ReadContext: dataSourceTrueNASCronjobRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Description: "Cronjob ID",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"user": &schema.Schema{
				Description: "Account that is used to run the job",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"command": &schema.Schema{
				Description: "Command or script that runs on schedule",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": &schema.Schema{
				Description: "Optional cronjob description",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"enabled": &schema.Schema{
				Description: "`true` if cronjob is enabled",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"hide_stdout": &schema.Schema{
				Description: "if `false` any standard output is mailed to the user account used to run the command",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"hide_stderr": &schema.Schema{
				Description: "if `false` any error output is mailed to the user account used to run the command",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"schedule": &schema.Schema{
				Description: "Cronjob schedule",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"minute": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"hour": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"dom": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"month": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"dow": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTrueNASCronjobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.Client)
	id := d.Get("id").(int)

	resp, err := c.CronjobAPI.Get(ctx, strconv.Itoa(id))

	if err != nil {
		return diag.Errorf("error getting cronjob: %s", err)
	}

	d.Set("user", resp.User)
	d.Set("command", resp.Command)
	d.Set("description", resp.Description)
	d.Set("enabled", resp.Enabled)
	d.Set("hide_stdout", resp.STDOUT)
	d.Set("hide_stderr", resp.STDERR)

	if err := d.Set("schedule", flattenSchedule(resp.Schedule)); err != nil {
		return diag.Errorf("error setting schedule: %s", err)
	}

	d.SetId(strconv.Itoa(resp.ID))

	return diags
}

func flattenSchedule(s api.JobSchedule) []interface{} {
	var res []interface{}

	schedule := map[string]interface{}{
		"minute": s.Minute,
		"hour":   s.Hour,
		"dom":    s.Dom,
		"month":  s.Month,
		"dow":    s.Dow,
	}

	res = append(res, schedule)

	return res
}
