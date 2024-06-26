package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"strings"
)

func dataSourceTrueNASZVOL() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTrueNASZVOLRead,
		Schema: map[string]*schema.Schema{
			"zvol_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"blocksize": &schema.Schema{
				Description: "Volume blocksize",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"comments": &schema.Schema{
				Description: "Any notes about this volume.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"compression": &schema.Schema{
				Description: "Current zvol compression level",
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
			"encryption_algorithm": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_format": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"locked": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"key_loaded": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"pbkdf2iters": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"pool": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"readonly": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
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
				Type:        schema.TypeString,
				Description: "Sets the data write synchronization. `inherit` takes the sync settings from the parent dataset, `standard` uses the settings that have been requested by the client software, `always` waits for data writes to complete, and `disabled` never waits for writes to complete.",
				Computed:    true,
			},
			"volsize": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceTrueNASZVOLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id := d.Get("zvol_id").(string)

	resp, _, err := c.DatasetApi.GetDataset(ctx, id).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error getting zvol: %s\n%s", err, body)
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

	if resp.Sync != nil {
		d.Set("sync", strings.ToLower(*resp.Sync.Value))
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

	if resp.Encrypted != nil {
		d.Set("encrypted", *resp.Encrypted)
	}

	d.SetId(resp.Id)

	return diags
}
