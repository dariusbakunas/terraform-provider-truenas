package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strconv"
	"strings"
)

func resourceTrueNASZVOL() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage ZFS Volume (zvol), use of TF `prevent_destroy` (https://www.terraform.io/docs/language/meta-arguments/lifecycle.html#prevent_destroy) flag is recommended to avoid accidental deletion",
		CreateContext: resourceTrueNASZVOLCreate,
		ReadContext:   resourceTrueNASZVOLRead,
		UpdateContext: resourceTrueNASZVOLUpdate,
		DeleteContext: resourceTrueNASZVOLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"zvol_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"blocksize": &schema.Schema{
				Description:  "Volume blocksize",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"4K", "8K", "16K", "32K", "64K", "128K"}, false),
				Default:      "32K",
			},
			"comments": &schema.Schema{
				Description: "Any notes about this volume.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"compression": &schema.Schema{
				Description:  "Compression level",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(supportedCompression, false),
			},
			"copies": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"deduplication": &schema.Schema{
				Description:  "Transparently reuse a single copy of duplicated data to save space. Deduplication can improve storage capacity, but is RAM intensive. Compressing data is generally recommended before using deduplication. Deduplicating data is a one-way process. *Deduplicated data cannot be undeduplicated!*.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "off",
				ValidateFunc: validation.StringInSlice([]string{"on", "off", "verify"}, false),
			},
			"encrypted": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"encryption_algorithm": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(encryptionAlgorithms, false),
			},
			"encryption_root": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"force_size": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "The system restricts creating a zvol that brings the pool to over 80% capacity. Set to force creation of the zvol (not recommended)",
				Optional:    true,
				Default:     false,
			},
			"key_format": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_loaded": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"locked": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": &schema.Schema{
				Description:  "Unique identifier for the volume. Cannot be changed after the zvol is created.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringDoesNotContainAny("/"),
				ForceNew:     true,
			},
			"inherit_encryption": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Use the encryption properties of the root dataset.",
				ForceNew:    true,
				Optional:    true,
				Default:     false,
			},
			"parent": &schema.Schema{
				Description: "Parent dataset",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
			},
			"pbkdf2iters": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"pool": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringDoesNotContainAny("/"),
				ForceNew:     true,
			},
			"readonly": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Set to prevent the zvol from being modified",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"on", "off", "inherit"}, false),
			},
			"ref_reservation": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"reservation": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"sync": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Sets the data write synchronization. `inherit` takes the sync settings from the parent dataset, `standard` uses the settings that have been requested by the client software, `always` waits for data writes to complete, and `disabled` never waits for writes to complete.",
				Optional:     true,
				Default:      "standard",
				ValidateFunc: validation.StringInSlice([]string{"always", "standard", "disabled", "inherit"}, false),
			},
			"volsize": &schema.Schema{
				Description: "Volume size in bytes, should be multiples of block size",
				Type:        schema.TypeInt,
				Required:    true,
			},
		},
	}
}

func resourceTrueNASZVOLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id := d.Id()

	resp, _, err := c.DatasetApi.GetDataset(ctx, id).Execute()

	if err != nil {
		return diag.Errorf("error getting zvol: %s", err)
	}

	if resp.Type != "VOLUME" {
		return diag.Errorf("invalid volume, %s is %s type", id, resp.Type)
	}

	dpath := newDatasetPath(id)

	d.Set("pool", dpath.Pool)
	d.Set("parent", dpath.Parent)
	d.Set("name", dpath.Name)

	if resp.Volblocksize != nil {
		d.Set("blocksize", *resp.Volblocksize.Value)
	}

	if resp.EncryptionRoot != nil {
		d.Set("encryption_root", resp.EncryptionRoot)
	}

	if resp.KeyLoaded != nil {
		d.Set("key_loaded", resp.KeyLoaded)
	}

	if resp.Locked != nil {
		d.Set("locked", resp.Locked)
	}

	if resp.Comments != nil {
		// TrueNAS does not seem to change comments case in any way
		d.Set("comments", resp.Comments.Value)
	}

	if resp.Compression != nil {
		d.Set("compression", strings.ToLower(*resp.Compression.Value))
	}

	if resp.Deduplication != nil {
		d.Set("deduplication", strings.ToLower(*resp.Deduplication.Value))
	}

	if resp.KeyFormat != nil && resp.KeyFormat.Value != nil {
		d.Set("key_format", strings.ToLower(*resp.KeyFormat.Value))
	}

	if resp.Copies != nil {
		copies, err := strconv.Atoi(*resp.Copies.Value)

		if err != nil {
			return diag.Errorf("error parsing copies: %s", err)
		}

		d.Set("copies", copies)
	}

	if resp.Reservation != nil {
		resrv, err := strconv.Atoi(resp.Reservation.Rawvalue)

		if err != nil {
			return diag.Errorf("error parsing reservation: %s", err)
		}

		d.Set("reservation", resrv)
	}

	if resp.Refreservation != nil {
		resrv, err := strconv.Atoi(resp.Refreservation.Rawvalue)

		if err != nil {
			return diag.Errorf("error parsing refreservation: %s", err)
		}

		d.Set("ref_reservation", resrv)
	}

	if resp.Readonly != nil {
		d.Set("readonly", strings.ToLower(*resp.Readonly.Value))
	}

	if resp.EncryptionAlgorithm != nil && resp.EncryptionAlgorithm.Value != nil {
		d.Set("encryption_algorithm", *resp.EncryptionAlgorithm.Value)
	}

	if resp.Pbkdf2iters != nil {
		iters, err := strconv.Atoi(*resp.Pbkdf2iters.Value)

		if err != nil {
			return diag.Errorf("error parsing PBKDF2Iters: %s", err)
		}

		d.Set("pbkdf2iters", iters)
	}

	if resp.Volsize != nil {
		sz, err := strconv.Atoi(resp.Volsize.Rawvalue)

		if err != nil {
			return diag.Errorf("error parsing volsize rawvalue: %s", err)
		}

		d.Set("volsize", sz)
	}

	if resp.Sync != nil {
		d.Set("sync", strings.ToLower(*resp.Sync.Value))
	}

	if resp.Encrypted != nil {
		d.Set("encrypted", *resp.Encrypted)
	}

	d.Set("zvol_id", id)

	return diags
}

func resourceTrueNASZVOLCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	input := expandZvol(d)

	resp, _, err := c.DatasetApi.CreateDataset(ctx).CreateDatasetParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error creating zvol: %s\n%s", err, body)
	}

	d.SetId(resp.Id)

	return resourceTrueNASZVOLRead(ctx, d, m)
}

func resourceTrueNASZVOLDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id := d.Id()

	log.Printf("[DEBUG] Deleting TrueNAS zvol: %s", id)

	_, err := c.DatasetApi.DeleteDataset(ctx, id).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error deleting dataset: %s\n%s", err, body)
	}

	log.Printf("[INFO] TrueNAS zvol (%s) deleted", id)
	d.SetId("")

	return diags
}

func resourceTrueNASZVOLUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	input := api.UpdateDatasetParams{}

	if d.HasChange("comments") {
		input.Comments = getStringPtr(d.Get("comments").(string))
	}

	if d.HasChange("volsize") {
		o, n := d.GetChange("volsize")
		oldSz := o.(int)
		newSz := n.(int)

		if newSz < oldSz {
			return diag.Errorf("zvol volume size can only be increased, recreate the volume to shrink")
		}

		input.Volsize = getInt64Ptr(int64(newSz))
	}

	if d.HasChange("sync") {
		input.Sync = getStringPtr(strings.ToUpper(d.Get("sync").(string)))
	}

	if d.HasChange("compression") {
		input.Compression = getStringPtr(strings.ToUpper(d.Get("compression").(string)))
	}

	if d.HasChange("deduplication") {
		input.Deduplication = getStringPtr(strings.ToUpper(d.Get("deduplication").(string)))
	}

	if d.HasChange("readonly") {
		input.Readonly = getStringPtr(strings.ToUpper(d.Get("readonly").(string)))
	}

	if d.HasChange("force_size") {
		input.ForceSize = getBoolPtr(d.Get("force_size").(bool))
	}

	_, _, err := c.DatasetApi.UpdateDataset(ctx, d.Id()).UpdateDatasetParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error updating zvol: %s\n%s", err, body)
	}

	return resourceTrueNASZVOLRead(ctx, d, m)
}

func expandZvol(d *schema.ResourceData) api.CreateDatasetParams {
	p := datasetPath{
		Pool:   d.Get("pool").(string),
		Parent: d.Get("parent").(string),
		Name:   d.Get("name").(string),
	}

	input := api.CreateDatasetParams{
		Name: p.String(),
		Type: getStringPtr("VOLUME"),
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

	if forceSize, ok := d.GetOk("force_size"); ok {
		input.ForceSize = getBoolPtr(forceSize.(bool))
	}

	if inheritEncryption, ok := d.GetOk("inherit_encryption"); ok {
		input.InheritEncryption = getBoolPtr(inheritEncryption.(bool))
	}

	if readOnly, ok := d.GetOk("readonly"); ok {
		input.Readonly = getStringPtr(strings.ToUpper(readOnly.(string)))
	}

	if sync, ok := d.GetOk("sync"); ok {
		input.Sync = getStringPtr(strings.ToUpper(sync.(string)))
	}

	if volSize, ok := d.GetOk("volsize"); ok {
		input.Volsize = getInt64Ptr(int64(volSize.(int)))
	}

	if blockSize, ok := d.GetOk("blocksize"); ok {
		input.Volblocksize = getStringPtr(blockSize.(string))
	}

	return input
}
