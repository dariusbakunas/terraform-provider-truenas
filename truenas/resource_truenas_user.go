package truenas

import (
	"context"
	"encoding/json"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strconv"
)

func resourceTrueNASUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates a user",
		CreateContext: resourceTrueNASUserCreate,
		ReadContext:   resourceTrueNASUserRead,
		UpdateContext: resourceTrueNASUserUpdate,
		DeleteContext: resourceTrueNASUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"create_group": &schema.Schema{
				Description: "Create a new primary group for this user.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"email": &schema.Schema{
				Description: "User email address.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"full_name": &schema.Schema{
				Description: "User full name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"gid": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Primary Group ID. If not provided, it is automatically filled with the next one available.",
				Optional:    true,
				Computed:    true,
			},
			"group_ids": &schema.Schema{
				Description: "List of Group IDs the user belongs to.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"home_directory": &schema.Schema{
				Description: "Home directory path.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"home_mode": &schema.Schema{
				Description: "Home directory mode.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"locked": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Specifies whether the account is locked.",
				Optional:    true,
			},
			"microsoft_account": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Specifies whether the account is used for Microsoft authentication.",
				Optional:    true,
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Specifies the account password. Cannot be read (one-way hashed on the server.)",
				Optional:    true,
				Sensitive:   true,
			},
			"password_disabled": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Specifies whether the account can only use SSH keys for authentication.",
				Optional:    true,
			},
			"name": &schema.Schema{
				Description: "Specifies the username for the account.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"shell": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Specifies the shell executable for the user.",
				Optional:    true,
				Default:     "/bin/sh",
			},
			"smb": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Specifies whether the user should be mapped into an NT login account.",
				Optional:    true,
			},
			"ssh_public_key": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Specifies an initial SSH public key to establish on the account.",
				Optional:    true,
			},
			"sudo": &schema.Schema{
				Description: "Specifies whether the user may invoke sudo.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"sudo_commands": &schema.Schema{
				Description: "Specifies a list of executables the user may invoke via sudo without a password.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"sudo_no_password": &schema.Schema{
				Description: "Specifies whether the user may invoke sudo without a password.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"uid": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "User ID. If not provided, it is automatically filled with the next one available.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceTrueNASUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resp, http, err := c.UserApi.GetUser(ctx, int32(id)).Execute()

	if err != nil {
		// gracefully handle manual deletions
		d.SetId("")
		if http.StatusCode == 404 {
			return nil
		}
		return diag.Errorf("error getting user: %s", err)
	}

	d.Set("name", resp.Username)
	d.Set("full_name", resp.FullName)

	if resp.Email.IsSet() {
		d.Set("email", resp.Email.Get())
	}

	if resp.Group != nil {
		d.Set("gid", (*resp.Group).BsdgrpGid)
	}

	// The API returns TrueNAS group IDs, not BSD group IDs - we have to translate it
	if resp.Groups != nil {
		groupsResp := resp.Groups
		groups := make([]int, len(groupsResp))
		for i := range groupsResp {
			groupResp, _, err := c.GroupApi.GetGroup(ctx, int32(groupsResp[i])).Execute()
			if err != nil {
				return diag.Errorf("error getting group %s for user %s: %s", string(groupsResp[i]), d.Id(), err)
			}
			groups[i] = int(*groupResp.Gid)
		}
		d.Set("group_ids", groups)
	}

	if resp.Home != nil {
		d.Set("home", *resp.Home)
	}

	if resp.Locked != nil {
		d.Set("locked", *resp.Locked)
	}

	if resp.MicrosoftAccount != nil {
		d.Set("microsoft_account", *resp.MicrosoftAccount)
	}

	if resp.PasswordDisabled != nil {
		d.Set("password_disabled", *resp.PasswordDisabled)
	}

	if resp.Shell != nil {
		d.Set("shell", *resp.Shell)
	}

	if resp.Smb != nil {
		d.Set("smb", *resp.Smb)
	}

	if resp.Sshpubkey.IsSet() {
		d.Set("ssh_public_key", resp.Sshpubkey.Get())
	}

	if resp.SudoCommands != nil {
		d.Set("sudo_commands", resp.SudoCommands)
	}

	if resp.SudoNopasswd != nil {
		d.Set("sudo_no_password", *resp.SudoNopasswd)
	}

	if resp.Sudo != nil {
		d.Set("sudo", *resp.Sudo)
	}

	if resp.Uid != nil {
		d.Set("uid", *resp.Uid)
	}

	return diags
}

func resourceTrueNASUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	input, diags := expandUserCreate(ctx, d, m)
	if diags != nil {
		return *diags
	}

	resp, _, err := c.UserApi.CreateUser(ctx).CreateUserParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error creating user: %s\n%s", err, body)
	}

	d.SetId(strconv.Itoa(int(resp)))

	return resourceTrueNASUserRead(ctx, d, m)
}

func resourceTrueNASUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting TrueNAS user: %s", strconv.Itoa(id))
	input := api.DeleteUserParams{
		DeleteGroup: getBoolPtr(true),
	}
	_, err = c.UserApi.DeleteUser(ctx, int32(id)).DeleteUserParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error deleting user: %s\n%s", err, body)
	}

	log.Printf("[INFO] TrueNAS user (%s) deleted", d.Id())
	d.SetId("")

	return diags
}

func resourceTrueNASUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input, diags := expandUserUpdate(ctx, d, m)
	if diags != nil {
		return *diags
	}
	_, _, err = c.UserApi.UpdateUser(ctx, int32(id)).UpdateUserParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error updating user: %s\n%s", err, body)
	}

	return resourceTrueNASUserRead(ctx, d, m)
}

type APIGroup struct {
	Id           int
	Gid          int
	Group        string
	Builtin      bool
	Sudo         bool
	SudoNopasswd bool
	SudoCommands []string
	Smb          bool
	Users        []int
	Local        bool
	IdTypeBoth   bool
}

func enumerateGroups(ctx context.Context, m interface{}) ([]APIGroup, *diag.Diagnostics) {
	c := m.(*api.APIClient)

	// It seems the only way to map gid to group.id is to enumerate the whole list :(
	groupsResponse, err := c.GroupApi.ListGroups(ctx).Execute()
	if err != nil {
		diags := diag.Errorf("error getting groups: %s", err)
		return []APIGroup{}, &diags
	}

	var groupsList []APIGroup
	decoder := json.NewDecoder(groupsResponse.Body)
	err = decoder.Decode(&groupsList)
	if err != nil {
		diags := diag.Errorf("error decoding groups response: %s", err)
		return []APIGroup{}, &diags
	}

	return groupsList, nil
}

func expandUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) (api.CreateUserParams, *diag.Diagnostics) {
	apiGroups, diags := enumerateGroups(ctx, m)
	if diags != nil {
		return api.CreateUserParams{}, diags
	}

	input := api.CreateUserParams{
		Username: d.Get("name").(string),
		FullName: d.Get("full_name").(string),
	}

	if uid, ok := d.GetOk("uid"); ok {
		input.Uid = getInt32Ptr(int32(uid.(int)))
	}

	if group, ok := d.GetOk("gid"); ok {
		groupLookup := group.(int)
		// TrueNAS group ID, not bsd group ID
		var foundGroup *APIGroup = nil
		for _, x := range apiGroups {
			if x.Gid == groupLookup {
				foundGroup = &x
				break
			}
		}
		if foundGroup == nil {
			diags := diag.Errorf("error getting group %s: not found", strconv.Itoa(groupLookup))
			return api.CreateUserParams{}, &diags
		}
		input.Group = getInt32Ptr(int32(foundGroup.Id))
	}

	if createGroup, ok := d.GetOk("create_group"); ok {
		input.GroupCreate = getBoolPtr(createGroup.(bool))
	}

	if homeDirectory, ok := d.GetOk("home_directory"); ok {
		input.Home = getStringPtr(homeDirectory.(string))
	}

	if homeMode, ok := d.GetOk("home_directory_mode"); ok {
		input.HomeMode = getStringPtr(homeMode.(string))
	}

	if shell, ok := d.GetOk("shell"); ok {
		input.Shell = getStringPtr(shell.(string))
	}

	if email, ok := d.GetOk("email"); ok {
		input.Email.Set(getStringPtr(email.(string)))
	}

	if password, ok := d.GetOk("password"); ok {
		input.Password = getStringPtr(password.(string))
	}

	if passwordDisabled, ok := d.GetOk("password_disabled"); ok {
		input.PasswordDisabled = getBoolPtr(passwordDisabled.(bool))
	}

	if locked, ok := d.GetOk("locked"); ok {
		input.Locked = getBoolPtr(locked.(bool))
	}

	if microsoftAccount, ok := d.GetOk("microsoft_account"); ok {
		input.MicrosoftAccount = getBoolPtr(microsoftAccount.(bool))
	}

	if smb, ok := d.GetOk("smb"); ok {
		input.Smb = getBoolPtr(smb.(bool))
	}

	if sudo, ok := d.GetOk("sudo"); ok {
		input.Sudo = getBoolPtr(sudo.(bool))
	}

	if sudoNoPassword, ok := d.GetOk("sudo_no_password"); ok {
		input.SudoNopasswd = getBoolPtr(sudoNoPassword.(bool))
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

	if sshPublicKey, ok := d.GetOk("ssh_public_key"); ok {
		input.Sshpubkey.Set(getStringPtr(sshPublicKey.(string)))
	}

	if groupIds, ok := d.GetOk("group_ids"); ok {
		// Convert int group IDs to int32 for CreateUserParams model, mapping unix GID to TrueNAS middleware ID
		intGroupIds := groupIds.(*schema.Set).List()
		int32GroupIds := make([]int32, len(intGroupIds))
		for i := range intGroupIds {
			// TrueNAS group ID, not bsd group ID
			groupLookup := intGroupIds[i].(int)
			var foundGroup *APIGroup = nil
			for _, x := range apiGroups {
				if x.Gid == groupLookup {
					foundGroup = &x
					break
				}
			}
			if foundGroup == nil {
				diags := diag.Errorf("error getting group %s: not found", strconv.Itoa(groupLookup))
				return api.CreateUserParams{}, &diags
			}
			int32GroupIds[i] = int32(foundGroup.Id)
		}
		input.Groups = int32GroupIds
	}

	return input, nil
}

func expandUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) (api.UpdateUserParams, *diag.Diagnostics) {
	apiGroups, diags := enumerateGroups(ctx, m)
	if diags != nil {
		return api.UpdateUserParams{}, diags
	}
	input := api.UpdateUserParams{}

	if d.HasChange("uid") {
		input.Uid = getInt32Ptr(int32(d.Get("uid").(int)))
	}

	if name, ok := d.GetOk("name"); ok {
		input.Username = getStringPtr(name.(string))
	}

	if d.HasChange("gid") {
		groupLookup := d.Get("gid").(int)
		// TrueNAS group ID, not bsd group ID
		var foundGroup *APIGroup = nil
		for _, x := range apiGroups {
			if x.Gid == groupLookup {
				foundGroup = &x
				break
			}
		}
		if foundGroup == nil {
			diags := diag.Errorf("error getting group %s: not found", strconv.Itoa(groupLookup))
			return api.UpdateUserParams{}, &diags
		}
		input.Group = getInt32Ptr(int32(foundGroup.Id))
	}

	if homeDirectory, ok := d.GetOk("home_directory"); ok {
		input.Home = getStringPtr(homeDirectory.(string))
	}

	if homeMode, ok := d.GetOk("home_directory_mode"); ok {
		input.HomeMode = getStringPtr(homeMode.(string))
	}

	if shell, ok := d.GetOk("shell"); ok {
		input.Shell = getStringPtr(shell.(string))
	}
	if fullName, ok := d.GetOk("full_name"); ok {
		input.FullName = getStringPtr(fullName.(string))
	}

	if email, ok := d.GetOk("email"); ok {
		input.Email.Set(getStringPtr(email.(string)))
	}

	if password, ok := d.GetOk("password"); ok {
		input.Password = getStringPtr(password.(string))
	}

	if d.HasChange("password_disabled") {
		input.PasswordDisabled = getBoolPtr(d.Get("password_disabled").(bool))
	}

	if d.HasChange("locked") {
		input.Locked = getBoolPtr(d.Get("locked").(bool))
	}

	if d.HasChange("microsoft_account") {
		input.MicrosoftAccount = getBoolPtr(d.Get("microsoft_account").(bool))
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

	if sshPublicKey, ok := d.GetOk("ssh_public_key"); ok {
		input.Sshpubkey.Set(getStringPtr(sshPublicKey.(string)))
	}

	if d.HasChange("group_ids") {
		// Convert int group IDs to int32 for UpdateUserParams model, mapping unix GID to TrueNAS middleware ID
		intGroupIds := d.Get("group_ids").(*schema.Set).List()
		int32GroupIds := make([]int32, len(intGroupIds))
		for i := range intGroupIds {
			// TrueNAS group ID, not bsd group ID
			groupLookup := intGroupIds[i].(int)
			var foundGroup *APIGroup = nil
			for _, x := range apiGroups {
				if x.Gid == groupLookup {
					foundGroup = &x
					break
				}
			}
			if foundGroup == nil {
				diags := diag.Errorf("error getting group %s: not found", strconv.Itoa(groupLookup))
				return api.UpdateUserParams{}, &diags
			}
			int32GroupIds[i] = int32(foundGroup.Id)
		}
		input.Groups = int32GroupIds
	}

	return input, nil
}
