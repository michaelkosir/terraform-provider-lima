# Lima Disk Resource

This example demonstrates how to create a Lima disk using the `lima_disk` resource.

## Basic Usage

```terraform
resource "lima_disk" "data" {
  name = "data-disk"
  size = 50
}
```

## With Custom Format

```terraform
resource "lima_disk" "raw_disk" {
  name   = "storage"
  size   = 100
  format = "raw"
}
```

## Notes

- The `size` attribute is required and should be specified in GiB (as a number).
- The `format` attribute defaults to "qcow2" if not specified.
- Disks can be imported using `terraform import lima_disk.<name> <disk-name>`.
