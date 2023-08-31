from __future__ import annotations
import argparse
import json
import logging
import os
import re
import socket
import subprocess as sp
import sys
import time
import typing as t
from dataclasses import dataclass

import paramiko  # type: ignore
import scp  # type: ignore

from misc import run_subprocess

# Paramiko logs errors, even if we catch them, so sqlech non-fatal
logging.getLogger("paramiko").setLevel(logging.FATAL)
logging.captureWarnings(True)
logger = logging.getLogger()

PASSWORD: t.Final[str] = 'terraform'

@dataclass
class SSHClientWrapper:
    """Wrapper class for paramiko SSH client"""
    host: str
    port: int
    username: str = 'admin'
    password: str = PASSWORD
    timeout: int = 5
    ssh_retries: int = 0

    _client: t.Optional[paramiko.SSHClient] = None

    def __post_init__(self) -> None:
        assert self.ssh_retries >= 0
        assert self.timeout >= 0

    def __enter__(self) -> SSHClientWrapper:
        attempt = 1
        while True:
            start = time.time()
            try:
                new_client = paramiko.SSHClient()
                new_client.set_missing_host_key_policy(paramiko.client.WarningPolicy)
                new_client.connect(
                    self.host,
                    port=self.port,
                    username=self.username,
                    password=self.password,
                    timeout=self.timeout,
                    look_for_keys=False,
                )

                # everything worked, assign and return it
                self._client = new_client
                return self

            except paramiko.ssh_exception.SSHException as e:
                if attempt-1 == self.ssh_retries:
                    raise Exception(f"Failed SSH connection {self.ssh_retries} times") from e
                logger.info(f"SSH attempt {attempt} timed out in {time.time()-start:.2f}s")
                time.sleep(1)
                attempt += 1


    def __exit__(self, *args: t.Any, **kwargs: t.Any) -> None:
        assert self._client is not None, "SSHClient.__exit__() called but 'client' is None"
        self.client.close()

    @property
    def client(self) -> paramiko.SSHClient:
        assert self._client is not None
        return self._client

    def exec_command(self, cmd: str, target_rc: t.Optional[int] = 0, printall: bool = False):
        def print_stream(name: str, stream) -> None:
            print(f"Output from {name}:")
            print(stream.read().decode())

        stdin, stdout, stderr = self.client.exec_command(cmd)

        if target_rc is not None:
            rc = stdout.channel.recv_exit_status()
            if rc != target_rc:
                print_stream("stdout", stdout)
                print_stream("stderr", stderr)
                raise Exception(f"ssh.exec_command({cmd}) failed -> rc={rc}, expected={target_rc}")
        if printall:
            print_stream("stdout", stdout)
            print_stream("stderr", stderr)
        return stdin, stdout, stderr

    def sudo_exec_command(self, cmd: str, **kwargs: t.Any):
        return self.exec_command(cmd=f"echo '{PASSWORD}' | sudo -S {cmd}", **kwargs)

    def scp_get(self, remote_path: str, local_path: str) -> None:
        scp_client = scp.SCPClient(self.client.get_transport())
        scp_client.get(remote_path, local_path)
