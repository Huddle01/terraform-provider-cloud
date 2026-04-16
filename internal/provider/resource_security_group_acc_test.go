package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSecurityGroup_basic(t *testing.T) {
	name := accName("sg")
	region := testAccRegion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig(name, region),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("huddle_cloud_security_group.test", "name", name),
					resource.TestCheckResourceAttr("huddle_cloud_security_group.test", "region", region),
					resource.TestCheckResourceAttrSet("huddle_cloud_security_group.test", "id"),
					resource.TestCheckResourceAttrSet("huddle_cloud_security_group.test", "created_at"),
					// The SSH rule resource should have the correct attributes.
					resource.TestCheckResourceAttr("huddle_cloud_security_group_rule.ssh", "direction", "ingress"),
					resource.TestCheckResourceAttr("huddle_cloud_security_group_rule.ssh", "protocol", "tcp"),
					resource.TestCheckResourceAttr("huddle_cloud_security_group_rule.ssh", "port_range_min", "22"),
					resource.TestCheckResourceAttr("huddle_cloud_security_group_rule.ssh", "port_range_max", "22"),
					resource.TestCheckResourceAttr("huddle_cloud_security_group_rule.ssh", "remote_ip_prefix", "0.0.0.0/0"),
					resource.TestCheckResourceAttrSet("huddle_cloud_security_group_rule.ssh", "id"),
				),
			},
			// Import the security group by ID.
			{
				ResourceName:            "huddle_cloud_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
		},
	})
}

func testAccCheckSecurityGroupDestroyed(s *terraform.State) error {
	client := testAccClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "huddle_cloud_security_group" {
			continue
		}
		id := rs.Primary.ID
		region := rs.Primary.Attributes["region"]
		var out securityGroupDetailEnvelope
		err := client.get(context.Background(), "/security-groups/"+id, queryWithRegion(region), &out)
		if err == nil {
			return fmt.Errorf("security group %q still exists after destroy", id)
		}
		if !isNotFound(err) {
			return fmt.Errorf("unexpected error checking security group %q: %w", id, err)
		}
	}
	return nil
}

func testAccSecurityGroupConfig(name, region string) string {
	return fmt.Sprintf(`
resource "huddle_cloud_security_group" "test" {
  name   = %q
  region = %q
}

resource "huddle_cloud_security_group_rule" "ssh" {
  security_group_id = huddle_cloud_security_group.test.id
  direction         = "ingress"
  ether_type        = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  region            = %q
}
`, name, region, region)
}
