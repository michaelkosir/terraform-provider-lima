# Basic Lima instance with default Ubuntu template
resource "lima_instance" "default" {
  name = "default"
}

# Lima instance with Docker template
resource "lima_instance" "docker" {
  name     = "docker"
  template = "docker"
}

# Lima instance with custom resources
resource "lima_instance" "custom" {
  name   = "custom"
  cpus   = 4
  memory = 8
  disk   = 50
}

# Lima instance with advanced configuration
resource "lima_instance" "advanced" {
  name       = "advanced"
  template   = "ubuntu"
  cpus       = 2
  memory     = 4
  disk       = 30
  vm_type    = "qemu"
  containerd = "user"

  mount = [
    "/Users/myuser/projects:w",
    "/tmp"
  ]

  mount_type     = "reverse-sshfs"
  mount_writable = false

  network = [
    "lima:shared"
  ]
}

# Lima instance with plain mode (minimal features)
resource "lima_instance" "plain" {
  name   = "plain"
  plain  = true
  cpus   = 2
  memory = 2
}

# Lima instance from remote URL (use with caution)
resource "lima_instance" "remote" {
  name     = "alpine"
  template = "https://raw.githubusercontent.com/lima-vm/lima/master/templates/alpine.yaml"
  cpus     = 2
  memory   = 2
}
