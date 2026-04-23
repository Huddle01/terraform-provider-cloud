package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccNetwork_basic(t *testing.T) {
	if os.Getenv("HUDDLE_TEST_NETWORKS") == "" {
		t.Skip("set HUDDLE_TEST_NETWORKS=1 to run network tests (requires network creation support in the region)")
	}
	name := accName("net")
	region := testAccRegion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckNetworkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfig(name, region),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("huddle_cloud_network.test", "name", name),
					resource.TestCheckResourceAttr("huddle_cloud_network.test", "region", region),
					resource.TestCheckResourceAttrSet("huddle_cloud_network.test", "id"),
					resource.TestCheckResourceAttrSet("huddle_cloud_network.test", "status"),
					resource.TestCheckResourceAttrSet("huddle_cloud_network.test", "admin_state_up"),
				),
			},
			// Import by ID.
			{
				ResourceName:      "huddle_cloud_network.test",
				ImportState:       true,
				ImportStateVerify: true,
				// These optional create-time fields are not returned by the read API.
				ImportStateVerifyIgnore: []string{
					"description",
					"pool_cidr",
					"primary_subnet_cidr",
					"primary_subnet_size",
					"no_gateway",
					"enable_dhcp",
				},
			},
		},
	})
}

func testAccCheckNetworkDestroyed(s *terraform.State) error {
	client := testAccClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "huddle_cloud_network" {
			continue
		}
		id := rs.Primary.ID
		region := rs.Primary.Attributes["region"]
		q := queryWithRegion(region)
		q.Set("owned", "true")
		var out networkListEnvelope
		err := client.get(context.Background(), "/networks", q, &out)
		if err != nil {
			return fmt.Errorf("error listing networks during destroy check: %w", err)
		}
		for _, n := range out.Data.Networks {
			if n.ID == id {
				return fmt.Errorf("network %q still exists after destroy", id)
			}
		}
	}
	return nil
}

func testAccNetworkConfig(name, region string) string {
	return fmt.Sprintf(`
resource "huddle_cloud_network" "test" {
  name                = %q
  region              = %q
  pool_cidr           = "10.200.0.0/16"
  primary_subnet_cidr = "10.200.1.0/24"
}
`, name, region)
}
