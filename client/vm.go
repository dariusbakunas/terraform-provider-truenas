package client

import (
	"context"
	"github.com/dariusbakunas/truenas-go-sdk/angelfish"
	"github.com/dariusbakunas/truenas-go-sdk/bluefin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

type VMDevice struct {
	Id                   *int32
	Dtype                string
	Attributes           map[string]interface{}
	Order                *int32
	Vm                   *int32
	AdditionalProperties map[string]interface{}
}

type VMStatus struct {
	State                *string
	Pid                  *int32
	DomainState          *string
	AdditionalProperties map[string]interface{}
}

type VM struct {
	Id              int32
	Name            string
	Description     *string
	Vcpus           *int32
	Memory          *int64
	Autostart       *bool
	Time            *string
	Bootloader      *string
	Cores           *int32
	Threads         *int32
	ShutdownTimeout *int32
	Devices         []VMDevice
	Status          *VMStatus
}

func convertFromAngelfishVM(vm *angelfish.VM) *VM {
	devices := make([]VMDevice, 0, len(vm.Devices))

	for _, device := range vm.Devices {
		devices = append(devices, VMDevice{
			Id:         device.Id,
			Dtype:      device.Dtype,
			Vm:         device.Vm,
			Order:      device.Order,
			Attributes: device.Attributes,
		})
	}

	return &VM{
		Id:              vm.Id,
		Name:            vm.Name,
		Bootloader:      vm.Bootloader,
		Description:     vm.Description,
		Vcpus:           vm.Vcpus,
		Cores:           vm.Cores,
		Threads:         vm.Threads,
		Memory:          vm.Memory,
		Autostart:       vm.Autostart,
		ShutdownTimeout: vm.ShutdownTimeout,
		Time:            vm.Time,
		Devices:         devices,
	}
}

func convertFromBluefinVM(vm *bluefin.VM) *VM {
	devices := make([]VMDevice, 0, len(vm.Devices))

	for _, device := range vm.Devices {
		devices = append(devices, VMDevice{
			Id:         device.Id,
			Dtype:      device.Dtype,
			Vm:         device.Vm,
			Order:      device.Order,
			Attributes: device.Attributes,
		})
	}

	return &VM{
		Id:              vm.Id,
		Name:            vm.Name,
		Bootloader:      vm.Bootloader,
		Description:     vm.Description,
		Vcpus:           vm.Vcpus,
		Cores:           vm.Cores,
		Threads:         vm.Threads,
		Memory:          vm.Memory,
		Autostart:       vm.Autostart,
		ShutdownTimeout: vm.ShutdownTimeout,
		Time:            vm.Time,
		Devices:         devices,
	}
}

type NullableString struct {
	value *string
	isSet bool
}

type CreateVMParams struct {
	CpuMode             *string
	CpuModel            NullableString
	Name                *string
	Description         *string
	Vcpus               *int32
	Cores               *int32
	Threads             *int32
	Memory              *int64
	Bootloader          *string
	Devices             []VMDevice
	Autostart           *bool
	HideFromMsr         *bool
	EnsureDisplayDevice *bool
	Time                *string
	ShutdownTimeout     *int32
	ArchType            NullableString
	MachineType         NullableString
	Uuid                NullableString
}

type UpdateVMParams struct {
	Name                 *string
	CpuMode              *string
	CpuModel             NullableString
	Description          *string
	Vcpus                *int32
	Cores                *int32
	Threads              *int32
	Memory               *int64
	Bootloader           *string
	Devices              []VMDevice
	Autostart            *bool
	HideFromMsr          *bool
	EnsureDisplayDevice  *bool
	Time                 *string
	ShutdownTimeout      *int32
	ArchType             NullableString
	MachineType          NullableString
	Uuid                 NullableString
	AdditionalProperties map[string]interface{}
}

type VmApi struct {
	Client interface{}
}

func (api *VmApi) GetVM(ctx context.Context, id int) (*VM, diag.Diagnostics) {
	switch client := api.Client.(type) {
	case *angelfish.APIClient:
		resp, _, err := client.VmApi.GetVM(ctx, int32(id)).Execute()

		if err != nil {
			if apiErr, ok := err.(*angelfish.GenericOpenAPIError); ok {
				return nil, diag.Errorf("error reading VM: %s\n%s", err, apiErr.Body())
			}

			return nil, diag.Errorf("error reading VM: %s", err)
		}

		return convertFromAngelfishVM(resp), nil
	case *bluefin.APIClient:
		resp, _, err := client.VmApi.GetVM(ctx, int32(id)).Execute()

		if err != nil {
			if apiErr, ok := err.(*bluefin.GenericOpenAPIError); ok {
				return nil, diag.Errorf("error reading VM: %s\n%s", err, apiErr.Body())
			}

			return nil, diag.Errorf("error reading VM: %s", err)
		}

		return convertFromBluefinVM(resp), nil
	default:
		return nil, diag.Errorf("unsupported api client: %T", api)
	}
}

func (api *VmApi) CreateVM(ctx context.Context, vm CreateVMParams) (*VM, diag.Diagnostics) {
	switch client := api.Client.(type) {
	case *angelfish.APIClient:
		devices := make([]angelfish.VMDevice, 0, len(vm.Devices))

		for _, device := range vm.Devices {
			devices = append(devices, angelfish.VMDevice{
				Id:         device.Id,
				Dtype:      device.Dtype,
				Vm:         device.Vm,
				Order:      device.Order,
				Attributes: device.Attributes,
			})
		}

		input := angelfish.CreateVMParams{
			Name:            vm.Name,
			Description:     vm.Description,
			Bootloader:      vm.Bootloader,
			Autostart:       vm.Autostart,
			Time:            vm.Time,
			ShutdownTimeout: vm.ShutdownTimeout,
			Vcpus:           vm.Vcpus,
			Cores:           vm.Cores,
			Threads:         vm.Threads,
			Memory:          vm.Memory,
			Devices:         devices,
		}

		resp, _, err := client.VmApi.CreateVM(ctx).CreateVMParams(input).Execute()

		if err != nil {
			if apiErr, ok := err.(*angelfish.GenericOpenAPIError); ok {
				return nil, diag.Errorf("error creating VM: %s\n%s", err, apiErr.Body())
			}

			return nil, diag.Errorf("error creating VM: %s", err)
		}

		return convertFromAngelfishVM(resp), nil
	case *bluefin.APIClient:
		// bluefin does not have devices prop here
		// TODO: add support for new properties, like MinMemory
		input := bluefin.CreateVMParams{
			Name:            vm.Name,
			Description:     vm.Description,
			Bootloader:      vm.Bootloader,
			Autostart:       vm.Autostart,
			Time:            vm.Time,
			ShutdownTimeout: vm.ShutdownTimeout,
			Vcpus:           vm.Vcpus,
			Cores:           vm.Cores,
			Threads:         vm.Threads,
			Memory:          vm.Memory,
		}

		resp, _, err := client.VmApi.CreateVM(ctx).CreateVMParams(input).Execute()

		if err != nil {
			if apiErr, ok := err.(*bluefin.GenericOpenAPIError); ok {
				return nil, diag.Errorf("error creating VM: %s\n%s", err, apiErr.Body())
			}

			return nil, diag.Errorf("error creating VM: %s", err)
		}

		return convertFromBluefinVM(resp), nil
	default:
		return nil, diag.Errorf("unsupported api client: %T", api)
	}
}

func (api *VmApi) DeleteVM(ctx context.Context, id int32) diag.Diagnostics {
	switch client := api.Client.(type) {
	case *angelfish.APIClient:
		_, err := client.VmApi.DeleteVM(ctx, id).Execute()

		if err != nil {
			if apiErr, ok := err.(*angelfish.GenericOpenAPIError); ok {
				return diag.Errorf("error deleting VM: %s\n%s", err, apiErr.Body())
			}

			return diag.Errorf("error deleting VM: %s", err)
		}
	case *bluefin.APIClient:
		_, err := client.VmApi.DeleteVM(ctx, id).Execute()

		if err != nil {
			if apiErr, ok := err.(*bluefin.GenericOpenAPIError); ok {
				return diag.Errorf("error deleting VM: %s\n%s", err, apiErr.Body())
			}

			return diag.Errorf("error deleting VM: %s", err)
		}
	default:
		return diag.Errorf("unsupported api client: %T", api)
	}

	return nil
}

func (api *VmApi) UpdateVM(ctx context.Context, id int32, vm UpdateVMParams) (*VM, diag.Diagnostics) {
	switch client := api.Client.(type) {
	case *angelfish.APIClient:
		input := angelfish.UpdateVMParams{
			Name:            vm.Name,
			Description:     vm.Description,
			Bootloader:      vm.Bootloader,
			Autostart:       vm.Autostart,
			Time:            vm.Time,
			ShutdownTimeout: vm.ShutdownTimeout,
			Vcpus:           vm.Vcpus,
			Cores:           vm.Cores,
			Threads:         vm.Threads,
			Memory:          vm.Memory,
		}

		if vm.Devices != nil {
			devices := make([]angelfish.VMDevice, 0, len(vm.Devices))

			for _, device := range vm.Devices {
				devices = append(devices, angelfish.VMDevice{
					Id:         device.Id,
					Dtype:      device.Dtype,
					Vm:         device.Vm,
					Order:      device.Order,
					Attributes: device.Attributes,
				})
			}
		}

		resp, _, err := client.VmApi.UpdateVM(ctx, id).UpdateVMParams(input).Execute()

		if err != nil {
			if apiErr, ok := err.(*angelfish.GenericOpenAPIError); ok {
				return nil, diag.Errorf("error updating VM: %s\n%s", err, apiErr.Body())
			}

			return nil, diag.Errorf("error updating VM: %s", err)
		}

		return convertFromAngelfishVM(resp), nil
	case *bluefin.APIClient:
		input := bluefin.UpdateVMParams{
			Name:            vm.Name,
			Description:     vm.Description,
			Bootloader:      vm.Bootloader,
			Autostart:       vm.Autostart,
			Time:            vm.Time,
			ShutdownTimeout: vm.ShutdownTimeout,
			Vcpus:           vm.Vcpus,
			Cores:           vm.Cores,
			Threads:         vm.Threads,
			Memory:          vm.Memory,
		}

		resp, _, err := client.VmApi.UpdateVM(ctx, id).UpdateVMParams(input).Execute()

		if err != nil {
			if apiErr, ok := err.(*bluefin.GenericOpenAPIError); ok {
				return nil, diag.Errorf("error updating VM: %s\n%s", err, apiErr.Body())
			}

			return nil, diag.Errorf("error updating VM: %s", err)
		}

		return convertFromBluefinVM(resp), nil
	default:
		return nil, diag.Errorf("unsupported api client: %T", api)
	}
}
