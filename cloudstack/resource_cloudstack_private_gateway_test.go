package cloudstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/xanzy/go-cloudstack/v2/cloudstack"
)

func TestAccCloudStackPrivateGateway_basic(t *testing.T) {
	var gateway cloudstack.PrivateGateway

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackPrivateGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackPrivateGateway_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackPrivateGatewayExists(
						"cloudstack_private_gateway.foo", &gateway),
					testAccCheckCloudStackPrivateGatewayAttributes(&gateway),
				),
			},
		},
	})
}

func TestAccCloudStackPrivateGateway_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackPrivateGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackPrivateGateway_basic,
			},

			{
				ResourceName:      "cloudstack_private_gateway.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCloudStackPrivateGatewayExists(
	n string, gateway *cloudstack.PrivateGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Private Gateway ID is set")
		}

		cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
		pgw, _, err := cs.VPC.GetPrivateGatewayByID(rs.Primary.ID)

		if err != nil {
			return err
		}

		if pgw.Id != rs.Primary.ID {
			return fmt.Errorf("Private Gateway not found")
		}

		*gateway = *pgw

		return nil
	}
}

func testAccCheckCloudStackPrivateGatewayAttributes(
	gateway *cloudstack.PrivateGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if gateway.Gateway != "10.1.1.254" {
			return fmt.Errorf("Bad Gateway: %s", gateway.Gateway)
		}

		if gateway.Ipaddress != "192.168.0.1" {
			return fmt.Errorf("Bad Gateway: %s", gateway.Ipaddress)
		}

		if gateway.Netmask != "255.255.255.0" {
			return fmt.Errorf("Bad Gateway: %s", gateway.Netmask)
		}

		return nil
	}
}

func testAccCheckCloudStackPrivateGatewayDestroy(s *terraform.State) error {
	cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstack_private_gateway" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No private gateway ID is set")
		}

		gateway, _, err := cs.VPC.GetPrivateGatewayByID(rs.Primary.ID)
		if err == nil && gateway.Id != "" {
			return fmt.Errorf("Private gateway %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccCloudStackPrivateGateway_basic = `
resource "cloudstack_vpc" "foo" {
  name = "terraform-vpc"
  cidr = "10.0.0.0/8"
  vpc_offering = "Default VPC offering"
  zone = "Sandbox-simulator"
}

resource "cloudstack_network_acl" "foo" {
  name = "terraform-acl"
  vpc_id = "${cloudstack_vpc.foo.id}"
}

resource "cloudstack_private_gateway" "foo" {
  gateway = "10.1.1.254"
  ip_address = "192.168.0.1"
  netmask = "255.255.255.0"
  vlan = "1"
  vpc_id = "${cloudstack_vpc.foo.id}"
  acl_id = "${cloudstack_network_acl.foo.id}"
}`
