import logging
import socket
import typing as t
from dataclasses import dataclass

from vbox_base import VirtualBoxMachine


logger = logging.getLogger()


@dataclass
class TrueNASVirtualBoxMachine(VirtualBoxMachine):
    truenas_api_port: t.Optional[int] = None
    ssh_port: t.Optional[int] = None

    def create_from_ova(self, ova_path: str) -> None:
        super().create_from_ova(ova_path=ova_path)
        self.ensure_truenas_api_port()
        self.ensure_ssh_port()

    def create_via_clone(self, vm_to_clone: str) -> None:
        super().create_via_clone(vm_to_clone=vm_to_clone)
        self.ensure_truenas_api_port()
        self.ensure_ssh_port()

    def ensure_truenas_api_port(self):
        # Add a port forwarding port
        if self.truenas_api_port is None:
            # By binding to port 0, we get a port created
            sock = socket.socket()
            sock.bind(('', 0))
            self.truenas_api_port = sock.getsockname()[1]
            self.modifyvm([f"--natpf1=http,tcp,,{self.truenas_api_port},,80"])

    def ensure_ssh_port(self):
        if self.ssh_port is None:
            # By binding to port 0, we get a port created
            sock = socket.socket()
            sock.bind(('', 0))
            self.ssh_port = sock.getsockname()[1]
            self.modifyvm([f"--natpf1=ssh,tcp,,{self.ssh_port},,22"])

    def get_ssh_port(self) -> int:
        assert self.ssh_port is not None
        return self.ssh_port
