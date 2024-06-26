package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/oauth2"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Description: "TrueNAS API key",
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TRUENAS_API_KEY", nil),
			},
			"base_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "TrueNAS API base URL, eg. https://your.nas/api/v2.0",
				DefaultFunc: schema.EnvDefaultFunc("TRUENAS_BASE_URL", nil),
			},
			"debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "DEBUG: dump all API requests/responses",
				DefaultFunc: schema.EnvDefaultFunc("TRUENAS_DEBUG", false),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"truenas_cronjob":   resourceTrueNASCronjob(),
			"truenas_dataset":   resourceTrueNASDataset(),
			"truenas_share_nfs": resourceTrueNASShareNFS(),
			"truenas_share_smb": resourceTrueNASShareSMB(),
			"truenas_zvol":      resourceTrueNASZVOL(),
			"truenas_vm":        resourceTrueNASVM(),
			"truenas_group":     resourceTrueNASGroup(),
			"truenas_user":      resourceTrueNASUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"truenas_cronjob":               dataSourceTrueNASCronjob(),
			"truenas_dataset":               dataSourceTrueNASDataset(),
			"truenas_network_configuration": dataSourceTrueNASNetworkConfiguration(),
			"truenas_pool_ids":              dataSourceTrueNASPoolIDs(),
			"truenas_service":               dataSourceTrueNASService(),
			"truenas_share_nfs":             dataSourceTrueNASShareNFS(),
			"truenas_share_smb":             dataSourceTrueNASShareSMB(),
			"truenas_vm":                    dataSourceTrueNASVM(),
			"truenas_zvol":                  dataSourceTrueNASZVOL(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	baseURL := d.Get("base_url").(string)
	debug := d.Get("debug").(bool)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: apiKey},
	)
	tc := oauth2.NewClient(ctx, ts)

	config := api.NewConfiguration()
	config.Servers = api.ServerConfigurations{
		{
			URL: baseURL,
		},
	}
	config.Debug = debug
	config.HTTPClient = tc

	c := api.NewAPIClient(config)
	return c, diags
}
