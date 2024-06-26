package truenas

import (
	"context"
	api "github.com/dellathefella/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strconv"
)

func resourceTrueNASVM() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceTrueNASVMRead,
		CreateContext: resourceTrueNASVMCreate,
		DeleteContext: resourceTrueNASVMDelete,
		UpdateContext: resourceTrueNASVMUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"vm_id": &schema.Schema{
				Description: "VM ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": &schema.Schema{
				Description: "VM name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": &schema.Schema{
				Description: "VM description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"bootloader": &schema.Schema{
				Description:  "VM bootloader",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"UEFI", "UEFI_CSM", "GRUB"}, false),
				Default:      "UEFI",
			},
			"autostart": &schema.Schema{
				Description: "Set to start this VM when the system boots",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"time": &schema.Schema{
				Description:  "VM system time. Default is `Local`",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "LOCAL",
				ValidateFunc: validation.StringInSlice([]string{"LOCAL", "UTC"}, false),
			},
			"shutdown_timeout": &schema.Schema{
				Description: "The time in seconds the system waits for the VM to cleanly shut down. During system shutdown, the system initiates poweroff for the VM after the shutdown timeout has expired.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     "90",
			},
			"vcpus": &schema.Schema{
				Description: "Number of virtual CPUs to allocate to the virtual machine. The maximum is 16, or fewer if the host CPU limits the maximum. The VM operating system might also have operational or licensing restrictions on the number of CPUs.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     "1",
			},
			"cores": &schema.Schema{
				Description: "Specify the number of cores per virtual CPU socket. The product of vCPUs, cores, and threads must not exceed 16.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     "1",
			},
			"threads": &schema.Schema{
				Description: "Specify the number of threads per core. The product of vCPUs, cores, and threads must not exceed 16.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     "1",
			},
			"memory": &schema.Schema{
				Description: "Allocate RAM for the VM. Minimum value is 256 * 1024 * 1024 B. Units are bytes. Allocating too much memory can slow the system or prevent VMs from running",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     "536870912", // 512MiB
			},
			"device": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Description: "Device ID",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"type": &schema.Schema{
							Description:  "Device type",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"NIC", "DISK", "CDROM", "PCI", "DISPLAY", "RAW"}, false),
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
							Description: "Device attributes specific to device type, check VM resource examples for example device configurations",
							Type:        schema.TypeMap,
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
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

func resourceTrueNASVMRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, _, err := c.VmApi.GetVM(ctx, int32(id)).Execute()

	if err != nil {
		return diag.Errorf("error getting VM: %s", err)
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

	d.Set("vm_id", strconv.Itoa(int(resp.Id)))
	return nil
}

func resourceTrueNASVMCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	input := api.CreateVMParams{
		Name: getStringPtr(d.Get("name").(string)),
	}

	if description, ok := d.GetOk("description"); ok {
		input.Description = getStringPtr(description.(string))
	}

	if bootloader, ok := d.GetOk("bootloader"); ok {
		input.Bootloader = getStringPtr(bootloader.(string))
	}

	if autostart, ok := d.GetOk("autostart"); ok {
		input.Autostart = getBoolPtr(autostart.(bool))
	}

	if time, ok := d.GetOk("time"); ok {
		input.Time = getStringPtr(time.(string))
	}

	if shutdownTimeout, ok := d.GetOk("shutdown_timeout"); ok {
		input.ShutdownTimeout = getInt32Ptr(int32(shutdownTimeout.(int)))
	}

	if vcpus, ok := d.GetOk("vcpus"); ok {
		input.Vcpus = getInt32Ptr(int32(vcpus.(int)))
	}

	if cores, ok := d.GetOk("cores"); ok {
		input.Cores = getInt32Ptr(int32(cores.(int)))
	}

	if threads, ok := d.GetOk("threads"); ok {
		input.Threads = getInt32Ptr(int32(threads.(int)))
	}

	if memory, ok := d.GetOk("memory"); ok {
		input.Memory = getInt64Ptr(int64(memory.(int)))
	}

	if devices, ok := d.GetOk("device"); ok {
		dv, err := expandVMDevice(devices.(*schema.Set).List())

		if err != nil {
			return diag.Errorf("error creating VM: %s", err)
		}

		input.Devices = dv
	}

	resp, _, err := c.VmApi.CreateVM(ctx).CreateVMParams(input).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}
		return diag.Errorf("error creating VM: %s\n%s", err, body)
	}

	d.SetId(strconv.Itoa(int(resp.Id)))
	return resourceTrueNASVMRead(ctx, d, m)
}

func resourceTrueNASVMDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.VmApi.DeleteVM(ctx, int32(id)).Execute()

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}

		return diag.Errorf("error deleting VM: %s\n%s", err, body)
	}

	d.SetId("")

	return nil
}

func resourceTrueNASVMUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := api.UpdateVMParams{
		Name: getStringPtr(d.Get("name").(string)),
	}

	if d.HasChange("description") {
		input.Description = getStringPtr(d.Get("description").(string))
	}

	if d.HasChange("bootloader") {
		input.Bootloader = getStringPtr(d.Get("bootloader").(string))
	}

	if d.HasChange("autostart") {
		input.Autostart = getBoolPtr(d.Get("autostart").(bool))
	}

	if d.HasChange("time") {
		input.Time = getStringPtr(d.Get("time").(string))
	}

	if d.HasChange("shutdown_timeout") {
		input.ShutdownTimeout = getInt32Ptr(int32(d.Get("shutdown_timeout").(int)))
	}

	if d.HasChange("vcpus") {
		input.Vcpus = getInt32Ptr(int32(d.Get("vcpus").(int)))
	}

	if d.HasChange("cores") {
		input.Cores = getInt32Ptr(int32(d.Get("cores").(int)))
	}

	if d.HasChange("threads") {
		input.Threads = getInt32Ptr(int32(d.Get("threads").(int)))
	}

	if d.HasChange("memory") {
		input.Memory = getInt64Ptr(int64(d.Get("memory").(int)))
	}

	if d.HasChange("device") {
		input.Devices, err = expandVMDeviceForUpdate(d.Get("device").(*schema.Set).List(), getInt32Ptr(int32(id)))

		if err != nil {
			return diag.Errorf("error updating VM: %s", err)
		}
	}

	_, _, err = c.VmApi.UpdateVM(ctx, int32(id)).UpdateVMParams(input).Execute()

	// TODO: handle error response like:
	//{{
	//	"vm_update.name": [
	//{
	//"message": "Only alphanumeric characters are allowed.",
	//"errno": 22
	//}
	//]
	//}}

	if err != nil {
		var body []byte
		if apiErr, ok := err.(*api.GenericOpenAPIError); ok {
			body = apiErr.Body()
		}

		return diag.Errorf("error updating VM: %s\n%s", err, body)
	}

	return resourceTrueNASVMRead(ctx, d, m)
}

// TrueNAS api requires vm attribute set on updates even if it is new device
// while that attribute cannot be set during creation (bug?)
func expandVMDeviceForUpdate(d []interface{}, vmID *int32) ([]api.VMDevice, error) {
	if len(d) == 0 {
		return []api.VMDevice{}, nil
	}

	result := make([]api.VMDevice, 0, len(d))

	for _, item := range d {
		dMap := item.(map[string]interface{})
		dType := dMap["type"].(string)

		if dType == "" {
			continue
		}

		device := &api.VMDevice{
			Dtype: dType,
		}

		// assuming order cannot be 0
		if order, ok := dMap["order"].(int); ok && order != 0 {
			device.Order = getInt32Ptr(int32(order))
		}

		// assuming vm cannot be 0
		device.Vm = vmID

		if idStr, ok := dMap["id"]; ok && idStr != "" {
			id, err := strconv.Atoi(idStr.(string))

			if err != nil {
				return nil, err
			}

			device.Id = getInt32Ptr(int32(id))
		}

		if attr, ok := dMap["attributes"]; ok {
			attrMap := attr.(map[string]interface{})

			// a hack to preserve booleans
			for key, val := range attrMap {
				if val.(string) == "false" {
					attrMap[key] = false
				}
				if val.(string) == "true" {
					attrMap[key] = true
				}
			}

			device.Attributes = attrMap
		}

		result = append(result, *device)
	}

	return result, nil
}

func expandVMDevice(d []interface{}) ([]api.VMDevice, error) {
	if len(d) == 0 {
		return []api.VMDevice{}, nil
	}

	result := make([]api.VMDevice, 0, len(d))

	for _, item := range d {
		dMap := item.(map[string]interface{})
		dType := dMap["type"].(string)

		if dType == "" {
			continue
		}

		device := &api.VMDevice{
			Dtype: dType,
		}

		if attr, ok := dMap["attributes"]; ok {
			attrMap := attr.(map[string]interface{})

			// a hack to preserve booleans
			for key, val := range attrMap {
				if val.(string) == "false" {
					attrMap[key] = false
				}
				if val.(string) == "true" {
					attrMap[key] = true
				}
			}

			device.Attributes = attrMap
		}

		result = append(result, *device)
	}

	return result, nil
}
