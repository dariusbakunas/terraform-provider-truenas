package truenas

import (
	"context"
	"fmt"
	api "github.com/dellathefella/truenas-go-sdk"
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
		Description:   "A TrueNAS dataset is a file system that is created within a data storage pool. Datasets can contain files, directories (child datasets), and have individual permissions or flags",
		CreateContext: resourceTrueNASDatasetCreate,
		ReadContext:   resourceTrueNASDatasetRead,
		UpdateContext: resourceTrueNASDatasetUpdate,
		DeleteContext: resourceTrueNASDatasetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(4 * time.Minute),
			Update: schema.DefaultTimeout(4 * time.Minute),
			Delete: schema.DefaultTimeout(4 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"dataset_id": &schema.Schema{
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
				Description:  "Unique identifier for the dataset. Cannot be changed after the dataset is created.",
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
				Description: "Notes about the dataset.",
				Type:        schema.TypeString,
				Optional:    true,
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
				Description:  "Sets the data write synchronization. `inherit` takes the sync settings from the parent dataset, `standard` uses the settings that have been requested by the client software, `always` waits for data writes to complete, and `disabled` never waits for writes to complete.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"always", "standard", "disabled", "inherit"}, false),
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
	c := m.(*api.APIClient)

	input := expandDataset(d)

	log.Printf("[DEBUG] Creating TrueNAS dataset: %+v", input)

	resp, _, err := c.DatasetApi.CreateDataset(ctx).CreateDatasetParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error creating dataset: %s\n%s", err, body)
	}

	d.SetId(resp.Id)

	log.Printf("[INFO] TrueNAS dataset (%s) created", resp.Id)

	return resourceTrueNASDatasetRead(ctx, d, m)
}

func resourceTrueNASDatasetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)

	id := d.Id()

	resp, _, err := c.DatasetApi.GetDataset(ctx, id).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error getting dataset: %s\n%s", err, body)
	}

	dpath := newDatasetPath(resp.Id)

	d.Set("dataset_id", resp.Id)
	d.Set("pool", dpath.Pool)
	d.Set("parent", dpath.Parent)
	d.Set("name", dpath.Name)

	if resp.Mountpoint != nil {
		d.Set("mount_point", *resp.Mountpoint)
	}

	if resp.Aclmode != nil && resp.Aclmode.Value != nil {
		d.Set("acl_mode", strings.ToLower(*resp.Aclmode.Value))
	}

	if resp.Acltype != nil && resp.Acltype.Value != nil {
		d.Set("acl_type", strings.ToLower(*resp.Acltype.Value))
	}

	if resp.Atime != nil && resp.Atime.Value != nil {
		d.Set("atime", strings.ToLower(*resp.Atime.Value))
	}

	if resp.Casesensitivity != nil && resp.Casesensitivity.Value != nil {
		d.Set("case_sensitivity", strings.ToLower(*resp.Casesensitivity.Value))
	}

	if resp.Comments != nil && resp.Comments.Value != nil {
		// TrueNAS does not seem to change comments case in any way
		d.Set("comments", resp.Comments.Value)
	}

	if resp.Compression != nil && resp.Compression.Value != nil {
		d.Set("compression", strings.ToLower(*resp.Compression.Value))
	}

	if resp.Deduplication != nil && resp.Deduplication.Value != nil {
		d.Set("deduplication", strings.ToLower(*resp.Deduplication.Value))
	}

	if resp.Exec != nil && resp.Exec.Value != nil {
		d.Set("exec", strings.ToLower(*resp.Exec.Value))
	}

	if resp.Managedby != nil && resp.Managedby.Value != nil {
		d.Set("managed_by", *resp.Managedby.Value)
	}

	if resp.Copies != nil && resp.Copies.Value != nil {
		copies, err := strconv.Atoi(*resp.Copies.Value)

		if err != nil {
			return diag.Errorf("error parsing copies: %s", err)
		}

		d.Set("copies", copies)
	}

	if resp.Quota != nil {
		quota, err := strconv.Atoi(resp.Quota.Rawvalue)

		if err != nil {
			return diag.Errorf("error parsing quota: %s", err)
		}

		d.Set("quota_bytes", quota)
	}

	if resp.QuotaCritical != nil && resp.QuotaCritical.Value != nil {
		quota, err := strconv.Atoi(*resp.QuotaCritical.Value)

		if err != nil {
			return diag.Errorf("error parsing quota_critical: %s", err)
		}

		d.Set("quota_critical", quota)
	}

	if resp.QuotaWarning != nil && resp.QuotaWarning.Value != nil {
		quota, err := strconv.Atoi(*resp.QuotaWarning.Value)

		if err != nil {
			return diag.Errorf("error parsing quota_warning: %s", err)
		}

		d.Set("quota_warning", quota)
	}

	if resp.Refquota != nil {
		quota, err := strconv.Atoi(resp.Refquota.Rawvalue)

		if err != nil {
			return diag.Errorf("error parsing refquota: %s", err)
		}

		d.Set("ref_quota_bytes", quota)
	}

	if resp.RefquotaCritical != nil && resp.RefquotaCritical.Value != nil {
		quota, err := strconv.Atoi(*resp.RefquotaCritical.Value)

		if err != nil {
			return diag.Errorf("error parsing refquota_critical: %s", err)
		}

		d.Set("ref_quota_critical", quota)
	}

	if resp.RefquotaWarning != nil && resp.RefquotaWarning.Value != nil {
		quota, err := strconv.Atoi(*resp.RefquotaWarning.Value)

		if err != nil {
			return diag.Errorf("error parsing refquota_warning: %s", err)
		}

		d.Set("ref_quota_warning", quota)
	}

	if resp.Readonly != nil && resp.Readonly.Value != nil {
		d.Set("readonly", strings.ToLower(*resp.Readonly.Value))
	}

	if resp.Recordsize != nil && resp.Recordsize.Value != nil {
		d.Set("record_size", *resp.Recordsize.Value)
	}

	// TODO: doublecheck, does not seem to be ever returned
	//if resp.ShareType != nil {
	//	if err := d.Set("share_type", strings.ToLower(*resp.ShareType.Value)); err != nil {
	//		return diag.Errorf("error setting share_type: %s", err)
	//	}
	//}

	if resp.Sync != nil {
		d.Set("sync", strings.ToLower(*resp.Sync.Value))
	}

	if resp.Snapdir != nil && resp.Snapdir.Value != nil {
		d.Set("snap_dir", strings.ToLower(*resp.Snapdir.Value))
	}

	if resp.EncryptionAlgorithm != nil && resp.EncryptionAlgorithm.Value != nil {
		d.Set("encryption_algorithm", *resp.EncryptionAlgorithm.Value)
	}

	if resp.Pbkdf2iters != nil && resp.Pbkdf2iters.Value != nil {
		iters, err := strconv.Atoi(*resp.Pbkdf2iters.Value)

		if err != nil {
			return diag.Errorf("error parsing PBKDF2Iters: %s", err)
		}

		d.Set("pbkdf2iters", iters)
	}

	if resp.Encrypted != nil {
		d.Set("encrypted", *resp.Encrypted)
	}

	return diags
}

func resourceTrueNASDatasetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	input := expandDatasetForUpdate(d)

	log.Printf("[DEBUG] Updating TrueNAS dataset: %+v", input)

	_, _, err := c.DatasetApi.UpdateDataset(ctx, d.Id()).UpdateDatasetParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error updating dataset: %s\n%s", err, body)
	}

	log.Printf("[INFO] TrueNAS dataset (%s) updated", d.Id())

	return resourceTrueNASDatasetRead(ctx, d, m)
}

func resourceTrueNASDatasetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id := d.Id()

	log.Printf("[DEBUG] Deleting TrueNAS dataset: %s", id)

	_, err := c.DatasetApi.DeleteDataset(ctx, id).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error deleting dataset: %s\n%s", err, body)
	}

	log.Printf("[INFO] TrueNAS dataset (%s) deleted", id)
	d.SetId("")

	return diags
}

func expandDataset(d *schema.ResourceData) api.CreateDatasetParams {
	p := datasetPath{
		Pool:   d.Get("pool").(string),
		Parent: d.Get("parent").(string),
		Name:   d.Get("name").(string),
	}

	input := api.CreateDatasetParams{
		Name: p.String(),
	}

	if sync, ok := d.GetOk("sync"); ok {
		input.Sync = getStringPtr(strings.ToUpper(sync.(string)))
	}

	if caseSensitivity, ok := d.GetOk("case_sensitivity"); ok {
		input.Casesensitivity = getStringPtr(strings.ToUpper(caseSensitivity.(string)))
	}

	if comments, ok := d.GetOk("comments"); ok {
		input.Comments = getStringPtr(comments.(string))
	}

	if compression, ok := d.GetOk("compression"); ok {
		input.Compression = getStringPtr(strings.ToUpper(compression.(string)))
	}

	if deduplication, ok := d.GetOk("deduplication"); ok {
		input.Deduplication = getStringPtr(strings.ToUpper(deduplication.(string)))
	}

	if copies, ok := d.GetOk("copies"); ok {
		input.Copies = getInt32Ptr(int32(copies.(int)))
	}

	if exec, ok := d.GetOk("exec"); ok {
		input.Exec = getStringPtr(strings.ToUpper(exec.(string)))
	}

	if aclmode, ok := d.GetOk("acl_mode"); ok {
		input.Aclmode = getStringPtr(strings.ToUpper(aclmode.(string)))
	}

	if atime, ok := d.GetOk("atime"); ok {
		input.Atime = getStringPtr(strings.ToUpper(atime.(string)))
	}

	if quota, ok := d.GetOk("quota_bytes"); ok {
		input.Quota = getInt64Ptr(int64(quota.(int)))
	}

	if quotaCritical, ok := d.GetOk("quota_critical"); ok {
		input.QuotaCritical = getInt64Ptr(int64(quotaCritical.(int)))
	}

	if quotaWarning, ok := d.GetOk("quota_warning"); ok {
		input.QuotaWarning = getInt64Ptr(int64(quotaWarning.(int)))
	}

	if refQuota, ok := d.GetOk("ref_quota_bytes"); ok {
		input.Refquota = getInt64Ptr(int64(refQuota.(int)))
	}

	if refQuotaCritical, ok := d.GetOk("ref_quota_critical"); ok {
		input.RefquotaCritical = getInt64Ptr(int64(refQuotaCritical.(int)))
	}

	if refQuotaWarning, ok := d.GetOk("ref_quota_warning"); ok {
		input.RefquotaWarning = getInt64Ptr(int64(refQuotaWarning.(int)))
	}

	if readonly, ok := d.GetOk("readonly"); ok {
		input.Readonly = getStringPtr(strings.ToUpper(readonly.(string)))
	}

	if recordSize, ok := d.GetOk("record_size"); ok {
		input.Recordsize = getStringPtr(strings.ToUpper(recordSize.(string)))
	}

	if shareType, ok := d.GetOk("share_type"); ok {
		input.ShareType = getStringPtr(strings.ToUpper(shareType.(string)))
	}

	if snapDir, ok := d.GetOk("snap_dir"); ok {
		input.Snapdir = getStringPtr(strings.ToUpper(snapDir.(string)))
	}

	encrypted := d.Get("encrypted")

	if encrypted != nil {
		input.Encryption = getBoolPtr(encrypted.(bool))
	}

	inheritEncryption := d.Get("inherit_encryption")

	if inheritEncryption != nil {
		input.InheritEncryption = getBoolPtr(inheritEncryption.(bool))
	}

	encOptions := &api.CreateDatasetParamsEncryptionOptions{}

	if algorithm, ok := d.GetOk("encryption_algorithm"); ok {
		encOptions.Algorithm = getStringPtr(algorithm.(string))
	}

	if genKey, ok := d.GetOk("generate_key"); ok {
		encOptions.GenerateKey = getBoolPtr(genKey.(bool))
	}

	if passphrase, ok := d.GetOk("passphrase"); ok {
		encOptions.Passphrase = getStringPtr(passphrase.(string))
	}

	if key, ok := d.GetOk("encryption_key"); ok {
		encOptions.Key = getStringPtr(key.(string))
	}

	input.EncryptionOptions = encOptions

	input.Type = getStringPtr(datasetType)
	return input
}

func expandDatasetForUpdate(d *schema.ResourceData) api.UpdateDatasetParams {
	input := api.UpdateDatasetParams{}

	if sync, ok := d.GetOk("sync"); ok {
		input.Sync = getStringPtr(strings.ToUpper(sync.(string)))
	}

	if comments, ok := d.GetOk("comments"); ok {
		input.Comments = getStringPtr(comments.(string))
	}

	if compression, ok := d.GetOk("compression"); ok {
		input.Compression = getStringPtr(strings.ToUpper(compression.(string)))
	}

	if deduplication, ok := d.GetOk("deduplication"); ok {
		input.Deduplication = getStringPtr(strings.ToUpper(deduplication.(string)))
	}

	if copies, ok := d.GetOk("copies"); ok {
		input.Copies = getInt32Ptr(int32(copies.(int)))
	}

	if exec, ok := d.GetOk("exec"); ok {
		input.Exec = getStringPtr(strings.ToUpper(exec.(string)))
	}

	if aclmode, ok := d.GetOk("acl_mode"); ok {
		input.Aclmode = getStringPtr(strings.ToUpper(aclmode.(string)))
	}

	if atime, ok := d.GetOk("atime"); ok {
		input.Atime = getStringPtr(strings.ToUpper(atime.(string)))
	}

	if quota, ok := d.GetOk("quota_bytes"); ok {
		input.Quota = getInt64Ptr(int64(quota.(int)))
	}

	if refQuota, ok := d.GetOk("ref_quota_bytes"); ok {
		input.Refquota = getInt64Ptr(int64(refQuota.(int)))
	}

	if readonly, ok := d.GetOk("readonly"); ok {
		input.Readonly = getStringPtr(strings.ToUpper(readonly.(string)))
	}

	if recordSize, ok := d.GetOk("record_size"); ok {
		input.Recordsize = getStringPtr(strings.ToUpper(recordSize.(string)))
	}

	if snapDir, ok := d.GetOk("snap_dir"); ok {
		input.Snapdir = getStringPtr(strings.ToUpper(snapDir.(string)))
	}

	return input
}
