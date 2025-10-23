package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLimaDiskResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLimaDiskResourceConfig("test-disk", 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lima_disk.test", "name", "test-disk"),
					resource.TestCheckResourceAttr("lima_disk.test", "size", "10"),
					resource.TestCheckResourceAttr("lima_disk.test", "format", "qcow2"),
					resource.TestCheckResourceAttr("lima_disk.test", "id", "test-disk"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "lima_disk.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Size and format cannot be imported, only the name/id
				ImportStateVerifyIgnore: []string{"size", "format"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccLimaDiskResourceWithFormat(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLimaDiskResourceConfigWithFormat("test-disk-raw", 20, "raw"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lima_disk.test", "name", "test-disk-raw"),
					resource.TestCheckResourceAttr("lima_disk.test", "size", "20"),
					resource.TestCheckResourceAttr("lima_disk.test", "format", "raw"),
				),
			},
		},
	})
}

func TestAccLimaDiskResourceResize(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with initial size
			{
				Config: testAccLimaDiskResourceConfig("test-disk-resize", 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lima_disk.test", "name", "test-disk-resize"),
					resource.TestCheckResourceAttr("lima_disk.test", "size", "10"),
				),
			},
			// Resize disk (increase size)
			{
				Config: testAccLimaDiskResourceConfig("test-disk-resize", 20),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lima_disk.test", "name", "test-disk-resize"),
					resource.TestCheckResourceAttr("lima_disk.test", "size", "20"),
				),
			},
		},
	})
}

func testAccLimaDiskResourceConfig(name string, size float64) string {
	return fmt.Sprintf(`
resource "lima_disk" "test" {
  name = %[1]q
  size = %[2]g
}
`, name, size)
}

func testAccLimaDiskResourceConfigWithFormat(name string, size float64, format string) string {
	return fmt.Sprintf(`
resource "lima_disk" "test" {
  name   = %[1]q
  size   = %[2]g
  format = %[3]q
}
`, name, size, format)
}
