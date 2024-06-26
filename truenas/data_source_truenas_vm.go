package truenas

import (
	"context"
	"fmt"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func dataSourceTrueNASVM() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTrueNASVMRead,
		Schema: map[string]*schema.Schema{
			"vm_id": &schema.Schema{
				Description: "VM ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": &schema.Schema{
				Description: "VM name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": &schema.Schema{
				Description: "VM description",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"bootloader": &schema.Schema{
				Description: "VM bootloader",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vcpus": &schema.Schema{
				Description: "Number of virtual CPUs",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"cores": &schema.Schema{
				Description: "Number of CPU cores",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"threads": &schema.Schema{
				Description: "Number of CPU threads",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"shutdown_timeout": &schema.Schema{
				Description: "Shutdown timeout in seconds",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"memory": &schema.Schema{
				Description: "Total memory available for VM (bytes)",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"autostart": &schema.Schema{
				Description: "`true` if VM is set to autostart",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"time": &schema.Schema{
				Description: "VM system time. Default is `Local`",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"device": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Description: "Device ID",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"type": &schema.Schema{
							Description: "Device type",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"order": &schema.Schema{
							Description: "Device order",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"vm": &schema.Schema{
							Description: "Device VM ID",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"attributes": &schema.Schema{
							Type: schema.TypeMap,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
						},
					},
				},
			},
			"status": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"state": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"pid": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"domain_state": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTrueNASVMRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Get("vm_id").(string))

	if err != nil {
		return diag.FromErr(err)
	}

	resp, _, err := c.VmApi.GetVM(ctx, int32(id)).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error getting VM: %s\n%s", err, body)
	}

	d.Set("name", resp.Name)

	if resp.Bootloader != nil {
		d.Set("bootloader", *resp.Bootloader)
	}

	if resp.Description != nil {
		d.Set("description", *resp.Description)
	}

	if resp.Vcpus != nil {
		d.Set("vcpus", *resp.Vcpus)
	}

	if resp.Cores != nil {
		d.Set("cores", *resp.Cores)
	}

	if resp.Threads != nil {
		d.Set("threads", *resp.Threads)
	}

	if resp.Memory != nil {
		d.Set("memory", *resp.Memory)
	}

	if resp.Autostart != nil {
		d.Set("autostart", *resp.Autostart)
	}

	if resp.ShutdownTimeout != nil {
		d.Set("shutdown_timeout", *resp.ShutdownTimeout)
	}

	if resp.Time != nil {
		d.Set("time", *resp.Time)
	}

	if resp.Devices != nil {
		if err := d.Set("device", flattenVMDevices(resp.Devices)); err != nil {
			return diag.Errorf("error setting VM devices: %s", err)
		}
	}

	if resp.Status != nil {
		if err := d.Set("status", flattenVMStatus(*resp.Status)); err != nil {
			return diag.Errorf("error setting VM status: %s", err)
		}
	}

	d.SetId(strconv.Itoa(int(resp.Id)))

	return diags
}

func flattenVMDevices(d []api.VMDevice) []interface{} {
	res := make([]interface{}, len(d))

	for i, d := range d {
		device := map[string]interface{}{}

		device["id"] = strconv.Itoa(int(*d.Id))
		device["type"] = d.Dtype

		if d.Order != nil {
			device["order"] = *d.Order
		}

		if d.Vm != nil {
			device["vm"] = *d.Vm
		}

		if d.Attributes != nil {
			device["attributes"] = flattenVMDeviceAttributes(d.Attributes)
		}

		res[i] = device
	}

	return res
}

func flattenVMDeviceAttributes(a map[string]interface{}) map[string]string {
	res := make(map[string]string, len(a))

	for k, v := range a {
		if v != nil {
			res[k] = fmt.Sprintf("%v", v)
		}
	}

	return res
}

func flattenVMStatus(s api.VMStatus) []interface{} {
	var res []interface{}

	status := map[string]interface{}{}

	if s.State != nil {
		status["state"] = *s.State
	}

	if s.Pid != nil {
		status["pid"] = *s.Pid
	}

	if s.DomainState != nil {
		status["domain_state"] = *s.DomainState
	}

	res = append(res, status)

	return res
}
