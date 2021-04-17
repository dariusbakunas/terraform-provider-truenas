package truenas

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strconv"
	"strings"
)

type datasetPath struct {
	Pool   string
	Parent string
	Name   string
}

var supportedCompression = []string{"off", "lz4", "gzip", "gzip-1", "gzip-9", "zstd", "zstd-fast", "zle", "lzjb", "zstd-1", "zstd-2", "zstd-3", "zstd-4", "zstd-5", "zstd-6", "zstd-7", "zstd-8", "zstd-9", "zstd-10", "zstd-11", "zstd-12", "zstd-13", "zstd-14", "zstd-15", "zstd-16", "zstd-17", "zstd-18", "zstd-19", "zstd-fast-1", "zstd-fast-2", "zstd-fast-3", "zstd-fast-4", "zstd-fast-5", "zstd-fast-6", "zstd-fast-7", "zstd-fast-8", "zstd-fast-9", "zstd-fast-10", "zstd-fast-20", "zstd-fast-30", "zstd-fast-40", "zstd-fast-50", "zstd-fast-60", "zstd-fast-70", "zstd-fast-80", "zstd-fast-90", "zstd-fast-100", "zstd-fast-500", "zstd-fast-1000"}

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
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringDoesNotContainAny("/"),
			},
			"parent": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringDoesNotContainAny("/"),
			},
			"aclmode": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Determine how chmod behaves when adjusting file ACLs. See the zfs(8) aclmode property.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"passthrough", "restricted"}, false),
			},
			"atime": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Choose 'on' to update the access time for files when they are read. Choose 'off' to prevent producing log traffic when reading files",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
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
			"copies": &schema.Schema {
				Type: schema.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.IntBetween(1,3),
			},
			"sync": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "'standard' uses the sync settings that have been requested by the client software, 'always' waits for data writes to complete, and 'disabled' never waits for writes to complete.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"standard", "always", "disabled"}, false),
			},
		},
	}
}

func resourceTrueNASDatasetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*Client)

	name := datasetPath{
		Pool:   d.Get("pool").(string),
		Parent: d.Get("parent").(string),
		Name:   d.Get("name").(string),
	}

	input := &CreateDatasetInput{
		Name: name.String(),
	}

	if sync, ok := d.GetOk("sync"); ok {
		input.Sync = strings.ToUpper(sync.(string))
	}

	if comments, ok := d.GetOk("comments"); ok {
		input.Comments = comments.(string)
	}

	if compression, ok := d.GetOk("compression"); ok {
		input.Compression = strings.ToUpper(compression.(string))
	}

	if copies, ok := d.GetOk("copies"); ok {
		input.Copies = copies.(int)
	}

	if aclmode, ok := d.GetOk("aclmode"); ok {
		input.ACLMode = strings.ToUpper(aclmode.(string))
	}

	if atime, ok := d.GetOk("atime"); ok {
		input.ATime = strings.ToUpper(atime.(string))
	}

	resp, err := c.Datasets.Create(ctx, input)

	if err != nil {
		return diag.Errorf("error creating dataset: %s", err)
	}

	d.SetId(resp.ID)

	// TODO: is this common practice? or should it just return empty diags
	return append(diags, resourceTrueNASDatasetRead(ctx, d, m)...)
}

func resourceTrueNASDatasetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*Client)

	id := d.Id()

	resp, err := c.Datasets.Get(ctx, id)

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

	if resp.ACLMode != nil {
		if err := d.Set("aclmode", strings.ToLower(*resp.ACLMode.Value)); err != nil {
			return diag.Errorf("error setting aclmode: %s", err)
		}
	}

	if resp.ATime != nil {
		if err := d.Set("atime", strings.ToLower(*resp.ATime.Value)); err != nil {
			return diag.Errorf("error setting atime: %s", err)
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

	if resp.Copies != nil {
		copies, err := strconv.Atoi(*resp.Copies.Value)

		if err != nil {
			return diag.Errorf("error parsing copies: %s", err)
		}

		if err := d.Set("copies", copies); err != nil {
			return diag.Errorf("error setting copies: %s", err)
		}
	}

	if resp.Sync != nil {
		if err := d.Set("sync", strings.ToLower(*resp.Sync.Value)); err != nil {
			return diag.Errorf("error setting sync: %s", err)
		}
	}

	return diags
}

func resourceTrueNASDatasetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceTrueNASDatasetRead(ctx, d, m)
}

func resourceTrueNASDatasetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*Client)
	id := d.Id()

	err := c.Datasets.Delete(ctx, id)

	if err != nil {
		return diag.Errorf("error deleting dataset: %s", err)
	}

	return diags
}
