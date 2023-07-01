import json
import IPython
import logging
import os
import sys
import tempfile
import time
import typing as t
import unittest

import requests

from create_base_test_vm import BASE_VM_NAME
from misc import run_subprocess
from ssh_utils import SSHClientWrapper
from vbox_cli import clone_vm, get_all_vm_names, import_ova
from vbox_truenas import TrueNASVirtualBoxMachine

TRUENAS_TEST_OVA: t.Optional[str] = os.environ.get("TRUENAS_TEST_OVA")
TRUENAS_VM_NAME_PREFIX: str = "truenas-provider-tests"
THIS_DIR: str = os.path.dirname(__file__)

logger = logging.getLogger()

class SSHClientWrapperKwargs(t.TypedDict):
    host: str
    port: int
    username: str
    password: str

class TemporaryDirectorySentinel(tempfile.TemporaryDirectory):

    def __getitem__(self, *args, **kwargs) -> t.Any:
        raise Exception("Called .__getitem__() on TemporaryDirectorySentinel - serious error")

    def __setitem__(self, *args, **kwargs) -> None:
        raise Exception("Called .__setitem__() on TemporaryDirectorySentinel - serious error")

    def cleanup(self) -> None:
        raise Exception("Called .cleanup() on TemporaryDirectorySentinel - serious error")


class TrueNasVboxTests(unittest.TestCase):
    logging.basicConfig(level=logging.INFO)
    # Flag if the VM/etc is broken, nothing more should be run
    borked: t.ClassVar[bool] = False
    vm_name: str = "FAKE - overwritten in setUp()"
    vm: TrueNASVirtualBoxMachine = TrueNASVirtualBoxMachine("FAKE")
    temp_dir: tempfile.TemporaryDirectory = TemporaryDirectorySentinel()
    api_key: t.Optional[str] = None

    ###########################################################################
    ############################ TEST SETUP ###################################

    @classmethod
    def setUpClass(cls) -> None:
        """Set up resources for tests

        Includes
        - VM instance
        """

        # Ensure we have the base VM image
        if BASE_VM_NAME not in get_all_vm_names():
            if TRUENAS_TEST_OVA is None:
                raise Exception(f"VM image {BASE_VM_NAME} not found, please set TRUENAS_TEST_OVA env var")
            import_ova(TRUENAS_TEST_OVA, BASE_VM_NAME)

        # Create temporary directory
        cls.temp_dir = tempfile.TemporaryDirectory()
        print(f"Using temp_dir: {cls.temp_dir.name}")
        cls.vm_name = f"{TRUENAS_VM_NAME_PREFIX}-{cls.__name__}"
        cls.vm = TrueNASVirtualBoxMachine(cls.vm_name)

        cls.vm.create_via_clone(BASE_VM_NAME)

        # If anything goes wrong on the rest, we'll want to destroy VM, because tearDownClass won't be called
        try:
            cls.custom_setup_pre_start()
            cls.vm.start()
            cls.wait_for_ssh()
            cls.create_api_key()

            cls.custom_setup_post_start()
            cls.write_main_tf()
            # Ensure we can init/plan, otherwise definitely screwed
            cls.tf_init()
            cls.tf_plan()
            cls.tf_apply()

        except Exception as e:
            cls.vm.stop_hard()
            cls.vm.destroy()
            raise

    @classmethod
    def custom_setup_pre_start(cls) -> None:
        """Custom actions before starting VM (usually VM modifications)"""

    @classmethod
    def custom_setup_post_start(cls) -> None:
        """Custom actions after starting VM (usually API modifications)"""

    def setUp(self) -> None:
        if self.borked:
            raise self.skipTest("VM borked, further tests would be invalid")

    def tearDown(self) -> None:
        try:
            # Delete all test-created .tf files
            for fname in os.listdir(self.temp_dir.name):
                if fname.endswith(".tf") and fname not in ["main.tf"]:
                    os.remove(os.path.join(self.temp_dir.name, fname))

            self.tf_destroy()
        except Exception as e:
            self.__class__.borked = True
            raise Exception("terraform destroy failed, setting VM as borked")

    @classmethod
    def tearDownClass(cls) -> None:
        cls.temp_dir.cleanup()
        cls.vm.stop_hard()
        cls.vm.destroy()

    ############################ TEST SETUP ###################################
    ###########################################################################

    ###########################################################################
    ################################ API ######################################

    @classmethod
    def create_api_key(cls) -> None:
        logger.info(f"Attempting to login over SSH (local port {cls.vm.ssh_port}) to create API key")
        start = time.time()
        with cls.vm_ssh_client() as ssh:
            cmd = 'midclt call api_key.create \'{"name":"api godkey", "allowlist": [{"method": "*", "resource": "*"}]}\''
            _, ssh_stdout, ssh_stderr = ssh.exec_command(cmd)
            #logger.info("OUT", ssh_stdout.readlines())
            for line in ssh_stdout.readlines():
                if line.startswith("{"):
                    j = json.loads(line)
                    cls.api_key = j["key"]
                    return
            logger.info("Couldn't find key, STDOUT:")
            for line in ssh_stderr.readlines():
                sys.stdout.write(line)
            sys.stdout.flush()
            raise Exception("Could not create API key, please see logged stderr")

    @classmethod
    def get_api_key(cls) -> str:
        assert cls.api_key is not None
        return cls.api_key

    @classmethod
    def api_headers(cls) -> t.Mapping[str, str]:
        return {"Authorization": f"Bearer {cls.get_api_key()}"}

    @classmethod
    def api_post(cls, route: str, data: str) -> t.Mapping[str, t.Any]:
        assert route.startswith("/")
        URL = f"http://localhost:{cls.vm.truenas_api_port}/api/v2.0{route}"
        ret = requests.post(URL, headers=cls.api_headers(), data=data)
        ret.raise_for_status()
        return ret.json()

    @classmethod
    def api_get(cls, route: str) -> t.Mapping[str, t.Any]:
        assert route.startswith("/")
        URL = f"http://localhost:{cls.vm.truenas_api_port}/api/v2.0{route}"
        ret = requests.get(URL, headers=cls.api_headers())
        ret.raise_for_status()
        return ret.json()

    ################################ API ######################################
    ###########################################################################

    ###########################################################################
    ############################# TERRAFORM ###################################

    @classmethod
    def add_tf_file(cls, fname: str, content: str) -> None:
        assert fname.endswith(".tf")
        target_path = os.path.join(cls.temp_dir.name, fname)
        assert not os.path.exists(target_path)
        with open(target_path, "w+") as f:
            f.write(content)

    @classmethod
    def write_main_tf(cls, printall: bool = False) -> None:
        # Create main.tf
        with open(os.path.join(THIS_DIR, "res/main.template.tf")) as f:
            main_tf = f.read()
        for k, v in {
            "<<<API_KEY>>>": cls.get_api_key(),
            "<<<API_PORT>>>": str(cls.vm.truenas_api_port),
        }.items():
            main_tf = main_tf.replace(k, v)

        if printall:
            logger.info("======================================")
            logger.info("========== Writing main.tf ===========")
            logger.info(main_tf)
            logger.info("========== Writing main.tf ===========")
            logger.info("======================================")
        cls.add_tf_file("main.tf", main_tf)

    @classmethod
    def tf_run(cls, tf_cmd: t.List[str]) -> None:
        cmd = [
            "terraform",
            f"-chdir={cls.temp_dir.name}",
        ]
        run_subprocess(
            cmd + tf_cmd,
            f"$ {' '.join(cmd + tf_cmd)}",
        )

    @classmethod
    def tf_init(cls) -> None:
        cls.tf_run(["init"])

    @classmethod
    def tf_plan(cls) -> None:
        cls.tf_run(["plan"])

    @classmethod
    def tf_apply(cls) -> None:
        cls.tf_run(["apply", "-auto-approve"])

    @classmethod
    def tf_destroy(cls) -> None:
        cls.tf_run(["destroy", "-auto-approve"])

    ############################# TERRAFORM ###################################
    ###########################################################################

    @classmethod
    def vm_ssh_client(cls, ssh_timeout=5, **kwargs: t.Any) -> SSHClientWrapper:
        return SSHClientWrapper(
            host='localhost',
            port=cls.vm.get_ssh_port(),
            username="admin",
            password="terraform",
            timeout=ssh_timeout,
            **kwargs,
        )

    @classmethod
    def wait_for_ssh(cls) -> None:
        # Just ensure we can connect, don't do anything
        with cls.vm_ssh_client(ssh_retries=20):
            pass
