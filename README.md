<a href="https://terraform.io">
    <img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" alt="Terraform logo" title="Terraform" align="right" height="50" />
</a>

# Terraform Provider for TrueNAS

Still in early development, check `examples` folder for working examples

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

Run `make install` to install provider binary under `~/.terraform.d/plugins/dariusbakunas/providers/truenas/${VERSION}/${OS_ARCH}`.

After it is installed, terraform should be able to detect it during `terraform init` phase.

### Testing

To run unit tests, simply run:

```bash
make test
```

To run acceptance tests, make sure `TRUENAS_API_KEY` and `TRUENAS_BASE_URL` environment variables are set and execute:

```bash
make testacc
```

**Note:** Acceptance tests create/destroy real resources, while they are named using `tf-test-` testing prefix, take some caution. 