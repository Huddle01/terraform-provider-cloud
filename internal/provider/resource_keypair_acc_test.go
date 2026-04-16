package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKeypair_basic(t *testing.T) {
	name := accName("key")
	pubKey := testAccSSHPublicKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckKeypairDestroyed(name),
		Steps: []resource.TestStep{
			{
				Config: testAccKeypairConfig(name, pubKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("huddle_cloud_keypair.test", "name", name),
					resource.TestCheckResourceAttr("huddle_cloud_keypair.test", "public_key", pubKey),
					resource.TestCheckResourceAttrSet("huddle_cloud_keypair.test", "fingerprint"),
					resource.TestCheckResourceAttrSet("huddle_cloud_keypair.test", "api_id"),
					resource.TestCheckResourceAttrSet("huddle_cloud_keypair.test", "created_at"),
				),
			},
			// Verify import round-trips all stored attributes.
			{
				ResourceName:            "huddle_cloud_keypair.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
		},
	})
}

func testAccCheckKeypairDestroyed(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := testAccClient()
		var out keyPairDetailEnvelope
		err := client.get(context.Background(), "/keypairs/"+name, nil, &out)
		if err == nil {
			return fmt.Errorf("keypair %q still exists after destroy", name)
		}
		if isNotFound(err) {
			return nil
		}
		return fmt.Errorf("unexpected error checking keypair %q: %w", name, err)
	}
}

func testAccKeypairConfig(name, pubKey string) string {
	return fmt.Sprintf(`
resource "huddle_cloud_keypair" "test" {
  name       = %q
  public_key = %q
}
`, name, pubKey)
}
