package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTrueNASNetworkConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "Get network configuration",
		ReadContext: dataSourceTrueNASNetworkConfigurationRead,
		Schema: map[string]*schema.Schema{
			"hostname": &schema.Schema{
				Description: "TrueNAS hostname",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"domain": &schema.Schema{
				Description: "TrueNAS domain",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ipv4gateway": &schema.Schema{
				Description: "Gateway IPv4 address",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ipv6gateway": &schema.Schema{
				Description: "Gateway IPv6 address",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"nameserver1": &schema.Schema{
				Description: "Nameserver 1 IPv4 address",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"nameserver2": &schema.Schema{
				Description: "Nameserver 2 IPv4 address",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"nameserver3": &schema.Schema{
				Description: "Nameserver 3 IPv4 address",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"httpproxy": &schema.Schema{
				Description: "HTTP proxy address",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"netwait_enabled": &schema.Schema{
				Description: "`true` if netwait feature is enabled",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"netwait_ips": &schema.Schema{
				Description: "List of IP addresses to ping if netwait is enabled",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"service_announcement": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"netbios": &schema.Schema{
							Description: "Advertises the SMB service NetBIOS Name",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"mdns": &schema.Schema{
							Description: "Multicast DNS. Uses the system Hostname to advertise enabled and running services",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"wsd": &schema.Schema{
							Description: "Uses the SMB Service NetBIOS Name to advertise the server to WS-Discovery clients",
							Type:        schema.TypeBool,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTrueNASNetworkConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)

	config, _, err := c.NetworkApi.GetNetworkConfiguration(ctx).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error getting network configuration: %s\n%s", err, body)
	}

	if config.Hostname != nil {
		d.Set("hostname", *config.Hostname)
	}

	if config.Domain != nil {
		d.Set("domain", *config.Domain)
	}

	if config.Ipv4gateway != nil {
		d.Set("ipv4gateway", *config.Ipv4gateway)
	}

	if config.Ipv6gateway != nil {
		d.Set("ipv6gateway", *config.Ipv6gateway)
	}

	if config.Nameserver1 != nil {
		d.Set("nameserver1", *config.Nameserver1)
	}

	if config.Nameserver2 != nil {
		d.Set("nameserver2", *config.Nameserver2)
	}

	if config.Nameserver3 != nil {
		d.Set("nameserver3", *config.Nameserver3)
	}

	if config.Httpproxy != nil {
		d.Set("httpproxy", *config.Httpproxy)
	}

	if config.NetwaitEnabled != nil {
		d.Set("netwait_enabled", *config.NetwaitEnabled)
	}

	if config.NetwaitIp != nil {
		if err := d.Set("netwait_ips", flattenStringList(config.NetwaitIp)); err != nil {
			return diag.Errorf("error setting netwait_ips: %s", err)
		}
	}

	if config.ServiceAnnouncement != nil {
		if err := d.Set("service_announcement", flattenServiceAnnouncement(*config.ServiceAnnouncement)); err != nil {
			return diag.Errorf("error setting service_announcement: %s", err)
		}
	}

	d.SetId("network-summary")

	return diags
}

func flattenServiceAnnouncement(s api.NetworkConfigServiceAnnouncement) []interface{} {
	var res []interface{}

	announcement := map[string]interface{}{}

	if s.Mdns != nil {
		announcement["mdns"] = *s.Mdns
	}

	if s.Netbios != nil {
		announcement["netbios"] = *s.Netbios
	}

	if s.Wsd != nil {
		announcement["wsd"] = *s.Wsd
	}

	res = append(res, announcement)

	return res
}
