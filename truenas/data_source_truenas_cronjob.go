package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func dataSourceTrueNASCronjob() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about specific cronjob",
		ReadContext: dataSourceTrueNASCronjobRead,
		Schema: map[string]*schema.Schema{
			"cronjob_id": &schema.Schema{
				Description: "Cronjob ID",
				Type:        schema.TypeString,
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

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Get("cronjob_id").(string))

	if err != nil {
		return diag.FromErr(err)
	}

	resp, _, err := c.CronjobApi.GetCronJob(ctx, int32(id)).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error getting cronjob: %s\n%s", err, body)
	}

	if resp.User != nil {
		d.Set("user", *resp.User)
	}

	if resp.Command != nil {
		d.Set("command", *resp.Command)
	}

	if resp.Description != nil {
		d.Set("description", *resp.Description)
	}

	d.Set("enabled", *resp.Enabled)
	d.Set("hide_stdout", *resp.Stdout)
	d.Set("hide_stderr", *resp.Stderr)

	if resp.Schedule != nil {
		if err := d.Set("schedule", flattenSchedule(*resp.Schedule)); err != nil {
			return diag.Errorf("error setting schedule: %s", err)
		}
	}

	d.SetId(strconv.Itoa(int(*resp.Id)))

	return diags
}

func flattenSchedule(s api.CronJobSchedule) []interface{} {
	var res []interface{}

	schedule := map[string]interface{}{}

	if s.Minute != nil {
		schedule["minute"] = *s.Minute
	}

	if s.Hour != nil {
		schedule["hour"] = *s.Hour
	}

	if s.Dom != nil {
		schedule["dom"] = *s.Dom
	}

	if s.Month != nil {
		schedule["month"] = *s.Month
	}

	if s.Dow != nil {
		schedule["dow"] = *s.Dow
	}

	res = append(res, schedule)

	return res
}
