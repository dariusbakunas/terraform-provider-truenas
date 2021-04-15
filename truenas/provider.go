package truenas

import (
	"context"
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
				DefaultFunc: schema.EnvDefaultFunc("TRUENAS_API_KEY", nil),
			},
			"base_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "TrueNAS API base URL, eg. https://your.nas/api/v2.0",
				DefaultFunc: schema.EnvDefaultFunc("TRUENAS_BASE_URL", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"truenas_dataset": resourceTrueNASDataset(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"truenas_pools": dataSourceTrueNASPools(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	baseURL := d.Get("base_url").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: apiKey},
	)
	tc := oauth2.NewClient(ctx, ts)
	c, err := NewClient(baseURL, tc)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create TrueNAS API client",
			Detail:   err.Error(),
		})
		return nil, diags
	}

	return c, diags
}
