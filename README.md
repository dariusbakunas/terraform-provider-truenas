<a href="https://terraform.io">
    <img src="https://raw.githubusercontent.com/hashicorp/terraform-website/master/content/source/assets/images/logo-terraform-main.svg" alt="Terraform logo" title="Terraform" align="right" height="50" />
</a>

# Terraform Provider for TrueNAS

Check `examples` folder for working examples

## Documentation

Full documentation is available on the Terraform website:

https://registry.terraform.io/providers/jdella/truenas/latest/docs

To generate provider documentation:

```bash
go generate
```

## Development

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.15.0+ (to run acceptance tests)
- [Go](https://golang.org/doc/install) 1.15.8+ (to build the provider plugin)

### Quick Start

First, clone the repository:

```bash
git clone git@github.com:jdella/terraform-provider-truenas.git
```

To compile the provider, run make build. This will build the provider and put the provider binary in the current directory.

```bash
make build
```

To test new binary, create `.terraformrc` in your home folder, with contents:

```terraform
provider_installation {
	dev_overrides {
    	"registry.terraform.io/jdella/terraform" = "<local path to cloned provider repo>"
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
│  - jdella/terraform in <...>/terraform-provider-terraform
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

        TF_REATTACH_PROVIDERS='{"jdella/truenas":{"Protocol":"grpc","Pid":30101,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/mq/00hw97gj08323ybqfm763plr0000gn/T/plugin900766792"}}}'
```

Copy the line starting with `TF_REATTACH_PROVIDERS` from your provider's output. Either export it, or prefix every Terraform command with it. Run Terraform as usual. Any breakpoints you have set will halt execution and show you the current variable values.

### VirtualBox Python Testing Suite

To avoid risks with testing directly on existing TrueNAS instances, a python-based virtualbox testing suite is provided.

#### Dependencies

To use this suite, you must have installed
- VirtualBox
- python3, with packages:
  - mypy
  - paramiko
  - requests
  - scp

#### One-time setup

First, a pre-installed VM must be created:
```
$ python3 tests/create_base_test_vm.py
```
The script will give detailed (albeit simple) instructions that the user must do manually. This is because there is no good automated installer available for TrueNAS.

Once complete, a new VM `truenas-test-base` will exist, which will be cloned by test classes.

#### Running the tests

The tests can be run easily through pytest:
```
$ python3 -m pytest
```
NOTE: The tests will take at least a few minutes **per test class**, so be patient. This is because each test class needs a fresh VM, and the VM itself takes about a minute to fully start up.

WARNING: There appear to be some intermittent bugs, for example they sometimes fail with:
```
E           Exception: Failed to find base dataset-pool dataset
```
This is likely due to poor exception handling and/or retrying. Re-running the tests usually will overcome this issue.

#### How the tests work

The tests work as follows:
- The test class sets up a clone of the base VM
  - Clones the base VM
  - Optionally runs some test-specific VirtualBox setup commands
  - Starts the VM
  - Waits for SSH to become available
  - Uses SSH to login and create an API key
  - Optionally runs some test-specific TrueNAS API setup commands
  - Creates a temp dir for terraform files
  - Writes `main.tf` and validates that `terraform init/plan/apply` all work
- Each test runs:
  - Creates .tf files
  - Runs `terraform apply`
  - Validates results using TrueNAS API
  - On failure, marks VM (for class) as borked, as it may not have cleaned up correctly
  - On tearDown, runs `terraform destroy`
- The test class tears everything down
  - Deletes the temp dir where the .tf files live
  - Runs a hard-stop on the VM
  - Destroys the VM along with any attached drives

#### Code Quality

The code is linted using the `mypy` python package:
```
$ python3 -m mypy .
Success: no issues found in 9 source files
```
