# Terraform Provider for Lima

A [Terraform](https://www.terraform.io) provider for managing [Lima](https://github.com/lima-vm/lima) VM instances. Lima launches Linux virtual machines with automatic file sharing and port forwarding (similar to WSL2), and containerd.

> **Note**: This provider was built as a learning exercise and interacts with `limactl` locally rather than some external API.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)
- [Lima](https://github.com/lima-vm/lima) installed and `limactl` in PATH

## Installing Lima

### macOS

```shell
brew install lima
```

### Linux

See the [Lima installation guide](https://lima-vm.io/docs/installation/) for your distribution.

## Using the Provider

```terraform
terraform {
  required_providers {
    lima = {
      source = "michaelkosir/lima"
    }
  }
}

provider "lima" {} # No configuration required

resource "lima_disk" "ml_models" {
  name = "models"
  size = 20
}

resource "lima_instance" "dev" {
  name     = "dev"
  template = "docker"

  mount_none = true

  disks {
    name        = lima_disk.ml_models.name
    mount_point = "/mnt/models"
  }
}
```

## Building The Provider

1. Build the provider using the Make `install` command:

```shell
make install
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine.

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create real Lima instances and require `limactl` to be installed.

```shell
make testacc
```

## Documentation

Documentation is available in the `docs/` directory and in the `examples/` directory.

## License

This provider is released under the MPL-2.0 License. See [LICENSE](LICENSE) for details.
