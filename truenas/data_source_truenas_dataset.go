package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"strings"
)

func dataSourceTrueNASDataset() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTrueNASDatasetRead,
		Schema: map[string]*schema.Schema{
			"dataset_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"pool": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"acl_mode": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Determine how chmod behaves when adjusting file ACLs. See the zfs(8) aclmode property.",
				Computed:    true,
			},
			"acl_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"atime": &schema.Schema{
				Type:        schema.TypeString,
				Description: "When set to 'on' access time for files when they are read is updated. Set to 'off' to prevent producing log traffic when reading files",
				Computed:    true,
			},
			"case_sensitivity": &schema.Schema{
				Description: "`sensitive` assumes filenames are case sensitive. `insensitive` assumes filenames are not case sensitive. `mixed` understands both types of filenames.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"comments": &schema.Schema{
				Description: "Any notes about this dataset.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"compression": &schema.Schema{
				Description: "Current dataset compression level",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"copies": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"deduplication": &schema.Schema{
				Description: "Deduplication can improve storage capacity, but is RAM intensive",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"encrypted": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"encryption_root": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_encryption": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"encryption_algorithm": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_loaded": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"pbkdf2iters": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			//"passphrase": &schema.Schema{
			//	Type:     schema.TypeString,
			//	Computed: true,
			//},
			//"encryption_key": &schema.Schema{
			//	Type:     schema.TypeString,
			//  Sensitive: true,
			//	Computed: true,
			//},
			//"generate_key": &schema.Schema{
			//	Type:     schema.TypeBool,
			//	Computed: true,
			//},
			"exec": &schema.Schema{
				Description: "`on` if processes can be executed from within this dataset.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"locked": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"key_format": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"managed_by": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"mount_point": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"origin": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"quota_bytes": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"quota_critical": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"quota_warning": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ref_quota_bytes": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ref_quota_critical": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ref_quota_warning": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ref_reservation": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"readonly": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"record_size": &schema.Schema{
				Description: "Matching the fixed size of data, as in a database, may result in better performance",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"record_size_bytes": &schema.Schema{
				Description: "Matching the fixed size of data, as in a database, may result in better performance",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"reservation": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"snap_dir": &schema.Schema{
				Description: ".zfs snapshot directory visibility.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"sync": &schema.Schema{
				Description: "`standard` uses the sync settings that have been requested by the client software, `always` waits for data writes to complete, and `disabled` never waits for writes to complete.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"xattr": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func updateDatasetResourceFromResponse(resp *api.Dataset, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	if resp.Mountpoint != nil {
		d.Set("mount_point", *resp.Mountpoint)
	}

	if resp.EncryptionRoot != nil {
		d.Set("encryption_root", *resp.EncryptionRoot)
	}

	if resp.KeyLoaded != nil {
		d.Set("key_loaded", *resp.KeyLoaded)
	}

	if resp.Locked != nil {
		d.Set("locked", *resp.Locked)
	}

	if resp.Aclmode != nil {
		if err := d.Set("acl_mode", strings.ToLower(*resp.Aclmode.Value)); err != nil {
			return diag.Errorf("error setting acl_mode: %s", err)
		}
	}

	if resp.Acltype != nil {
		if err := d.Set("acl_type", strings.ToLower(*resp.Acltype.Value)); err != nil {
			return diag.Errorf("error setting acl_type: %s", err)
		}
	}

	if resp.Atime != nil {
		if err := d.Set("atime", strings.ToLower(*resp.Atime.Value)); err != nil {
			return diag.Errorf("error setting atime: %s", err)
		}
	}

	if resp.Casesensitivity != nil {
		if err := d.Set("case_sensitivity", strings.ToLower(*resp.Casesensitivity.Value)); err != nil {
			return diag.Errorf("error setting case_sensitivity: %s", err)
		}
	}

	if resp.Comments != nil {
		// TrueNAS does not seem to change comments case in any way
		if err := d.Set("comments", resp.Comments.Value); err != nil {
			return diag.Errorf("error setting comments: %s", err)
		}
	}

	if resp.Compression != nil {
		if err := d.Set("compression", strings.ToLower(*resp.Compression.Value)); err != nil {
			return diag.Errorf("error setting compression: %s", err)
		}
	}

	if resp.Deduplication != nil {
		if err := d.Set("deduplication", strings.ToLower(*resp.Deduplication.Value)); err != nil {
			return diag.Errorf("error setting deduplication: %s", err)
		}
	}

	if resp.Exec != nil {
		if err := d.Set("exec", strings.ToLower(*resp.Exec.Value)); err != nil {
			return diag.Errorf("error setting exec: %s", err)
		}
	}

	if resp.KeyFormat != nil && resp.KeyFormat.Value != nil {
		if err := d.Set("key_format", strings.ToLower(*resp.KeyFormat.Value)); err != nil {
			return diag.Errorf("error setting key_format: %s", err)
		}
	}

	if resp.Managedby != nil {
		if err := d.Set("managed_by", resp.Managedby.Value); err != nil {
			return diag.Errorf("error setting managed_by: %s", err)
		}
	}

	if resp.Copies != nil {
		copies, err := strconv.Atoi(*resp.Copies.Value)

		if err != nil {
			return diag.Errorf("error parsing copies: %s", err)
		}

		if err := d.Set("copies", copies); err != nil {
			return diag.Errorf("error setting copies: %s", err)
		}
	}

	if resp.Quota != nil {
		quota, err := strconv.Atoi(resp.Quota.Rawvalue)

		if err != nil {
			return diag.Errorf("error parsing quota: %s", err)
		}

		if err := d.Set("quota_bytes", quota); err != nil {
			return diag.Errorf("error setting quota_bytes: %s", err)
		}
	}

	if resp.QuotaCritical != nil {
		quota, err := strconv.Atoi(*resp.QuotaCritical.Value)

		if err != nil {
			return diag.Errorf("error parsing quota_critical: %s", err)
		}

		if err := d.Set("quota_critical", quota); err != nil {
			return diag.Errorf("error setting quota_critical: %s", err)
		}
	}

	if resp.QuotaWarning != nil {
		quota, err := strconv.Atoi(*resp.QuotaWarning.Value)

		if err != nil {
			return diag.Errorf("error parsing quota_warning: %s", err)
		}

		if err := d.Set("quota_warning", quota); err != nil {
			return diag.Errorf("error setting quota_warning: %s", err)
		}
	}

	if resp.Reservation != nil {
		resrv, err := strconv.Atoi(resp.Reservation.Rawvalue)

		if err != nil {
			return diag.Errorf("error parsing reservation: %s", err)
		}

		if err := d.Set("reservation", resrv); err != nil {
			return diag.Errorf("error setting reservation: %s", err)
		}

	}

	if resp.Refreservation != nil {
		resrv, err := strconv.Atoi(resp.Refreservation.Rawvalue)

		if err != nil {
			return diag.Errorf("error parsing refreservation: %s", err)
		}

		if err := d.Set("ref_reservation", resrv); err != nil {
			return diag.Errorf("error setting ref_reservation: %s", err)
		}

	}

	if resp.Refquota != nil {
		quota, err := strconv.Atoi(resp.Refquota.Rawvalue)

		if err != nil {
			return diag.Errorf("error parsing refquota: %s", err)
		}

		if err := d.Set("ref_quota_bytes", quota); err != nil {
			return diag.Errorf("error setting ref_quota_bytes: %s", err)
		}
	}

	if resp.RefquotaCritical != nil {
		quota, err := strconv.Atoi(*resp.RefquotaCritical.Value)

		if err != nil {
			return diag.Errorf("error parsing refquota_critical: %s", err)
		}

		if err := d.Set("ref_quota_critical", quota); err != nil {
			return diag.Errorf("error setting ref_quota_critical: %s", err)
		}
	}

	if resp.RefquotaWarning != nil {
		quota, err := strconv.Atoi(*resp.RefquotaWarning.Value)

		if err != nil {
			return diag.Errorf("error parsing refquota_warning: %s", err)
		}

		if err := d.Set("ref_quota_warning", quota); err != nil {
			return diag.Errorf("error setting ref_quota_warning: %s", err)
		}
	}

	if resp.Readonly != nil {
		if err := d.Set("readonly", strings.ToLower(*resp.Readonly.Value)); err != nil {
			return diag.Errorf("error setting readonly: %s", err)
		}
	}

	if resp.Recordsize != nil {
		if err := d.Set("record_size", *resp.Recordsize.Value); err != nil {
			return diag.Errorf("error setting record_size: %s", err)
		}

		sz, err := strconv.Atoi(resp.Recordsize.Rawvalue)

		if err != nil {
			return diag.Errorf("error parsing recordsize rawvalue: %s", err)
		}

		if err := d.Set("record_size_bytes", sz); err != nil {
			return diag.Errorf("error setting record_size_bytes: %s", err)
		}
	}

	if resp.Sync != nil {
		if err := d.Set("sync", strings.ToLower(*resp.Sync.Value)); err != nil {
			return diag.Errorf("error setting sync: %s", err)
		}
	}

	if resp.Snapdir != nil {
		if err := d.Set("snap_dir", strings.ToLower(*resp.Snapdir.Value)); err != nil {
			return diag.Errorf("error setting snap_dir: %s", err)
		}
	}

	if resp.EncryptionAlgorithm != nil && resp.EncryptionAlgorithm.Value != nil {
		if err := d.Set("encryption_algorithm", *resp.EncryptionAlgorithm.Value); err != nil {
			return diag.Errorf("error setting encryption_algorithm: %s", err)
		}
	}

	if resp.Pbkdf2iters != nil {
		iters, err := strconv.Atoi(*resp.Pbkdf2iters.Value)

		if err != nil {
			return diag.Errorf("error parsing PBKDF2Iters: %s", err)
		}

		if iters >= 0 {
			if err := d.Set("pbkdf2iters", iters); err != nil {
				return diag.Errorf("error setting PBKDF2Iters: %s", err)
			}
		}
	}

	if resp.Origin != nil {
		if err := d.Set("origin", strings.ToLower(*resp.Origin.Value)); err != nil {
			return diag.Errorf("error setting origin: %s", err)
		}
	}

	if resp.Xattr != nil {
		if err := d.Set("xattr", strings.ToLower(*resp.Xattr.Value)); err != nil {
			return diag.Errorf("error setting xattr: %s", err)
		}
	}

	if resp.Encrypted != nil {
		d.Set("encrypted", *resp.Encrypted)
	}

	return diags
}

func dataSourceTrueNASDatasetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id := d.Get("dataset_id").(string)

	resp, _, err := c.DatasetApi.GetDataset(ctx, id).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error getting dataset: %s\n%s", err, body)
	}

	if resp.Type != "FILESYSTEM" {
		return diag.Errorf("invalid dataset, %s is %s type", id, resp.Type)
	}

	dpath := newDatasetPath(id)

	d.Set("pool", dpath.Pool)

	if dpath.Parent != "" {
		d.Set("parent", dpath.Parent)
	}

	d.Set("name", dpath.Name)

	diags = append(diags, updateDatasetResourceFromResponse(resp, d)...)

	d.SetId(resp.Id)

	return diags
}
