package truenas

import (
	"context"
	"fmt"
	"github.com/dariusbakunas/terraform-provider-truenas/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const datasetType = "FILESYSTEM"

type datasetPath struct {
	Pool   string
	Parent string
	Name   string
}

var supportedCompression = []string{"off", "lz4", "gzip", "gzip-1", "gzip-9", "zstd", "zstd-fast", "zle", "lzjb", "zstd-1", "zstd-2", "zstd-3", "zstd-4", "zstd-5", "zstd-6", "zstd-7", "zstd-8", "zstd-9", "zstd-10", "zstd-11", "zstd-12", "zstd-13", "zstd-14", "zstd-15", "zstd-16", "zstd-17", "zstd-18", "zstd-19", "zstd-fast-1", "zstd-fast-2", "zstd-fast-3", "zstd-fast-4", "zstd-fast-5", "zstd-fast-6", "zstd-fast-7", "zstd-fast-8", "zstd-fast-9", "zstd-fast-10", "zstd-fast-20", "zstd-fast-30", "zstd-fast-40", "zstd-fast-50", "zstd-fast-60", "zstd-fast-70", "zstd-fast-80", "zstd-fast-90", "zstd-fast-100", "zstd-fast-500", "zstd-fast-1000"}
var encryptionAlgorithms = []string{"AES-128-CCM", "AES-192-CCM", "AES-256-CCM", "AES-128-GCM", "AES-192-GCM", "AES-256-GCM"}
var recordSizes = []string{"512", "1K", "2K", "4K", "8K", "16K", "32K", "64K", "128K", "256K", "512K", "1024K"}

// newDatasetPath creates new datasetPath struct
// from TrueNAS dataset ID string, that comes in format: Pool/Parent/dataset_name
func newDatasetPath(id string) datasetPath {
	s := strings.Split(id, "/")

	if len(s) == 2 {
		// there is no Parent
		return datasetPath{Pool: s[0], Name: s[1], Parent: ""}
	}

	// first element is Pool Name, last - dataset Name and in-between - Parent
	return datasetPath{Pool: s[0], Name: s[len(s)-1], Parent: strings.Join(s[1:len(s)-1], "/")}
}

func (d datasetPath) String() string {
	if d.Parent == "" {
		return fmt.Sprintf("%s/%s", d.Pool, d.Name)
	} else {
		return fmt.Sprintf("%s/%s/%s", d.Pool, strings.Trim(d.Parent, "/"), d.Name)
	}
}

func resourceTrueNASDataset() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTrueNASDatasetCreate,
		ReadContext:   resourceTrueNASDatasetRead,
		UpdateContext: resourceTrueNASDatasetUpdate,
		DeleteContext: resourceTrueNASDatasetDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(4 * time.Minute),
			Update: schema.DefaultTimeout(4 * time.Minute),
			Delete: schema.DefaultTimeout(4 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringDoesNotContainAny("/"),
				ForceNew:     true,
			},
			"parent": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringDoesNotContainAny("/"),
				ForceNew:     true,
			},
			"acl_mode": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "Determine how chmod behaves when adjusting file ACLs. See the zfs(8) aclmode property.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"share_type"},
				ValidateFunc:  validation.StringInSlice([]string{"passthrough", "restricted"}, false),
			},
			"acl_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"atime": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Choose 'on' to update the access time for files when they are read. Choose 'off' to prevent producing log traffic when reading files",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
			},
			"case_sensitivity": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"share_type"},
				ValidateFunc:  validation.StringInSlice([]string{"sensitive", "insensitive", "mixed"}, false),
			},
			"comments": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"compression": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(supportedCompression, false),
			},
			"copies": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 3),
			},
			"deduplication": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"on", "off", "verify"}, false),
			},
			"encrypted": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"inherit_encryption": &schema.Schema{
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
			},
			"encryption_algorithm": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(encryptionAlgorithms, false),
			},
			"pbkdf2iters": &schema.Schema{
				Type: schema.TypeInt,
				//ConflictsWith: []string{"encryption_options.key"},
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"passphrase": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"encryption_key": &schema.Schema{
				Type:          schema.TypeString,
				ConflictsWith: []string{"passphrase"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile("^[a-fA-F0-9]+$"), "key must be in hexadecimal format"),
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				Sensitive:     true,
			},
			"generate_key": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"exec": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
			},
			"managed_by": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"mount_point": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"quota_bytes": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"quota_critical": &schema.Schema{
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"quota_warning": &schema.Schema{
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"ref_quota_bytes": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"ref_quota_critical": &schema.Schema{
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"ref_quota_warning": &schema.Schema{
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"readonly": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
			},
			"record_size": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(recordSizes, false),
			},
			"share_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"generic", "smb"}, false),
			},
			"sync": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "'standard' uses the sync settings that have been requested by the client software, 'always' waits for data writes to complete, and 'disabled' never waits for writes to complete.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"standard", "always", "disabled"}, false),
			},
			"snap_dir": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"visible", "hidden"}, false),
			},
		},
	}
}

func resourceTrueNASDatasetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.Client)

	input := flattenDataset(d)

	log.Printf("[DEBUG] Creating TrueNAS dataset: %+v", input)

	resp, err := c.DatasetAPI.Create(ctx, input)

	if err != nil {
		return diag.Errorf("error creating dataset: %s", err)
	}

	d.SetId(resp.ID)

	log.Printf("[INFO] TrueNAS dataset (%s) created", resp.ID)

	return append(diags, resourceTrueNASDatasetRead(ctx, d, m)...)
}

func resourceTrueNASDatasetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.Client)

	id := d.Id()

	resp, err := c.DatasetAPI.Get(ctx, id)

	if err != nil {
		return diag.Errorf("error getting dataset: %s", err)
	}

	dpath := newDatasetPath(resp.ID)

	if err := d.Set("id", resp.ID); err != nil {
		return diag.Errorf("error setting id: %s", err)
	}

	if err := d.Set("pool", dpath.Pool); err != nil {
		return diag.Errorf("error setting pool: %s", err)
	}

	if err := d.Set("parent", dpath.Parent); err != nil {
		return diag.Errorf("error setting parent: %s", err)
	}

	if err := d.Set("name", dpath.Name); err != nil {
		return diag.Errorf("error setting name: %s", err)
	}

	if err := d.Set("mount_point", resp.MountPoint); err != nil {
		return diag.Errorf("error setting mount_point: %s", err)
	}

	if resp.ACLMode != nil {
		if err := d.Set("acl_mode", strings.ToLower(*resp.ACLMode.Value)); err != nil {
			return diag.Errorf("error setting acl_mode: %s", err)
		}
	}

	if resp.ACLType != nil {
		if err := d.Set("acl_type", strings.ToLower(*resp.ACLType.Value)); err != nil {
			return diag.Errorf("error setting acl_type: %s", err)
		}
	}

	if resp.ATime != nil {
		if err := d.Set("atime", strings.ToLower(*resp.ATime.Value)); err != nil {
			return diag.Errorf("error setting atime: %s", err)
		}
	}

	if resp.CaseSensitivity != nil {
		if err := d.Set("case_sensitivity", strings.ToLower(*resp.CaseSensitivity.Value)); err != nil {
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

	if resp.ManagedBy != nil {
		if err := d.Set("managed_by", resp.ManagedBy.Value); err != nil {
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
		quota, err := strconv.Atoi(resp.Quota.RawValue)

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

	if resp.RefQuota != nil {
		quota, err := strconv.Atoi(resp.RefQuota.RawValue)

		if err != nil {
			return diag.Errorf("error parsing refquota: %s", err)
		}

		if err := d.Set("ref_quota_bytes", quota); err != nil {
			return diag.Errorf("error setting ref_quota_bytes: %s", err)
		}
	}

	if resp.RefQuotaCritical != nil {
		quota, err := strconv.Atoi(*resp.RefQuotaCritical.Value)

		if err != nil {
			return diag.Errorf("error parsing refquota_critical: %s", err)
		}

		if err := d.Set("ref_quota_critical", quota); err != nil {
			return diag.Errorf("error setting ref_quota_critical: %s", err)
		}
	}

	if resp.RefQuotaWarning != nil {
		quota, err := strconv.Atoi(*resp.RefQuotaWarning.Value)

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
	}

	if resp.ShareType != nil {
		if err := d.Set("share_type", strings.ToLower(*resp.ShareType.Value)); err != nil {
			return diag.Errorf("error setting share_type: %s", err)
		}
	}

	if resp.Sync != nil {
		if err := d.Set("sync", strings.ToLower(*resp.Sync.Value)); err != nil {
			return diag.Errorf("error setting sync: %s", err)
		}
	}

	if resp.SnapDir != nil {
		if err := d.Set("snap_dir", strings.ToLower(*resp.SnapDir.Value)); err != nil {
			return diag.Errorf("error setting snap_dir: %s", err)
		}
	}

	if resp.EncryptionAlgorithm != nil && resp.EncryptionAlgorithm.Value != nil {
		if err := d.Set("encryption_algorithm", *resp.EncryptionAlgorithm.Value); err != nil {
			return diag.Errorf("error setting encryption_algorithm: %s", err)
		}
	}

	if resp.PBKDF2Iters != nil {
		iters, err := strconv.Atoi(*resp.PBKDF2Iters.Value)

		if err != nil {
			return diag.Errorf("error parsing PBKDF2Iters: %s", err)
		}

		if iters > 0 {
			if err := d.Set("pbkdf2iters", iters); err != nil {
				return diag.Errorf("error setting PBKDF2Iters: %s", err)
			}
		}
	}

	if err := d.Set("encrypted", resp.Encrypted); err != nil {
		return diag.Errorf("error setting encrypted: %s", err)
	}

	return diags
}

func resourceTrueNASDatasetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.Client)

	input := flattenDatasetForUpdate(d)

	log.Printf("[DEBUG] Updating TrueNAS dataset: %+v", input)

	err := c.DatasetAPI.Update(ctx, d.Id(), input)

	if err != nil {
		return diag.Errorf("error creating dataset: %s", err)
	}

	log.Printf("[INFO] TrueNAS dataset (%s) updated", d.Id())

	return append(diags, resourceTrueNASDatasetRead(ctx, d, m)...)
}

func resourceTrueNASDatasetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.Client)
	id := d.Id()

	log.Printf("[DEBUG] Deleting TrueNAS dataset: %s", id)

	err := c.DatasetAPI.Delete(ctx, id)

	if err != nil {
		return diag.Errorf("error deleting dataset: %s", err)
	}

	log.Printf("[INFO] TrueNAS dataset (%s) deleted", id)

	return diags
}

func flattenDataset(d *schema.ResourceData) *api.CreateDatasetInput {
	p := datasetPath{
		Pool:   d.Get("pool").(string),
		Parent: d.Get("parent").(string),
		Name:   d.Get("name").(string),
	}

	input := &api.CreateDatasetInput{
		Name: p.String(),
	}

	if sync, ok := d.GetOk("sync"); ok {
		input.Sync = strings.ToUpper(sync.(string))
	}

	if caseSensitivity, ok := d.GetOk("case_sensitivity"); ok {
		input.CaseSensitivity = strings.ToUpper(caseSensitivity.(string))
	}

	if comments, ok := d.GetOk("comments"); ok {
		input.Comments = comments.(string)
	}

	if compression, ok := d.GetOk("compression"); ok {
		input.Compression = strings.ToUpper(compression.(string))
	}

	if deduplication, ok := d.GetOk("deduplication"); ok {
		input.Deduplication = strings.ToUpper(deduplication.(string))
	}

	if copies, ok := d.GetOk("copies"); ok {
		input.Copies = copies.(int)
	}

	if exec, ok := d.GetOk("exec"); ok {
		input.Exec = strings.ToUpper(exec.(string))
	}

	if aclmode, ok := d.GetOk("acl_mode"); ok {
		input.ACLMode = strings.ToUpper(aclmode.(string))
	}

	if atime, ok := d.GetOk("atime"); ok {
		input.ATime = strings.ToUpper(atime.(string))
	}

	if quota, ok := d.GetOk("quota_bytes"); ok {
		input.Quota = quota.(int)
	}

	if quotaCritical, ok := d.GetOk("quota_critical"); ok {
		input.QuotaCritical = getIntPtr(quotaCritical.(int))
	}

	if quotaWarning, ok := d.GetOk("quota_warning"); ok {
		input.QuotaWarning = getIntPtr(quotaWarning.(int))
	}

	if refQuota, ok := d.GetOk("ref_quota_bytes"); ok {
		input.RefQuota = refQuota.(int)
	}

	if refQuotaCritical, ok := d.GetOk("ref_quota_critical"); ok {
		input.RefQuotaCritical = getIntPtr(refQuotaCritical.(int))
	}

	if refQuotaWarning, ok := d.GetOk("ref_quota_warning"); ok {
		input.RefQuotaWarning = getIntPtr(refQuotaWarning.(int))
	}

	if readonly, ok := d.GetOk("readonly"); ok {
		input.Readonly = strings.ToUpper(readonly.(string))
	}

	if recordSize, ok := d.GetOk("record_size"); ok {
		input.RecordSize = strings.ToUpper(recordSize.(string))
	}

	if shareType, ok := d.GetOk("share_type"); ok {
		input.ShareType = strings.ToUpper(shareType.(string))
	}

	if snapDir, ok := d.GetOk("snap_dir"); ok {
		input.SnapDir = strings.ToUpper(snapDir.(string))
	}

	encrypted := d.Get("encrypted")

	if encrypted != nil {
		input.Encrypted = getBoolPtr(encrypted.(bool))
	}

	inheritEncryption := d.Get("inherit_encryption")

	if inheritEncryption != nil {
		input.InheritEncryption = getBoolPtr(inheritEncryption.(bool))
	}

	encOptions := &api.EncryptionOptions{}

	if algorithm, ok := d.GetOk("encryption_algorithm"); ok {
		encOptions.Algorithm = algorithm.(string)
	}

	if genKey, ok := d.GetOk("generate_key"); ok {
		encOptions.GenerateKey = getBoolPtr(genKey.(bool))
	}

	if passphrase, ok := d.GetOk("passphrase"); ok {
		encOptions.Passphrase = passphrase.(string)
	}

	if key, ok := d.GetOk("encryption_key"); ok {
		encOptions.Key = key.(string)
	}

	if (api.EncryptionOptions{}) != *encOptions {
		input.EncryptionOptions = encOptions
	}

	input.Type = datasetType
	return input
}

func flattenDatasetForUpdate(d *schema.ResourceData) *api.UpdateDatasetInput {
	input := &api.UpdateDatasetInput{}

	if sync, ok := d.GetOk("sync"); ok {
		input.Sync = strings.ToUpper(sync.(string))
	}

	if comments, ok := d.GetOk("comments"); ok {
		input.Comments = comments.(string)
	}

	if compression, ok := d.GetOk("compression"); ok {
		input.Compression = strings.ToUpper(compression.(string))
	}

	if deduplication, ok := d.GetOk("deduplication"); ok {
		input.Deduplication = strings.ToUpper(deduplication.(string))
	}

	if copies, ok := d.GetOk("copies"); ok {
		input.Copies = copies.(int)
	}

	if exec, ok := d.GetOk("exec"); ok {
		input.Exec = strings.ToUpper(exec.(string))
	}

	if aclmode, ok := d.GetOk("acl_mode"); ok {
		input.ACLMode = strings.ToUpper(aclmode.(string))
	}

	if atime, ok := d.GetOk("atime"); ok {
		input.ATime = strings.ToUpper(atime.(string))
	}

	if quota, ok := d.GetOk("quota_bytes"); ok {
		input.Quota = quota.(int)
	}

	if refQuota, ok := d.GetOk("ref_quota_bytes"); ok {
		input.RefQuota = refQuota.(int)
	}

	if readonly, ok := d.GetOk("readonly"); ok {
		input.Readonly = strings.ToUpper(readonly.(string))
	}

	if recordSize, ok := d.GetOk("record_size"); ok {
		input.RecordSize = strings.ToUpper(recordSize.(string))
	}

	if snapDir, ok := d.GetOk("snap_dir"); ok {
		input.SnapDir = strings.ToUpper(snapDir.(string))
	}

	return input
}
