package truenas

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
)

type datasetPath struct {
	Pool   string
	Parent string
	Name   string
}

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
			"comments": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"sync": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "'standard' uses the sync settings that have been requested by the client software, 'always' waits for data writes to complete, and 'disabled' never waits for writes to complete.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"standard", "always", "disabled", "inherit"}, false),
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

	if resp.Comments != nil {
		if err := d.Set("comments", resp.Comments.Value); err != nil {
			return diag.Errorf("error setting comments: %s", err)
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
