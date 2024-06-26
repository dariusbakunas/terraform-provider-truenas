package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func resourceTrueNASCronjob() *schema.Resource {
	return &schema.Resource{
		Description:   "TrueNAS allows users to run specific commands or scripts on a regular schedule using cron(8). This can be helpful for running repetitive tasks.",
		CreateContext: resourceTrueNASCronjobCreate,
		ReadContext:   resourceTrueNASCronjobRead,
		UpdateContext: resourceTrueNASCronjobUpdate,
		DeleteContext: resourceTrueNASCronjobDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"cronjob_id": &schema.Schema{
				Description: "Cronjob ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"user": &schema.Schema{
				Description: "Account that is used to run the job",
				Type:        schema.TypeString,
				Required:    true,
			},
			"command": &schema.Schema{
				Description: "Command or script that runs on schedule",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": &schema.Schema{
				Description: "Optional cronjob description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": &schema.Schema{
				Description: "`true` if cronjob is enabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"hide_stdout": &schema.Schema{
				Description: "if `false` any standard output is mailed to the user account used to run the command",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"hide_stderr": &schema.Schema{
				Description: "if `false` any error output is mailed to the user account used to run the command",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"schedule": &schema.Schema{
				Description: "Cronjob schedule",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"minute": &schema.Schema{
							Type:     schema.TypeString,
							Default:  "00",
							Optional: true,
						},
						"hour": &schema.Schema{
							Type:     schema.TypeString,
							Default:  "*",
							Optional: true,
						},
						"dom": &schema.Schema{
							Type:     schema.TypeString,
							Default:  "*",
							Optional: true,
						},
						"month": &schema.Schema{
							Type:     schema.TypeString,
							Default:  "*",
							Optional: true,
						},
						"dow": &schema.Schema{
							Type:     schema.TypeString,
							Default:  "*",
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceTrueNASCronjobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)

	id, err := strconv.Atoi(d.Id())

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

	d.Set("cronjob_id", strconv.Itoa(int(*resp.Id)))
	d.Set("user", *resp.User)
	d.Set("command", *resp.Command)

	if resp.Description != nil {
		d.Set("description", *resp.Description)
	}

	if resp.Enabled != nil {
		d.Set("enabled", *resp.Enabled)
	}

	if resp.Stdout != nil {
		d.Set("hide_stdout", *resp.Stdout)
	}

	if resp.Stderr != nil {
		d.Set("hide_stderr", resp.Stderr)
	}

	if err := d.Set("schedule", flattenSchedule(*resp.Schedule)); err != nil {
		return diag.Errorf("error setting schedule: %s", err)
	}

	return diags
}

func resourceTrueNASCronjobCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)
	job := expandJobInput(d)

	resp, _, err := c.CronjobApi.CreateCronJob(ctx).
		CreateCronjobParams(job).
		Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error creating cronjob: %s\n%s", err, body)
	}

	d.SetId(strconv.Itoa(int(*resp.Id)))

	return resourceTrueNASCronjobRead(ctx, d, m)
}

func resourceTrueNASCronjobUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)
	job := expandJobInput(d)

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, _, err = c.CronjobApi.UpdateCronJob(ctx, int32(id)).CreateCronjobParams(job).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error updating cronjob: %s\n%s", err, body)
	}

	return resourceTrueNASCronjobRead(ctx, d, m)
}

func resourceTrueNASCronjobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.CronjobApi.DeleteCronJob(ctx, int32(id)).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error deleting cronjob: %s\n%s", err, body)
	}
	d.SetId("")

	return diags
}

func expandJobInput(d *schema.ResourceData) api.CreateCronjobParams {
	job := api.CreateCronjobParams{
		Command: d.Get("command").(string),
		User:    d.Get("user").(string),
	}

	if description, ok := d.GetOk("description"); ok {
		job.Description = getStringPtr(description.(string))
	}

	if schedule, ok := d.GetOk("schedule"); ok {
		job.Schedule = expandJobSchedule(schedule.([]interface{}))
	}

	enabled := d.Get("enabled")

	if enabled != nil {
		job.Enabled = getBoolPtr(enabled.(bool))
	}

	stdout := d.Get("hide_stdout")

	if stdout != nil {
		job.Stdout = getBoolPtr(stdout.(bool))
	}

	stderr := d.Get("hide_stderr")

	if stderr != nil {
		job.Stderr = getBoolPtr(stderr.(bool))
	}

	return job
}

func expandJobSchedule(s []interface{}) *api.CronJobSchedule {
	schedule := &api.CronJobSchedule{}

	if len(s) == 0 || s[0] == nil {
		return nil
	}

	mSchedule := s[0].(map[string]interface{})

	schedule.Minute = getStringPtr(mSchedule["minute"].(string))
	schedule.Hour = getStringPtr(mSchedule["hour"].(string))
	schedule.Dom = getStringPtr(mSchedule["dom"].(string))
	schedule.Month = getStringPtr(mSchedule["month"].(string))
	schedule.Dow = getStringPtr(mSchedule["dow"].(string))

	return schedule
}
