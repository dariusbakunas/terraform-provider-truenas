#!/usr/bin/env python
import logging
import os
import subprocess as sp
import sys
from typing import List

import vbox_cli
from misc import print_ind, run_subprocess
from vbox_base import VirtualBoxMachine


logger = logging.getLogger()
BASE_VM_NAME = "truenas-test-base"
BASE_HDD_NAME = "truenas-test-base-hdd"

INSTALL_INSTRUCTIONS = f"""
------------------------------------------------------------------------------
You will need to manually install TrueNAS in your new VM {BASE_VM_NAME}.

Installation instructions:
- Open VirtualBox
- Click on [{BASE_VM_NAME}] and run [Start]
- You should be prompted for DVD
- Browse to your TrueNAS installation .iso, then click [Mount and Retry Boot]
- If grub2 blue menu comes up, select main option (or wait 15s for it to auto)
- Now you should be within the TrueNAS setup menu
  - Select "Install/Upgrade" and hit <enter>
  - Press <space> to select "sda" HDD, then hit <enter>
  - Proceed? Select <Yes>
  - Select "Administrative user (admin)"
  - Password: "terraform" (without quotes)
  - Wait for installation to finish (~10mins?)
  - You will be prompted to reboot and remote installation media
    - Devices -> Optical Drive -> Remove disk -> Force unmount
    - click OK
  - Select "Reboot"
- Once it reboots, you should see the "Console setup" menu
  - Select (7) to "Open Linux Shell"
  - Type: 'systemctl enable ssh' and press enter
  - Press ctrl-d to exit, then select (9) to shutdown the VM

"""


def create_base_vm(vm_name: str):
    run_subprocess(
        [
            "vboxmanage",
            "createvm",
            f"--name={vm_name}",
            "--ostype=Debian_64",
            "--register",
        ],
        f"Creating VM {vm_name}",
    )

def configure_base_vm(vm_name: str):
    # Some main configuration options
    vbox_cli.modifyvm(
        vm_name,
        [
            "--memory=8192", # 8G RAM
            "--graphicscontroller=vmsvga",
            "--vram=16",
            "--audio=none",
            "--boot1=dvd",
            "--boot2=disk",
            "--boot3=none",
            "--boot4=none",
        ],
    )
    # Create storage controller for optical drive
    run_subprocess(
        [
            "vboxmanage",
            "storagectl",
            vm_name,
            "--name=optical",
            "--add=ide",
            "--controller=PIIX4",
            "--bootable=on",
        ],
        f"Adding optical drive to VM {vm_name}",
    )
    # Create storage controller for HDD
    run_subprocess(
        [
            "vboxmanage",
            "storagectl",
            vm_name,
            "--name=hdd",
            "--add=sata",
            "--controller=IntelAhci",
            # 1 for base HDD, 1 for default pool HDD, 12 for test usage
            "--portcount=14",
            "--bootable=on",
        ],
        f"Adding HDD controller to VM {vm_name}",
    )

def create_base_test_vm() -> None:
    assert BASE_VM_NAME not in vbox_cli.get_all_vm_names(), f"Base VM name {BASE_VM_NAME} already exists, delete it fully first"
    # Create VM
    vm = VirtualBoxMachine(BASE_VM_NAME)
    create_base_vm(vm.name)
    try:
        # Modify VM
        configure_base_vm(vm.name)
        # Add Add base HDD
        vm.create_and_attach_hdd(hdd_name=BASE_HDD_NAME, size_mb=8192, port=0, allow_base_ports=True)
        vbox_cli.attach_optical(vm_name=vm.name)

        while True:
            print_ind(INSTALL_INSTRUCTIONS.strip(), 5)
            print("")
            print("")
            print("Please follow the above installation instructions to install.")
            print("WARNING: Whether successful or not, please SHUT DOWN the VM before typing a command")
            command = input("Was the installation successful? (y/n)?").strip()
            if command.lower() == "y":
                break
            if command.lower() == "n":
                raise Exception("Cancelled by user")
            else:
                print(f"Unexpected input: '{command}'")

    except Exception as e:
        # Whether success or failure, we want to clean up VM
        logger.warning("An exception occurred, attempting to delete VM...")
        vbox_cli.delete_vm_by_name(BASE_VM_NAME)
        raise Exception(f"Setup failed, ensure VM is deleted before you retry") from e


if __name__ == "__main__":
    create_base_test_vm()
