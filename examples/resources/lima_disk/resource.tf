resource "lima_disk" "ml_models" {
  name = "models"
  size = 50
}

resource "lima_instance" "docker" {
  name     = "docker"
  template = "docker"

  mount_none = true

  disks {
    name        = lima_disk.ml_models.name
    mount_point = "/mnt/models"
  }
}
