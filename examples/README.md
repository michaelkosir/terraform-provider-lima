# Examples

This directory contains examples for the Lima Terraform provider.

## Quick Start

1. **Install Lima**:

   ```bash
   brew install lima  # macOS
   ```

2. **Configure the provider**:

   ```terraform
   terraform {
     required_providers {
       lima = {
         source = "michaelkosir/lima"
       }
     }
   }

   provider "lima" {}
   ```

3. **Create a Lima instance**:

   ```terraform
   resource "lima_instance" "example" {
     name   = "my-vm"
     cpus   = 2
     memory = 4
   }
   ```

4. **Apply the configuration**:
   ```bash
   terraform init
   terraform apply
   ```
