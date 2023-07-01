import json
import logging
import time
import typing as t

from test_base import TrueNasVboxTests


logger = logging.getLogger()


class TestDataset(TrueNasVboxTests):
    POOL = "dataset-pool"

    @classmethod
    def custom_setup_pre_start(cls) -> None:
        super().custom_setup_pre_start()
        cls.vm.create_and_attach_hdd(hdd_name="dataset1", size_mb=8192, port=2)

    @classmethod
    def custom_setup_post_start(cls) -> None:
        super().custom_setup_post_start()
        # Create pool "dataset-pool"
        data = json.dumps({
            "name": cls.POOL,
            "topology": {
                # Currently only have 1 hdd in base image, so "sdb" is the new one
                "data": [{"type": "STRIPE", "disks": ["sdb"]}],
            }
        })
        cls.api_post("/pool", data=data)

        # poll until initial dataset available (can take >1min)
        logger.info(f"Created default pool {cls.POOL}, waiting for corresponding dataset creation...")
        for i in range(40):
            time.sleep(5)
            if cls.api_get("/pool/dataset"):
                logger.info(f"Dataset for pool {cls.POOL} detected!")
                # a dataset exists, we can start tests
                break
        else:
            raise Exception(f"Failed to find base {cls.POOL} dataset")

    def test_create_dataset(self) -> None:
        target_ds_name = "test_dataset1"
        target_ds_id = f"{self.POOL}/{target_ds_name}"

        # Create and apply terraform
        self.add_tf_file("dataset.tf", f"""
resource "truenas_dataset" "{target_ds_name}" {{
  name = "{target_ds_name}"
  pool = "{self.POOL}"
}}
""")
        self.tf_apply()

        # Use API to confirm dataset exists
        j = self.api_get(f"/pool/dataset/id/{self.POOL}")
        assert "children" in j, f"Expected 'children' key, found keys: {j.keys()}"
        child_ds_ids = [x["id"] for x in j["children"]]
        assert target_ds_id in child_ds_ids, f"Expected id {target_ds_id} not in child DS ids: {child_ds_ids}"
