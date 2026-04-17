package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccVolumeAttachment_basic provisions an instance and a volume, attaches them,
// verifies the attachment attributes, tests import, and exercises the full destroy
// sequence — which also validates the detach-before-delete race condition fix.
//
// Requires env vars: HUDDLE_FLAVOR_ID, HUDDLE_IMAGE_ID
func TestAccVolumeAttachment_basic(t *testing.T) {
	region := testAccRegion()
	flavorID := testAccFlavorID(t)
	imageID := testAccImageID(t)
	pubKey := testAccSSHPublicKey()
	keyName := accName("key")
	vmName := accName("vm")
	volName := accName("vol")
	sgName := testAccSGName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckAttachmentAndVolumeDestroyed(region),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeAttachmentConfig(keyName, vmName, volName, region, flavorID, imageID, pubKey, sgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("huddle_cloud_volume_attachment.test", "id"),
					resource.TestCheckResourceAttrSet("huddle_cloud_volume_attachment.test", "volume_id"),
					resource.TestCheckResourceAttrSet("huddle_cloud_volume_attachment.test", "instance_id"),
					resource.TestCheckResourceAttr("huddle_cloud_volume_attachment.test", "region", region),
					resource.TestCheckResourceAttrSet("huddle_cloud_volume_attachment.test", "device"),
					// Note: huddle_cloud_volume.test.status is intentionally not checked here.
					// Terraform does not re-read the volume resource after the attachment
					// changes OpenStack's underlying status to "in-use", so the state
					// would reflect the stale "available" value from creation time.
				),
			},
			// Import the attachment by "volume_id/instance_id".
			{
				ResourceName:      "huddle_cloud_volume_attachment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccCheckAttachmentAndVolumeDestroyed verifies both the attachment is gone
// (volume is no longer in-use) and the volume itself is deleted.
// The volume uses delete_on_destroy=true in the test config.
func testAccCheckAttachmentAndVolumeDestroyed(region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccClient()
		for _, rs := range s.RootModule().Resources {
			switch rs.Type {
			case "huddle_cloud_volume":
				id := rs.Primary.ID
				var out volumeDetailEnvelope
				err := client.get(context.Background(), "/volumes/"+id, queryWithRegion(region), &out)
				if err == nil {
					return fmt.Errorf("volume %q still exists after destroy", id)
				}
				if !isNotFound(err) {
					return fmt.Errorf("unexpected error checking volume %q: %w", id, err)
				}
			case "huddle_cloud_instance":
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
		}
		return nil
	}
}

func testAccVolumeAttachmentConfig(keyName, vmName, volName, region, flavorID, imageID, pubKey, sgName string) string {
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

resource "huddle_cloud_volume" "test" {
  name              = %q
  size              = 10
  region            = %q
  delete_on_destroy = true
}

resource "huddle_cloud_volume_attachment" "test" {
  volume_id   = huddle_cloud_volume.test.id
  instance_id = huddle_cloud_instance.test.id
  region      = %q
}
`, keyName, pubKey,
		vmName, region, flavorID, imageID, sgName,
		volName, region,
		region)
}
