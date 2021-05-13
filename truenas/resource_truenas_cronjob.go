package truenas

import (
	"context"
	"github.com/dariusbakunas/terraform-provider-truenas/api"
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
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
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

	c := m.(*api.Client)
	id := d.Id()

	resp, err := c.CronjobAPI.Get(ctx, id)

	if err != nil {
		return diag.Errorf("error getting cronjob: %s", err)
	}

	d.Set("id", strconv.Itoa(resp.ID))
	d.Set("user", resp.User)
	d.Set("command", resp.Command)
	d.Set("description", resp.Description)
	d.Set("enabled", resp.Enabled)
	d.Set("hide_stdout", resp.STDOUT)
	d.Set("hide_stderr", resp.STDERR)

	if err := d.Set("schedule", flattenSchedule(resp.Schedule)); err != nil {
		return diag.Errorf("error setting schedule: %s", err)
	}

	return diags
}

func resourceTrueNASCronjobCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.Client)
	job := expandJobInput(d)

	resp, err := c.CronjobAPI.Create(ctx, job)

	if err != nil {
		return diag.Errorf("error creating cronjob: %s", err)
	}

	d.SetId(strconv.Itoa(resp.ID))

	return resourceTrueNASCronjobRead(ctx, d, m)
}

func resourceTrueNASCronjobUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.Client)
	job := expandJobInput(d)
	id := d.Id()

	_, err := c.CronjobAPI.Update(ctx, id, job)

	if err != nil {
		return diag.Errorf("error updating cronjob: %s", err)
	}

	return resourceTrueNASCronjobRead(ctx, d, m)
}

func resourceTrueNASCronjobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.Client)
	id := d.Id()

	err := c.CronjobAPI.Delete(ctx, id)

	if err != nil {
		return diag.Errorf("error deleting cronjob: %s", err)
	}

	return diags
}

func expandJobInput(d *schema.ResourceData) *api.JobInput {
	job := &api.JobInput{
		Command: d.Get("command").(string),
		User:    d.Get("user").(string),
	}

	if description, ok := d.GetOk("description"); ok {
		job.Description = description.(string)
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
		job.STDOUT = getBoolPtr(stdout.(bool))
	}

	stderr := d.Get("hide_stderr")

	if stderr != nil {
		job.STDERR = getBoolPtr(stderr.(bool))
	}

	return job
}

func expandJobSchedule(s []interface{}) *api.JobSchedule {
	schedule := &api.JobSchedule{}

	if len(s) == 0 || s[0] == nil {
		return nil
	}

	mSchedule := s[0].(map[string]interface{})

	schedule.Minute = mSchedule["minute"].(string)
	schedule.Hour = mSchedule["hour"].(string)
	schedule.Dom = mSchedule["dom"].(string)
	schedule.Month = mSchedule["month"].(string)
	schedule.Dow = mSchedule["dow"].(string)

	return schedule
}
