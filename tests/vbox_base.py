import logging
import typing as t
from dataclasses import dataclass

import vbox_cli


logger = logging.getLogger()


@dataclass
class VirtualBoxMachine:
    """Rough control of Virtualbox machine from python

    There are some existing python virtualbox libraries, but as of 2023-06-26 they didn't seem to work.
    """
    name: str

    def create_from_ova(self, ova_path: str) -> None:
        assert self.name not in vbox_cli.get_all_vm_names(), f"VM {self.name} already exists, cannot create"
        vbox_cli.import_ova(ova_path, self.name)

    def create_via_clone(self, vm_to_clone: str) -> None:
        vbox_cli.clone_vm(vm_to_clone, self.name)

    def destroy(self) -> None:
        vbox_cli.delete_vm_by_name(self.name)

    def start(self) -> None:
        vbox_cli.start_vm(self.name)

    def stop_hard(self) -> None:
        try:
            vbox_cli.stop_vm_hard(self.name)
        except:
            # If alread powered off, just warn
            if self.state == "poweroff":
                logger.warning(f"stop_hard() failed, but machine is off, continuing...")
            else:
                raise

    def vboxmanage(self, options: t.List[str]) -> None:
        vbox_cli.vboxmanage(options, f"vboxmanage on VM: {self.name}")

    def modifyvm(self, options: t.List[str]) -> None:
        vbox_cli.modifyvm(self.name, options)

    @property
    def state(self) -> str:
        vm_state = vbox_cli.vminfo(self.name)["VMState"]
        assert vm_state in [
            "running",
            "poweroff",
        ], f"Detected unknown VMState: '{vm_state}', please add to VirtualBoxMachine.state"
        return vm_state

    def create_and_attach_hdd(self, hdd_name: str, size_mb: int, port: int, allow_base_ports: bool = False) -> None:
        if port < 2 and not allow_base_ports:
            raise Exception("Ports 0-1 are reserved for base drives, set allow_base_ports=True if you want this")
        # Create hard drive
        vbox_cli.create_hdd(hdd_name=hdd_name, size_mb=size_mb)
        try:
            # attach drive
            vbox_cli.attach_hdd(vm_name=self.name, hdd_name=hdd_name, port=port)
        except:
            # delete drive
            vbox_cli.delete_hdd(hdd_name=hdd_name)
            raise NotImplementedError
