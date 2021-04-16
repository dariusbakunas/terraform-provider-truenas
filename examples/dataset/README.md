# Creating TrueNAS dataset

Adjust variables in `example.tfvars` first and then run:

```bash
terraform apply -var-file="example.tfvars"
```

Alternatively, specify them through environment:

```bash
TF_VAR_dataset_name=truenas_provider_test TF_VAR_dataset_pool=Tank terraform apply
```

Cleanup (WARNING: this will delete newly created dataset):

```bash
terraform destroy
```