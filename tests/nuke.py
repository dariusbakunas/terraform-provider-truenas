import sys
import time

from test_base import TRUENAS_VM_NAME_PREFIX
from vbox_base import VirtualBoxMachine
from vbox_cli import get_all_vm_names

DELAY = 10

def main() -> None:
    print("== Destroying all TrueNas Testing VirtualBox VMs ==")
    all_vm_names = get_all_vm_names()
    test_vm_names = sorted([x for x in all_vm_names if x.startswith(TRUENAS_VM_NAME_PREFIX)])
    del all_vm_names  # clean scope to ensure we don't access others
    for name in test_vm_names:
        print(f"- {name}")

    print(f"DELETING ALL IN {DELAY} SECONDS!!!")
    for i in reversed(range(1, DELAY)):
        print(f"{i}...")
        time.sleep(1)

    print("Deletion Started!")
    for name in test_vm_names:
        sys.stdout.write(f"{name}... ")
        sys.stdout.flush()
        vm = VirtualBoxMachine(name=name)
        vm.destroy()
        print("deleted!")

if __name__ == "__main__":
    main()
