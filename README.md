<a href="https://terraform.io">
    <img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" alt="Terraform logo" title="Terraform" align="right" height="50" />
</a>

# Terraform Provider for TrueNAS

Still in early development, check `examples` folder for working examples

## Documentation

Full documentation is available on the Terraform website:

https://registry.terraform.io/providers/dariusbakunas/truenas/latest/docs

## Development

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.15.0+ (to run acceptance tests)
- [Go](https://golang.org/doc/install) 1.15.8+ (to build the provider plugin)

### Quick Start

First, clone the repository:

```bash
git clone git@github.com:dariusbakunas/terraform-provider-truenas.git
```

To compile the provider, run make build. This will build the provider and put the provider binary in the current directory.

```bash
make build
```

To test new binary, create `.terraformrc` in your home folder, with contents:

```terraform
provider_installation {
	dev_overrides {
    	"registry.terraform.io/dariusbakunas/terraform" = "<local path to cloned provider repo>"
  	}

  	direct {}
}
```

You will get a warning next time you run terraform:
```bash
╷
│ Warning: Provider development overrides are in effect
│ 
│ The following provider development overrides are set in the CLI configuration:
│  - dariusbakunas/terraform in <...>/terraform-provider-terraform
│ 
│ The behavior may therefore not match any released version of the provider and applying changes may cause the state to become incompatible with published releases.
╵
```
When done, remove `.terraformrc` or comment out `dev_overrides` section.
### Testing

To run unit tests, simply run:

```bash
make test
```

To run acceptance tests, make sure `TRUENAS_API_KEY` and `TRUENAS_BASE_URL` environment variables are set and execute:

```bash
make testacc
```

**Note:** Acceptance tests create/destroy real resources, while they are named using `tf-acc-test-` testing prefix, take some caution. 