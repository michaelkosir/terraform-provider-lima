package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLimaInstanceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLimaInstanceResourceConfig("test-instance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lima_instance.test", "name", "test-instance"),
					resource.TestCheckResourceAttr("lima_instance.test", "id", "test-instance"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "lima_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore computed boolean attributes with defaults since they won't be in the imported state
				ImportStateVerifyIgnore: []string{"mount_inotify", "mount_none", "mount_writable", "plain", "rosetta", "video"},
			},
			// Update and Read testing - most changes force replacement
			{
				Config: testAccLimaInstanceResourceConfigWithCpus("test-instance", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lima_instance.test", "name", "test-instance"),
					resource.TestCheckResourceAttr("lima_instance.test", "cpus", "2"),
				),
			},
		},
	})
}

func TestAccLimaInstanceResourceWithTemplate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLimaInstanceResourceConfigWithTemplate("test-docker", "docker"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lima_instance.test", "name", "test-docker"),
					resource.TestCheckResourceAttr("lima_instance.test", "template", "docker"),
				),
			},
		},
	})
}

func TestAccLimaInstanceResourceInPlaceUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with initial resources
			{
				Config: testAccLimaInstanceResourceConfigWithResources("test-update", 2, 4.0, 100.0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lima_instance.test", "name", "test-update"),
					resource.TestCheckResourceAttr("lima_instance.test", "cpus", "2"),
					resource.TestCheckResourceAttr("lima_instance.test", "memory", "4"),
					resource.TestCheckResourceAttr("lima_instance.test", "disk", "100"),
				),
			},
			// Update resources in-place (should not force replacement)
			{
				Config: testAccLimaInstanceResourceConfigWithResources("test-update", 4, 8.0, 100.0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lima_instance.test", "name", "test-update"),
					resource.TestCheckResourceAttr("lima_instance.test", "cpus", "4"),
					resource.TestCheckResourceAttr("lima_instance.test", "memory", "8"),
					resource.TestCheckResourceAttr("lima_instance.test", "disk", "100"),
				),
			},
		},
	})
}

func TestAccLimaInstanceResourceWithDisks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create instance with additional disks
			{
				Config: testAccLimaInstanceResourceConfigWithDisks("test-disks"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lima_instance.test", "name", "test-disks"),
					resource.TestCheckResourceAttr("lima_disk.data", "name", "test-data-disk"),
					resource.TestCheckResourceAttr("lima_disk.data", "size", "10"),
					resource.TestCheckResourceAttr("lima_instance.test", "disks.0.name", "test-data-disk"),
					resource.TestCheckResourceAttr("lima_instance.test", "disks.0.mount_point", "/mnt/data"),
				),
			},
		},
	})
}

func testAccLimaInstanceResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "lima_instance" "test" {
  name = %[1]q
}
`, name)
}

func testAccLimaInstanceResourceConfigWithCpus(name string, cpus int) string {
	return fmt.Sprintf(`
resource "lima_instance" "test" {
  name = %[1]q
  cpus = %[2]d
}
`, name, cpus)
}

func testAccLimaInstanceResourceConfigWithTemplate(name string, template string) string {
	return fmt.Sprintf(`
resource "lima_instance" "test" {
  name     = %[1]q
  template = %[2]q
}
`, name, template)
}

func testAccLimaInstanceResourceConfigWithResources(name string, cpus int, memory float64, disk float64) string {
	return fmt.Sprintf(`
resource "lima_instance" "test" {
  name   = %[1]q
  cpus   = %[2]d
  memory = %[3]g
  disk   = %[4]g
}
`, name, cpus, memory, disk)
}

func testAccLimaInstanceResourceConfigWithDisks(name string) string {
	return fmt.Sprintf(`
resource "lima_disk" "data" {
  name = "test-data-disk"
  size = 10
}

resource "lima_instance" "test" {
  name       = %[1]q
  template   = "docker"
  mount_none = true

  disks {
    name        = lima_disk.data.name
    mount_point = "/mnt/data"
  }
}
`, name)
}
