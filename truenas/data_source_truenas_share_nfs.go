package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func dataSourceTrueNASShareNFS() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about specific NFS share",
		ReadContext: dataSourceTrueNASShareNFSRead,
		Schema: map[string]*schema.Schema{
			"sharenfs_id": &schema.Schema{
				Description: "NFS Share ID",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"comment": &schema.Schema{
				Description: "Any notes about this NFS share",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"hosts": &schema.Schema{
				Description: "Authorized hosts (IP/hostname)",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"alldirs": &schema.Schema{
				Description: "Allow mounting subdirectories",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"ro": &schema.Schema{
				Description: "Prohibit writing",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"quiet": &schema.Schema{
				Description: "Restrict some syslog diagnostics. See exports(5)",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"maproot_user": &schema.Schema{
				Description: "Maproot user",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"maproot_group": &schema.Schema{
				Description: "Maproot group",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"mapall_user": &schema.Schema{
				Description: "Mapall user",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"mapall_group": &schema.Schema{
				Description: "Mapall group",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"security": &schema.Schema{
				Description: "Security mechanism activation state and priority. Requires NFSv4. sys, krb5, krb5i, krb5p",
				// order matters
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"enabled": &schema.Schema{
				Description: "Enable this share",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"locked": &schema.Schema{
				Description: "Locked status",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"paths": &schema.Schema{
				Description: "Sharing paths",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"networks": &schema.Schema{
				Description: "Authorized networks (CIDR)",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceTrueNASShareNFSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id := d.Get("sharenfs_id").(int)

	resp, _, err := c.SharingApi.GetShareNFS(ctx, int32(id)).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error getting share: %s\n%s", err, body)
	}

	if resp.Comment != nil {
		d.Set("comment", *resp.Comment)
	}

	if resp.Hosts != nil {
		if err := d.Set("hosts", flattenStringList(resp.Hosts)); err != nil {
			return diag.Errorf("error setting hosts: %s", err)
		}
	}

	d.Set("alldirs", *resp.Alldirs)
	d.Set("ro", *resp.Ro)
	d.Set("quiet", *resp.Quiet)

	if resp.MaprootUser != nil {
		d.Set("maproot_user", *resp.MaprootUser)
	}

	if resp.MaprootGroup != nil {
		d.Set("maproot_group", *resp.MaprootGroup)
	}

	if resp.MapallUser != nil {
		d.Set("mapall_user", *resp.MapallUser)
	}

	if resp.MapallGroup != nil {
		d.Set("mapall_group", *resp.MapallGroup)
	}

	if resp.Security != nil {
		if err := d.Set("security", flattenStringList(resp.Security)); err != nil {
			return diag.Errorf("error setting security: %s", err)
		}
	}

	d.Set("enabled", *resp.Enabled)
	d.Set("locked", *resp.Locked)

	if err := d.Set("paths", flattenStringList(resp.Paths)); err != nil {
		return diag.Errorf("error setting paths: %s", err)
	}

	if resp.Networks != nil {
		if err := d.Set("networks", flattenStringList(resp.Networks)); err != nil {
			return diag.Errorf("error setting networks: %s", err)
		}
	}

	d.SetId(strconv.Itoa(int(resp.Id)))

	return diags
}
