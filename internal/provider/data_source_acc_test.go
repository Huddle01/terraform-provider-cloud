package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDataSourceRegions verifies that the regions data source returns
// at least one region with a non-empty name.
func TestAccDataSourceRegions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "huddle_cloud_regions" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.huddle_cloud_regions.test", "regions.%"),
				),
			},
		},
	})
}

// TestAccDataSourceFlavors verifies that the flavors data source returns
// a non-empty list with well-formed flavor entries.
func TestAccDataSourceFlavors(t *testing.T) {
	region := testAccRegion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFlavorsConfig(region),
				Check: resource.ComposeTestCheckFunc(
					// At least one flavor should be present.
					resource.TestCheckResourceAttrSet("data.huddle_cloud_flavors.test", "flavors.#"),
					resource.TestCheckResourceAttrSet("data.huddle_cloud_flavors.test", "flavors.0.id"),
					resource.TestCheckResourceAttrSet("data.huddle_cloud_flavors.test", "flavors.0.name"),
					resource.TestCheckResourceAttrSet("data.huddle_cloud_flavors.test", "flavors.0.vcpus"),
					resource.TestCheckResourceAttrSet("data.huddle_cloud_flavors.test", "flavors.0.ram"),
				),
			},
		},
	})
}

// TestAccDataSourceImages verifies that the images data source returns
// at least one image group with an image ID and version.
func TestAccDataSourceImages(t *testing.T) {
	region := testAccRegion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceImagesConfig(region),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.huddle_cloud_images.test", "image_groups.#"),
					resource.TestCheckResourceAttrSet("data.huddle_cloud_images.test", "image_groups.0.distro"),
					resource.TestCheckResourceAttrSet("data.huddle_cloud_images.test", "image_groups.0.versions.#"),
				),
			},
		},
	})
}

// TestAccDataSourceNetworks verifies that the networks data source can list
// networks (including externally-managed ones) in the configured region.
func TestAccDataSourceNetworks(t *testing.T) {
	region := testAccRegion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNetworksConfig(region),
				Check: resource.ComposeTestCheckFunc(
					// At minimum the networks list attribute must be present.
					resource.TestCheckResourceAttrSet("data.huddle_cloud_networks.test", "networks.#"),
				),
			},
		},
	})
}

// TestAccDataSourceInstance looks up an existing instance by ID.
// Requires HUDDLE_FLAVOR_ID and HUDDLE_IMAGE_ID to create a transient instance
// whose ID is then looked up via the data source.
func TestAccDataSourceInstance(t *testing.T) {
	region := testAccRegion()
	flavorID := testAccFlavorID(t)
	imageID := testAccImageID(t)
	pubKey := testAccSSHPublicKey()
	keyName := accName("key")
	vmName := accName("vm")
	sgName := testAccSGName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceInstanceConfig(keyName, vmName, region, flavorID, imageID, pubKey, sgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.huddle_cloud_instance.test", "id",
						"huddle_cloud_instance.test", "id",
					),
					resource.TestCheckResourceAttr("data.huddle_cloud_instance.test", "name", vmName),
					resource.TestCheckResourceAttr("data.huddle_cloud_instance.test", "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet("data.huddle_cloud_instance.test", "vcpus"),
					resource.TestCheckResourceAttrSet("data.huddle_cloud_instance.test", "ram"),
				),
			},
		},
	})
}

func testAccDataSourceFlavorsConfig(region string) string {
	return fmt.Sprintf(`
data "huddle_cloud_flavors" "test" {
  region = %q
}
`, region)
}

func testAccDataSourceImagesConfig(region string) string {
	return fmt.Sprintf(`
data "huddle_cloud_images" "test" {
  region = %q
}
`, region)
}

func testAccDataSourceNetworksConfig(region string) string {
	return fmt.Sprintf(`
data "huddle_cloud_networks" "test" {
  region = %q
}
`, region)
}

func testAccDataSourceInstanceConfig(keyName, vmName, region, flavorID, imageID, pubKey, sgName string) string {
	return fmt.Sprintf(`
resource "huddle_cloud_keypair" "test" {
  name       = %q
  public_key = %q
}

resource "huddle_cloud_instance" "test" {
  name                 = %q
  region               = %q
  flavor_id            = %q
  image_id             = %q
  boot_disk_size       = 20
  key_names            = [huddle_cloud_keypair.test.name]
  security_group_names = [%q]
  assign_public_ip     = true
}

data "huddle_cloud_instance" "test" {
  id     = huddle_cloud_instance.test.id
  region = %q
}
`, keyName, pubKey, vmName, region, flavorID, imageID, sgName, region)
}
