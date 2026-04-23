package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccInstance_basic provisions a real instance, verifies key attributes,
// tests import, and confirms the instance is deleted on destroy.
//
// Requires env vars: HUDDLE_FLAVOR_NAME, HUDDLE_IMAGE_NAME
// Optional: HUDDLE_SSH_PUBLIC_KEY
func TestAccInstance_basic(t *testing.T) {
	name := accName("vm")
	region := testAccRegion()
	flavorName := testAccFlavorName(t)
	imageName := testAccImageName(t)
	pubKey := testAccSSHPublicKey()
	keyName := accName("key")
	sgName := testAccSGName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroyed(region),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig(name, region, flavorName, imageName, keyName, pubKey, sgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("huddle_cloud_instance.test", "name", name),
					resource.TestCheckResourceAttr("huddle_cloud_instance.test", "region", region),
					resource.TestCheckResourceAttr("huddle_cloud_instance.test", "flavor_name", flavorName),
					resource.TestCheckResourceAttr("huddle_cloud_instance.test", "image_name", imageName),
					resource.TestCheckResourceAttr("huddle_cloud_instance.test", "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet("huddle_cloud_instance.test", "id"),
					resource.TestCheckResourceAttrSet("huddle_cloud_instance.test", "public_ipv4"),
					resource.TestCheckResourceAttrSet("huddle_cloud_instance.test", "vcpus"),
					resource.TestCheckResourceAttrSet("huddle_cloud_instance.test", "ram"),
					resource.TestCheckResourceAttrSet("huddle_cloud_instance.test", "created_at"),
				),
			},
			// Import by instance ID.
			{
				ResourceName:      "huddle_cloud_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
				// These fields are write-only inputs not returned by the GET API.
				ImportStateVerifyIgnore: []string{
					"power_state",
					"assign_public_ip",
					"boot_disk_size",
					"key_names",
					"security_group_names",
				},
			},
		},
	})
}

func testAccCheckInstanceDestroyed(region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccClient()
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "huddle_cloud_instance" {
				continue
			}
			id := rs.Primary.ID
			var out instanceResponseEnvelope
			err := client.get(context.Background(), "/instances/"+id, queryWithRegion(region), &out)
			if err == nil {
				return fmt.Errorf("instance %q still exists after destroy", id)
			}
			if !isNotFound(err) {
				return fmt.Errorf("unexpected error checking instance %q: %w", id, err)
			}
		}
		return nil
	}
}

func testAccInstanceConfig(name, region, flavorName, imageName, keyName, pubKey, sgName string) string {
	return fmt.Sprintf(`
resource "huddle_cloud_keypair" "test" {
  name       = %q
  public_key = %q
}

resource "huddle_cloud_instance" "test" {
  name                 = %q
  region               = %q
  flavor_name          = %q
  image_name           = %q
  boot_disk_size       = 20
  key_names            = [huddle_cloud_keypair.test.name]
  security_group_names = [%q]
  assign_public_ip     = true
}
`, keyName, pubKey, name, region, flavorName, imageName, sgName)
}
