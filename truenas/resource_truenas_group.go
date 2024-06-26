package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strconv"
)

func resourceTrueNASGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates a group",
		CreateContext: resourceTrueNASGroupCreate,
		ReadContext:   resourceTrueNASGroupRead,
		UpdateContext: resourceTrueNASGroupUpdate,
		DeleteContext: resourceTrueNASGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"allow_duplicate_gid": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Allows distinct group names to share the same gid. (Used only on group creation. Not recommended.)",
				Optional:    true,
				Default:     false,
			},
			"name": &schema.Schema{
				Description:  "Unique identifier for the group.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringDoesNotContainAny("/ "),
			},
			"gid": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Group ID. If not provided, it is automatically filled with the next one available.",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"smb": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Specifies whether the group should be mapped into an NT group.",
				Optional:    true,
			},
			"sudo": &schema.Schema{
				Description: "Specifies whether group members may invoke sudo.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"sudo_commands": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"sudo_no_password": &schema.Schema{
				Description: "Specifies whether group members may invoke sudo without a password.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			//"user_ids": &schema.Schema{
			//	Description: "List of User IDs belonging to the group.",
			//	Type:        schema.TypeSet,
			//	Optional:    true,
			//	Elem: &schema.Schema{
			//		Type: schema.TypeInt,
			//	},
			//},
		},
	}
}

func resourceTrueNASGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resp, http, err := c.GroupApi.GetGroup(ctx, int32(id)).Execute()

	if err != nil {
		// gracefully handle manual deletions
		d.SetId("")
		if http.StatusCode == 404 {
			return nil
		}
		return diag.Errorf("error getting group: %s", err)
	}

	d.Set("name", resp.Group)

	if resp.Gid != nil {
		d.Set("gid", *resp.Gid)
	}

	if resp.Smb != nil {
		d.Set("smb", *resp.Smb)
	}

	if resp.Sudo != nil {
		d.Set("sudo", *resp.Sudo)
	}

	if resp.SudoCommands != nil {
		d.Set("sudo_commands", resp.SudoCommands)
	}

	if resp.SudoNopasswd != nil {
		d.Set("sudo_no_password", *resp.SudoNopasswd)
	}

	//if resp.Users != nil {
	//	d.Set("user_ids", resp.Users)
	//}

	return diags
}

func resourceTrueNASGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	input := expandGroup(d)

	resp, _, err := c.GroupApi.CreateGroup(ctx).CreateGroupParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error creating group: %s\n%s", err, body)
	}

	d.SetId(strconv.Itoa(int(resp)))

	return resourceTrueNASGroupRead(ctx, d, m)
}

func resourceTrueNASGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting TrueNAS group: %s", strconv.Itoa(id))

	input := api.DeleteGroupParams{
		DeleteUsers: getBoolPtr(false),
	}
	_, err = c.GroupApi.DeleteGroup(ctx, int32(id)).DeleteGroupParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error deleting group: %s\n%s", err, body)
	}

	log.Printf("[INFO] TrueNAS group (%s) deleted", d.Id())
	d.SetId("")

	return diags
}

func resourceTrueNASGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Uses the create group API object - https://github.com/dellathefella/truenas-go-sdk/blob/main/docs/GroupApi.md#updategroup
	input := expandGroup(d)

	_, _, err = c.GroupApi.UpdateGroup(ctx, int32(id)).CreateGroupParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error updating group: %s\n%s", err, body)
	}

	return resourceTrueNASGroupRead(ctx, d, m)
}

func expandGroup(d *schema.ResourceData) api.CreateGroupParams {
	input := api.CreateGroupParams{
		Name: d.Get("name").(string),
	}

	if d.HasChange("allow_duplicate_gid") {
		input.AllowDuplicateGid = getBoolPtr(d.Get("allow_duplicate_gid").(bool))
	}

	if d.HasChange("gid") {
		input.Gid = getInt32Ptr(int32(d.Get("gid").(int)))
	}

	if d.HasChange("smb") {
		input.Smb = getBoolPtr(d.Get("smb").(bool))
	}

	if d.HasChange("sudo") {
		input.Sudo = getBoolPtr(d.Get("sudo").(bool))
	}

	if d.HasChange("sudo_no_password") {
		input.SudoNopasswd = getBoolPtr(d.Get("sudo_no_password").(bool))
	}

	if sudoCommands, ok := d.GetOk("sudo_commands"); ok {
		// Convert set of strings to []string
		tfStrings := sudoCommands.(*schema.Set).List()
		commandStrings := make([]string, len(tfStrings))
		for i := range commandStrings {
			commandStrings[i] = tfStrings[i].(string)
		}
		input.SudoCommands = commandStrings
	}

	return input
}
