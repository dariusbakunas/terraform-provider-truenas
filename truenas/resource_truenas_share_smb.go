package truenas

import (
	"context"
	"errors"
	"fmt"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strconv"
)

func resourceTrueNASShareSMB() *schema.Resource {
	return &schema.Resource{
		Description:   "SMB (also known as CIFS) is the native file sharing system in Windows. SMB shares can connect to any major operating system, including Windows, MacOS, and Linux. SMB can be used in TrueNAS to share files among single or multiple users or devices.",
		CreateContext: resourceTrueNASShareSMBCreate,
		ReadContext:   resourceTrueNASShareSMBRead,
		UpdateContext: resourceTrueNASShareSMBUpdate,
		DeleteContext: resourceTrueNASShareSMBDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"sharesmb_id": &schema.Schema{
				Description: "SMB Share ID",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"path": &schema.Schema{
				Description: "Path to shared directory",
				Type:        schema.TypeString,
				Required:    true,
			},
			"path_suffix": &schema.Schema{
				Description: "Append a suffix to the share connection path. This is used to provide unique shares on a per-user, per-computer, or per-IP address basis.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"purpose": &schema.Schema{
				Description:  "You can set a share purpose to apply and lock pre-determined advanced options for the share.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "NO_PRESET",
				ValidateFunc: validation.StringInSlice([]string{"NO_PRESET", "DEFAULT_SHARE", "ENHANCED_TIMEMACHINE", "MULTI_PROTOCOL_NFS", "PRIVATE_DATASETS", "WORM_DROPBOX"}, false),
			},
			"comment": &schema.Schema{
				Description: "Any notes about this SMB share",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"hostsallow": &schema.Schema{
				Description: "Authorized hosts (IP/hostname)",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"hostsdeny": &schema.Schema{
				Description: "Disallowed hosts (IP/hostname). Pass 'ALL' to use whitelist model.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"home": &schema.Schema{
				Description: "Use as home share",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"timemachine": &schema.Schema{
				Description: "Enable TimeMachine backups to this share",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"name": &schema.Schema{
				Description: "SMB share name. Defaults to last part of shared path.",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 80),
					validation.StringDoesNotContainAny(`"\/[]:|<>+=;,*?`),
					validation.StringNotInSlice([]string{"global", "homes", "printers"}, false),
				),
			},
			"ro": &schema.Schema{
				Description: "Prohibit writing",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"browsable": &schema.Schema{
				Description: "Browsable to network clients",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"recyclebin": &schema.Schema{
				Description: "Export recycle bin",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"shadowcopy": &schema.Schema{
				Description: "Export ZFS snapshots as Shadow Copies for Microsoft Volume Shadow Copy Service (VSS) clients",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"guestok": &schema.Schema{
				Description: "Allow access to this share without a password",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"aapl_name_mangling": &schema.Schema{
				Description: "Use Apple-style Character Encoding",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"abe": &schema.Schema{
				Description: "Access based share enumeration ",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"acl": &schema.Schema{
				Description: "Enable support for storing the SMB Security Descriptor as a Filesystem ACL",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"durablehandle": &schema.Schema{
				Description: "Enable SMB2/3 Durable Handles: Allow using open file handles that can withstand short disconnections",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"fsrvp": &schema.Schema{
				Description: "Enable support for the File Server Remote VSS Protocol.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"streams": &schema.Schema{
				Description: "Enable Alternate Data Streams: Allow multiple NTFS data streams. Disabling this option causes macOS to write streams to files on the filesystem.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"vuid": &schema.Schema{
				Description: "Share VUID (set when using as TimeMachine share)",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"auxsmbconf": &schema.Schema{
				Description: "Auxiliary smb4.conf parameters",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": &schema.Schema{
				Description: "Enable this share",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

func resourceTrueNASShareSMBRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, http, err := c.SharingApi.GetShareSMB(ctx, int32(id)).Execute()

	if err != nil {
		// gracefully handle manual deletions
		d.SetId("")
		if http.StatusCode == 404 {
			return nil
		}
		return diag.Errorf("error getting smb share: %s", err)
	}

	d.Set("path", resp.Path)

	if resp.PathSuffix != nil {
		d.Set("path_suffix", *resp.PathSuffix)
	}

	if resp.Purpose != nil {
		d.Set("purpose", *resp.Purpose)
	}

	if resp.Comment != nil {
		d.Set("comment", *resp.Comment)
	}

	if resp.Hostsallow != nil {
		if err := d.Set("hostsallow", flattenStringList(resp.Hostsallow)); err != nil {
			return diag.Errorf("error setting hostsallow: %s", err)
		}
	}

	if resp.Hostsdeny != nil {
		if err := d.Set("hostsdeny", flattenStringList(resp.Hostsdeny)); err != nil {
			return diag.Errorf("error setting hostsdeny: %s", err)
		}
	}

	d.Set("home", *resp.Home)
	d.Set("timemachine", *resp.Timemachine)

	if resp.Name != nil {
		d.Set("name", *resp.Name)
	}

	d.Set("ro", *resp.Ro)
	d.Set("browsable", *resp.Browsable)
	d.Set("recyclebin", *resp.Recyclebin)
	d.Set("shadowcopy", *resp.Shadowcopy)
	d.Set("guestok", *resp.Guestok)
	d.Set("aapl_name_mangling", *resp.AaplNameMangling)
	d.Set("abe", *resp.Abe)
	d.Set("acl", *resp.Acl)
	d.Set("durablehandle", *resp.Durablehandle)
	d.Set("streams", *resp.Streams)
	d.Set("fsrvp", *resp.Fsrvp)

	if resp.Vuid != nil {
		d.Set("vuid", *resp.Vuid)
	}

	if resp.Auxsmbconf != nil {
		d.Set("auxsmbconf", *resp.Auxsmbconf)
	}

	d.Set("enabled", *resp.Enabled)

	d.Set("sharesmb_id", int(resp.Id))

	return diags
}

func resourceTrueNASShareSMBCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	input, err := expandShareSMB(d)

	resp, _, err := c.SharingApi.CreateShareSMB(ctx).CreateShareSMBParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error creating SMB share: %s\n%s", err, body)
	}

	d.SetId(strconv.Itoa(int(resp.Id)))

	return resourceTrueNASShareSMBRead(ctx, d, m)
}

func resourceTrueNASShareSMBDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting TrueNAS SMB share: %s", strconv.Itoa(id))

	_, err = c.SharingApi.RemoveShareSMB(ctx, int32(id)).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error deleting SMB share: %s\n%s", err, body)
	}

	log.Printf("[INFO] TrueNAS SMB share (%s) deleted", strconv.Itoa(id))
	d.SetId("")

	return diags
}

func resourceTrueNASShareSMBUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)
	share, err := expandShareSMB(d)

	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, _, err = c.SharingApi.UpdateShareSMB(ctx, int32(id)).CreateShareSMBParams(share).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error updating SMB share: %s\n%s", err, body)
	}

	return resourceTrueNASShareSMBRead(ctx, d, m)
}

func expandShareSMB(d *schema.ResourceData) (api.CreateShareSMBParams, error) {
	share := api.CreateShareSMBParams{
		Path: d.Get("path").(string),
	}

	// first check for purpose since that might lock other parameters

	if purpose, ok := d.GetOk("purpose"); ok {
		share.Purpose = getStringPtr(purpose.(string))
	}

	if name, ok := d.GetOk("name"); ok {
		err := isParamLocked("Name", &share)
		if err != nil {
			return share, err
		}
		share.Name = getStringPtr(name.(string))
	}

	if path_suffix, ok := d.GetOk("path_suffix"); ok {
		err := isParamLocked("PathSuffix", &share)
		if err != nil {
			return share, err
		}
		share.PathSuffix = getStringPtr(path_suffix.(string))
	}

	if comment, ok := d.GetOk("comment"); ok {
		err := isParamLocked("Comment", &share)
		if err != nil {
			return share, err
		}
		share.Comment = getStringPtr(comment.(string))
	}

	if hostsallow, ok := d.GetOk("hostsallow"); ok {
		err := isParamLocked("Hostsallow", &share)
		if err != nil {
			return share, err
		}
		share.Hostsallow = expandStrings(hostsallow.(*schema.Set).List())
	}

	if hostsdeny, ok := d.GetOk("hostsdeny"); ok {
		err := isParamLocked("Hostsdeny", &share)
		if err != nil {
			return share, err
		}
		share.Hostsdeny = expandStrings(hostsdeny.(*schema.Set).List())
	}

	// GetOk also checks if the value is different than empty type default (false)

	home := d.Get("home")
	if home != nil {
		err := isParamLocked("Home", &share)
		if err != nil {
			return share, err
		}
		share.Home = getBoolPtr(home.(bool))
	}

	timemachine := d.Get("timemachine")
	if timemachine != nil {
		err := isParamLocked("Timemachine", &share)
		if err != nil {
			return share, err
		}
		share.Timemachine = getBoolPtr(timemachine.(bool))
	}

	ro := d.Get("ro")
	if ro != nil {
		err := isParamLocked("Ro", &share)
		if err != nil {
			return share, err
		}
		share.Ro = getBoolPtr(ro.(bool))
	}

	browsable := d.Get("browsable")
	if browsable != nil {
		err := isParamLocked("Browsable", &share)
		if err != nil {
			return share, err
		}
		share.Browsable = getBoolPtr(browsable.(bool))
	}

	recyclebin := d.Get("recyclebin")
	if recyclebin != nil {
		err := isParamLocked("Recyclebin", &share)
		if err != nil {
			return share, err
		}
		share.Recyclebin = getBoolPtr(recyclebin.(bool))
	}

	shadowcopy := d.Get("shadowcopy")
	if shadowcopy != nil {
		err := isParamLocked("Shadowcopy", &share)
		if err != nil {
			return share, err
		}
		share.Shadowcopy = getBoolPtr(shadowcopy.(bool))
	}

	guestok := d.Get("guestok")
	if guestok != nil {
		err := isParamLocked("Guestok", &share)
		if err != nil {
			return share, err
		}
		share.Guestok = getBoolPtr(guestok.(bool))
	}

	aapl_name_mangling := d.Get("aapl_name_mangling")
	if aapl_name_mangling != nil {
		err := isParamLocked("AaplNameMangling", &share)
		if err != nil {
			return share, err
		}
		share.AaplNameMangling = getBoolPtr(aapl_name_mangling.(bool))
	}

	abe := d.Get("abe")
	if abe != nil {
		err := isParamLocked("Abe", &share)
		if err != nil {
			return share, err
		}
		share.Abe = getBoolPtr(abe.(bool))
	}

	acl := d.Get("acl")
	if acl != nil {
		err := isParamLocked("Acl", &share)
		if err != nil {
			return share, err
		}
		share.Acl = getBoolPtr(acl.(bool))
	}

	durablehandle := d.Get("durablehandle")
	if durablehandle != nil {
		err := isParamLocked("Durablehandle", &share)
		if err != nil {
			return share, err
		}
		share.Durablehandle = getBoolPtr(durablehandle.(bool))
	}

	fsrvp := d.Get("fsrvp")
	if fsrvp != nil {
		err := isParamLocked("Fsrvp", &share)
		if err != nil {
			return share, err
		}
		share.Fsrvp = getBoolPtr(fsrvp.(bool))
	}

	streams := d.Get("streams")
	if streams != nil {
		err := isParamLocked("Streams", &share)
		if err != nil {
			return share, err
		}
		share.Streams = getBoolPtr(streams.(bool))
	}

	if auxsmbconf, ok := d.GetOk("auxsmbconf"); ok {
		err := isParamLocked("Auxsmbconf", &share)
		if err != nil {
			return share, err
		}
		share.Auxsmbconf = getStringPtr(auxsmbconf.(string))
	}

	enabled := d.Get("enabled")
	if enabled != nil {
		err := isParamLocked("Enabled", &share)
		if err != nil {
			return share, err
		}
		share.Enabled = getBoolPtr(enabled.(bool))
	}

	return share, nil
}

func isParamLocked(param string, share *api.CreateShareSMBParams) error {
	// validate settings in regards to presets (some are locked by specific presets)
	// does not set value to avoid using reflect
	preset_locked_attrs := getPresetLockedAttrs()

	purpose, ok := share.GetPurposeOk()
	if ok {
		for _, locked := range preset_locked_attrs[*purpose] {
			if param == locked {
				return errors.New(fmt.Sprintf("When using preset '%s', setting '%s' is locked.", *purpose, param))
			}
		}
	}

	return nil
}

func getPresetLockedAttrs() map[string][]string {
	// https://www.truenas.com/docs/core/sharing/smb/smbshare/#creating-the-smb-share
	// timemachine is listed as unlocked for ENHANCED_TIMEMACHINE, but the GUI begs to differ
	// MULTI_PROTOCOL_AFP will be dropped

	return map[string][]string{
		"NO_PRESET":            []string{},
		"DEFAULT_SHARE":        []string{"Acl", "Ro", "Browsable", "Abe", "Hostsallow", "Hostsdeny", "Home", "Timemachine", "Shadowcopy", "Recyclebin", "AaplNameMangling", "Streams", "Durablehandle", "Fsrvp", "PathSuffix"},
		"ENHANCED_TIMEMACHINE": []string{"Timemachine", "PathSuffix"},
		"MULTI_PROTOCOL_NFS":   []string{"Acl", "Streams", "Durablehandle"},
		"PRIVATE_DATASETS":     []string{"PathSuffix"},
		"WORM_DROPBOX":         []string{"PathSuffix"},
	}
}
