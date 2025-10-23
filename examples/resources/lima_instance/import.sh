

# Import existing Lima instances into Terraform

# If you have an existing Lima instance created outside of Terraform,
# you can import it into your Terraform state.

# First, create the resource configuration without applying:
resource "lima_instance" "existing" {
  name = "my-existing-instance"
  # Add other known configuration attributes
  # Note: Some attributes may not be fully recoverable during import
}

# Then import the instance:
# terraform import lima_instance.existing my-existing-instance

# After import, run terraform plan to see if there are any differences
# between your configuration and the actual instance state.
