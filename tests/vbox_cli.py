import logging
import os
import re
import subprocess as sp
import sys
import time
import typing as t
from dataclasses import dataclass
from typing import List, Optional


from misc import run_subprocess
from ssh_utils import SSHClientWrapper


logger = logging.getLogger()

def vboxmanage(cli_args: t.List[str], fail_notice: t.Optional[str] = None) -> None:
    run_subprocess(
        cli_args,
        fail_notice or "vboxmanage failed",
    )

def clone_vm(vm_to_clone: str, new_vm_name: str) -> None:
    run_subprocess(
        [
            "vboxmanage",
            "clonevm",
            vm_to_clone,
            f"--name={new_vm_name}",
            "--register",
        ],
        f"Cloning VM {vm_to_clone} -> {new_vm_name}",
    )

def delete_vm_by_name(name: str) -> None:
    run_subprocess(
        [
            "vboxmanage",
            "unregistervm",
            name,
            "--delete-all",
        ],
        f"Deleting VM {name}",
    )


def get_all_vm_names() -> List[str]:
    out = sp.check_output(["VBoxManage", "list", "vms"])
    lines = out.decode().strip().split("\n")
    vms = []
    for line in lines:
        m = re.match(r'"(.*)" .*', line)
        assert m is not None, f"Expected format '\"my-vmname\" {{<uuid>}}', got: '{line}'"
        assert len(m.groups()) == 1
        vms.append(m.group(1))
    return vms


def import_ova(ova_path: str, vm_name: str) -> None:
    run_subprocess(
        [
            "vboxmanage",
            "import",
            ova_path,
            "--vsys=0",
            f"--vmname={vm_name}",
        ],
        f"Importing .ova {ova_path} to new VM {vm_name}",
    )

def modifyvm(
    vm_name: str,
    options: List[str],
) -> None:
    cmd = [
        "vboxmanage",
        "modifyvm",
        vm_name,
    ]

    run_subprocess(
        cmd + options,
        f"Modifying VM {vm_name}",
    )

def start_vm(vm_name: str) -> None:
    run_subprocess(
        [
            "vboxmanage",
            "startvm",
            vm_name,
            "--type=headless",
        ],
        f"Starting VM {vm_name}"
    )

def stop_vm_hard(vm_name: str) -> None:
    run_subprocess(
        [
            "vboxmanage",
            "controlvm",
            vm_name,
            "poweroff",
        ],
        f"Stopping VM {vm_name}"
    )

def vminfo(vm_name: str) -> t.Mapping[str, t.Any]:
    cmd = ["vboxmanage", "showvminfo", vm_name, "--machinereadable"]
    ret = run_subprocess(cmd, f"Getting info for VM {vm_name}")
    info = {}
    for line in ret.stdout.decode().strip().split("\n"):
        # Skip some troublesome lines:
        if any(line.startswith(x) for x in [
            " rec_screen",
            "VideoMode=",
            "GuestAdditionsFacility_",
        ]):
            continue

        if '==' in line:
            key, val = line.split("==")
        elif '=' in line:
            key, _, val = line.partition("=")
        elif ':' in line:
            key, val = line.split(":")
        elif line.startswith(" rec_screen"):
            pass  # ignore
        else:
            raise NotImplementedError(f"Unsure how to parse vminfo line: {line}")

        key = key.strip()
        val = val.strip()

        if val.startswith('"') and val.endswith('"'):
            # string value, just strip quotes
            val = val[1:-1]
        elif val == "enabled":
            val = True
        elif val == "disabled":
            val = False
        elif val.isdigit():
            val = int(val)
        else:
            raise NotImplementedError(f"Unsure how to parse vminfo val: {val} (key={key})")

        info[key] = val

    return info

def create_hdd(hdd_name: str, size_mb: int):
    run_subprocess(
        [
            "vboxmanage",
            "createmedium",
            "disk",
            f"--filename={hdd_name}.vdi",
            f"--size={size_mb}",
        ],
        f"Creating VM HDD {hdd_name}",
    )

def delete_hdd(hdd_name: str):
    run_subprocess(
        [
            "vboxmanage",
            "closemedium",
            "disk",
            f"{hdd_name}.vdi",
            "--delete",
        ],
        f"Deleting VM HDD {hdd_name}",
    )

def attach_hdd(vm_name: str, hdd_name: str, port: int) -> None:
    run_subprocess(
        [
            "vboxmanage",
            "storageattach",
            vm_name,
            "--storagectl=hdd",
            f"--port={port}",
            "--device=0",
            f"--medium={hdd_name}.vdi",
            "--type=hdd",
        ],
        f"Attaching HDD {hdd_name} to VM {vm_name}",
    )

def attach_optical(vm_name) -> None:
    run_subprocess(
        [
            "vboxmanage",
            "storageattach",
            vm_name,
            "--storagectl=optical",
            "--port=0",
            "--device=0",
            f"--medium=emptydrive",
            "--type=dvddrive",
        ],
        f"Attaching optical drive to VM {vm_name}",
    )
