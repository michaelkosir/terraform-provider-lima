# Lima Instance Resource Examples

This directory contains examples for using the `lima_instance` resource.

## Basic Usage

The simplest way to create a Lima instance with the default Ubuntu template:

```terraform
resource "lima_instance" "default" {
  name = "default"
}
```

## Available Templates

Lima provides several built-in templates. List them with:

```bash
limactl create --list-templates
```

## After Creation

Once an instance is created, you can:

```bash
limactl shell <instance-name>
```
