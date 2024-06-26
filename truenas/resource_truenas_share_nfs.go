package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strconv"
)

func resourceTrueNASShareNFS() *schema.Resource {
	return &schema.Resource{
		Description:   "Creating a Network File System (NFS) share on TrueNAS gives the benefit of making lots of data easily available for anyone with share access. Depending how the share is configured, users accessing the share can be restricted to read or write privileges. To create a new share, make sure a dataset is available with all the data for sharing.",
		CreateContext: resourceTrueNASShareNFSCreate,
		ReadContext:   resourceTrueNASShareNFSRead,
		UpdateContext: resourceTrueNASShareNFSUpdate,
		DeleteContext: resourceTrueNASShareNFSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"sharenfs_id": &schema.Schema{
				Description: "NFS Share ID",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"comment": &schema.Schema{
				Description: "Any notes about this NFS share",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"hosts": &schema.Schema{
				Description: "Authorized hosts (IP/hostname)",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"alldirs": &schema.Schema{
				Description: "Allow mounting subdirectories",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"ro": &schema.Schema{
				Description: "Prohibit writing",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"quiet": &schema.Schema{
				Description: "Restrict some syslog diagnostics. See exports(5)",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"maproot_user": &schema.Schema{
				Description: "Maproot user",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"maproot_group": &schema.Schema{
				Description: "Maproot group",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"mapall_user": &schema.Schema{
				Description: "Mapall user",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"mapall_group": &schema.Schema{
				Description: "Mapall group",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"security": &schema.Schema{
				Description: "Security mechanism activation state and priority. Requires NFSv4. sys, krb5, krb5i, krb5p",
				// order matters, thus TypeList, not TypeSet
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"sys", "krb5", "krb5i", "krb5p"}, false),
				},
			},
			"enabled": &schema.Schema{
				Description: "Enable this share",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"paths": &schema.Schema{
				Description: "Sharing paths",
				Type:        schema.TypeSet,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
			},
			"networks": &schema.Schema{
				Description: "Authorized networks (CIDR)",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceTrueNASShareNFSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, http, err := c.SharingApi.GetShareNFS(ctx, int32(id)).Execute()

	if err != nil {
		// gracefully handle manual deletions
		d.SetId("")
		if http.StatusCode == 404 {
			return nil
		}
		return diag.Errorf("error getting nfs share: %s", err)
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

	if err := d.Set("paths", flattenStringList(resp.Paths)); err != nil {
		return diag.Errorf("error setting paths: %s", err)
	}

	if resp.Networks != nil {
		if err := d.Set("networks", flattenStringList(resp.Networks)); err != nil {
			return diag.Errorf("error setting networks: %s", err)
		}
	}

	d.Set("sharenfs_id", int(resp.Id))

	return diags
}

func resourceTrueNASShareNFSCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	input := expandShareNFS(d)

	resp, _, err := c.SharingApi.CreateShareNFS(ctx).CreateShareNFSParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error creating NFS share: %s\n%s", err, body)
	}

	d.SetId(strconv.Itoa(int(resp.Id)))

	return resourceTrueNASShareNFSRead(ctx, d, m)
}

func resourceTrueNASShareNFSDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting TrueNAS NFS share: %s", strconv.Itoa(id))

	_, err = c.SharingApi.RemoveShareNFS(ctx, int32(id)).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error deleting NFS share: %s\n%s", err, body)
	}

	log.Printf("[INFO] TrueNAS NFS share (%s) deleted", strconv.Itoa(id))
	d.SetId("")

	return diags
}

func resourceTrueNASShareNFSUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)
	share := expandShareNFS(d)

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, _, err = c.SharingApi.UpdateShareNFS(ctx, int32(id)).CreateShareNFSParams(share).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error updating NFS share: %s\n%s", err, body)
	}

	return resourceTrueNASShareNFSRead(ctx, d, m)
}

func expandShareNFS(d *schema.ResourceData) api.CreateShareNFSParams {
	share := api.CreateShareNFSParams{
		Paths: expandStrings(d.Get("paths").(*schema.Set).List()),
	}

	if comment, ok := d.GetOk("comment"); ok {
		share.Comment = getStringPtr(comment.(string))
	}

	if hosts, ok := d.GetOk("hosts"); ok {
		share.Hosts = expandStrings(hosts.(*schema.Set).List())
	}

	// GetOk also checks if the value is different than empty type default (false)

	alldirs := d.Get("alldirs")

	if alldirs != nil {
		share.Alldirs = getBoolPtr(alldirs.(bool))
	}

	ro := d.Get("ro")

	if ro != nil {
		share.Ro = getBoolPtr(ro.(bool))
	}

	quiet := d.Get("quiet")

	if quiet != nil {
		share.Quiet = getBoolPtr(quiet.(bool))
	}

	if maproot_user, ok := d.GetOk("maproot_user"); ok {
		share.MaprootUser = getStringPtr(maproot_user.(string))
	}

	if maproot_group, ok := d.GetOk("maproot_group"); ok {
		share.MaprootGroup = getStringPtr(maproot_group.(string))
	}

	if mapall_user, ok := d.GetOk("mapall_user"); ok {
		share.MapallUser = getStringPtr(mapall_user.(string))
	}

	if mapall_group, ok := d.GetOk("mapall_group"); ok {
		share.MapallGroup = getStringPtr(mapall_group.(string))
	}

	if security, ok := d.GetOk("security"); ok {
		share.Security = expandStrings(security.([]interface{}))
	}

	if networks, ok := d.GetOk("networks"); ok {
		share.Networks = expandStrings(networks.(*schema.Set).List())
	}

	enabled := d.Get("enabled")

	if enabled != nil {
		share.Enabled = getBoolPtr(enabled.(bool))
	}

	return share
}
