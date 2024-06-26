package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func dataSourceTrueNASShareSMB() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about specific SMB share",
		ReadContext: dataSourceTrueNASShareSMBRead,
		Schema: map[string]*schema.Schema{
			"sharesmb_id": &schema.Schema{
				Description: "SMB Share ID",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"path": &schema.Schema{
				Description: "Path to shared directory",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"path_suffix": &schema.Schema{
				Description: "Append a suffix to the share connection path. This is used to provide unique shares on a per-user, per-computer, or per-IP address basis.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"purpose": &schema.Schema{
				Description: "You can set a share purpose to apply and lock pre-determined advanced options for the share.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"comment": &schema.Schema{
				Description: "Any notes about this SMB share",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"hostsallow": &schema.Schema{
				Description: "Authorized hosts (IP/hostname)",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"hostsdeny": &schema.Schema{
				Description: "Disallowed hosts (IP/hostname). Pass 'ALL' to use whitelist model.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"home": &schema.Schema{
				Description: "Use as home share",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"timemachine": &schema.Schema{
				Description: "Enable TimeMachine backups to this share",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"name": &schema.Schema{
				Description: "SMB share name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ro": &schema.Schema{
				Description: "Prohibit writing",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"browsable": &schema.Schema{
				Description: "Browsable to network clients",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"recyclebin": &schema.Schema{
				Description: "Export recycle bin",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"shadowcopy": &schema.Schema{
				Description: "Export ZFS snapshots as Shadow Copies for Microsoft Volume Shadow Copy Service (VSS) clients",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"guestok": &schema.Schema{
				Description: "Allow access to this share without a password",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"aapl_name_mangling": &schema.Schema{
				Description: "Use Apple-style Character Encoding",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"abe": &schema.Schema{
				Description: "Access based share enumeration ",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"acl": &schema.Schema{
				Description: "Enable support for storing the SMB Security Descriptor as a Filesystem ACL",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"durablehandle": &schema.Schema{
				Description: "Enable SMB2/3 Durable Handles: Allow using open file handles that can withstand short disconnections",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"fsrvp": &schema.Schema{
				Description: "Enable support for the File Server Remote VSS Protocol.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"streams": &schema.Schema{
				Description: "Enable Alternate Data Streams: Allow multiple NTFS data streams. Disabling this option causes macOS to write streams to files on the filesystem.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"vuid": &schema.Schema{
				Description: "Share VUID (set when using as TimeMachine share)",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"auxsmbconf": &schema.Schema{
				Description: "Auxiliary smb4.conf parameters",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"enabled": &schema.Schema{
				Description: "Enable this share",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"locked": &schema.Schema{
				Description: "Locking status of this share",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func dataSourceTrueNASShareSMBRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id := d.Get("sharesmb_id").(int)

	resp, _, err := c.SharingApi.GetShareSMB(ctx, int32(id)).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error getting share: %s\n%s", err, body)
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
	d.Set("locked", *resp.Locked)

	d.SetId(strconv.Itoa(int(resp.Id)))

	return diags
}
