package truenas

import (
	"context"
	api "github.com/dariusbakunas/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"strings"
)

func dataSourceTrueNASZVOL() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTrueNASZVOLRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
			"vol_size_bytes": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceTrueNASZVOLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id := d.Get("id").(string)

	resp, _, err := c.DatasetApi.GetDataset(ctx, id).Execute()

	if err != nil {
		return diag.Errorf("error getting dataset: %s", err)
	}

	if resp.Type != "VOLUME" {
		return diag.Errorf("invalid volume, %s is %s type", id, resp.Type)
	}

	dpath := newDatasetPath(id)

	d.Set("pool", dpath.Pool)

	if dpath.Parent != "" {
		d.Set("parent", dpath.Parent)
	}

	d.Set("name", dpath.Name)

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

	if resp.KeyFormat != nil && resp.KeyFormat.Value != nil {
		if err := d.Set("key_format", strings.ToLower(*resp.KeyFormat.Value)); err != nil {
			return diag.Errorf("error setting key_format: %s", err)
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

	if resp.Readonly != nil {
		if err := d.Set("readonly", strings.ToLower(*resp.Readonly.Value)); err != nil {
			return diag.Errorf("error setting readonly: %s", err)
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

	if resp.Volsize != nil {
		sz, err := strconv.Atoi(resp.Volsize.Rawvalue)

		if err != nil {
			return diag.Errorf("error parsing volsize rawvalue: %s", err)
		}

		if err := d.Set("vol_size_bytes", sz); err != nil {
			return diag.Errorf("error setting vol_size_bytes: %s", err)
		}
	}

	if err := d.Set("encrypted", resp.Encrypted); err != nil {
		return diag.Errorf("error setting encrypted: %s", err)
	}

	d.SetId(resp.Id)

	return diags
}
