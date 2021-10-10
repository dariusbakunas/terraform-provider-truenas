<a href="https://terraform.io">
    <img src="https://raw.githubusercontent.com/hashicorp/terraform-website/master/content/source/assets/images/logo-terraform-main.svg" alt="Terraform logo" title="Terraform" align="right" height="50" />
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

In order to run a particular Acceptance test, export the variable TESTARGS. For example

```bash
export TESTARGS="-run TestAccDataSourceTruenasCronjob_basic"
```

**Note:** Acceptance tests create/destroy real resources, while they are named using `tf-acc-test-` testing prefix, take some caution.

### Debugging

Run your debugger (eg. [delve](https://github.com/go-delve/delve)), and pass it the provider binary as the command to run, specifying whatever flags, environment variables, or other input is necessary to start the provider in debug mode:

```bash
make build-debug
dlv exec --listen=:54526 --headless ./terraform-provider-truenas -- --debug
```

Note: IntelliJ may need additional flag `--api-version=2`

Connect your debugger (whether it's your IDE or the debugger client) to the debugger server. Example launch configuration for VSCode:

```json
{
    "apiVersion": 1,
    "name": "Debug",
    "type": "go",
    "request": "attach",
    "mode": "remote",
    "port": 54526, 
    "host": "127.0.0.1",
    "showLog": true
    //"trace": "verbose",            
}
```

Have it continue execution and it will print output like the following to stdout:

```bash
Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:

        TF_REATTACH_PROVIDERS='{"dariusbakunas/truenas":{"Protocol":"grpc","Pid":30101,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/mq/00hw97gj08323ybqfm763plr0000gn/T/plugin900766792"}}}'
```

Copy the line starting with `TF_REATTACH_PROVIDERS` from your provider's output. Either export it, or prefix every Terraform command with it. Run Terraform as usual. Any breakpoints you have set will halt execution and show you the current variable values.