import argparse
import os
import subprocess as sp
import sys
from typing import List

def print_ind(txt, i) -> None:
    lines = txt.split("\n")
    for line in lines:
        print(f"{' '*i}{line}")

def run_subprocess(
    cmd: List[str],
    notice: str,
) -> sp.CompletedProcess:
    sys.stdout.write(notice)
    sys.stdout.write("...")
    sys.stdout.flush()
    ret = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE)
    if ret.returncode == 0:
        print("success!")
        return ret
    else:
        print("failed!")
        print(f"- cmd: {cmd}")
        print("- stdout:")
        print_ind(ret.stdout.decode(), 5)
        print("- stderr:")
        print_ind(ret.stderr.decode(), 5)

        raise Exception(f"Something went wrong with: {notice}")
