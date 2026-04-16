package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccVolume_deleteOnDestroy verifies that when delete_on_destroy=true,
// terraform destroy actually removes the volume from the cloud.
func TestAccVolume_deleteOnDestroy(t *testing.T) {
	name := accName("vol")
	region := testAccRegion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroyed(region),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeConfig(name, region, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("huddle_cloud_volume.test", "name", name),
					resource.TestCheckResourceAttr("huddle_cloud_volume.test", "size", "10"),
					resource.TestCheckResourceAttr("huddle_cloud_volume.test", "region", region),
					resource.TestCheckResourceAttr("huddle_cloud_volume.test", "delete_on_destroy", "true"),
					resource.TestCheckResourceAttrSet("huddle_cloud_volume.test", "id"),
					resource.TestCheckResourceAttr("huddle_cloud_volume.test", "status", "available"),
					resource.TestCheckResourceAttr("huddle_cloud_volume.test", "bootable", "false"),
				),
			},
			// Note: ImportState is intentionally omitted here. After import,
			// delete_on_destroy resets to false (it's a lifecycle-only field not
			// stored by the API), which would cause the final terraform destroy to
			// retain the volume — defeating the purpose of this test. Import
			// behaviour is covered separately in TestAccVolume_import.
		},
	})
}

// TestAccVolume_retainOnDestroy verifies that when delete_on_destroy=false (default),
// terraform destroy removes the resource from state but leaves the volume in the cloud.
// The test manually cleans up the orphaned volume in CheckDestroy.
func TestAccVolume_retainOnDestroy(t *testing.T) {
	name := accName("vol-retain")
	region := testAccRegion()
	var volumeID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			client := testAccClient()
			// Volume should still exist because delete_on_destroy=false.
			var out volumeDetailEnvelope
			err := client.get(context.Background(), "/volumes/"+volumeID, queryWithRegion(region), &out)
			if err != nil {
				return fmt.Errorf("volume %q should still exist after destroy (delete_on_destroy=false), got error: %w", volumeID, err)
			}
			// Manually clean up the retained volume so we don't leave orphans.
			_ = client.delete(context.Background(), "/volumes/"+volumeID, queryWithRegion(region), generateIdempotencyKey(), nil)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeConfig(name, region, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("huddle_cloud_volume.test", "name", name),
					resource.TestCheckResourceAttr("huddle_cloud_volume.test", "delete_on_destroy", "false"),
					resource.TestCheckResourceAttrSet("huddle_cloud_volume.test", "id"),
					// Capture the volume ID for use in CheckDestroy.
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["huddle_cloud_volume.test"]
						if rs == nil {
							return fmt.Errorf("huddle_cloud_volume.test not found in state")
						}
						volumeID = rs.Primary.ID
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckVolumeDestroyed(region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccClient()
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "huddle_cloud_volume" {
				continue
			}
			id := rs.Primary.ID
			var out volumeDetailEnvelope
			err := client.get(context.Background(), "/volumes/"+id, queryWithRegion(region), &out)
			if err == nil {
				return fmt.Errorf("volume %q still exists after destroy", id)
			}
			if !isNotFound(err) {
				return fmt.Errorf("unexpected error checking volume %q: %w", id, err)
			}
		}
		return nil
	}
}

// TestAccVolume_import verifies that a volume can be imported and that all
// stored attributes round-trip correctly. delete_on_destroy is excluded from
// verify because it is a lifecycle-only field not tracked by the API; after
// import it resets to false which is the safe default.
func TestAccVolume_import(t *testing.T) {
	name := accName("vol-import")
	region := testAccRegion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroyed(region),
		Steps: []resource.TestStep{
			{
				// Use delete_on_destroy=true so the volume is cleaned up after the test.
				Config: testAccVolumeConfig(name, region, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("huddle_cloud_volume.test", "id"),
				),
			},
			{
				ResourceName:            "huddle_cloud_volume.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_on_destroy"},
			},
			// Re-apply the original config to restore delete_on_destroy=true in
			// state so that the final terraform destroy actually deletes the volume.
			{
				Config: testAccVolumeConfig(name, region, true),
			},
		},
	})
}

func testAccVolumeConfig(name, region string, deleteOnDestroy bool) string {
	return fmt.Sprintf(`
resource "huddle_cloud_volume" "test" {
  name              = %q
  size              = 10
  region            = %q
  delete_on_destroy = %t
}
`, name, region, deleteOnDestroy)
}
