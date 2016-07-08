package cloudstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

func TestAccCloudStackPrivateGateway_basic(t *testing.T) {
	var gateway cloudstack.PrivateGateway

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackPrivateGatewayDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
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

//func TestAccCloudStackIPAddress_vpc(t *testing.T) {
//	var ipaddr cloudstack.PublicIpAddress
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		Providers:    testAccProviders,
//		CheckDestroy: testAccCheckCloudStackIPAddressDestroy,
//		Steps: []resource.TestStep{
//			resource.TestStep{
//				Config: testAccCloudStackIPAddress_vpc,
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckCloudStackIPAddressExists(
//						"cloudstack_ipaddress.foo", &ipaddr),
//				),
//			},
//		},
//	})
//}

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
		pip, _, err := cs.Address.GetPrivateGatewayByID(rs.Primary.ID)

		if err != nil {
			return err
		}

		if pip.Id != rs.Primary.ID {
			return fmt.Errorf("Private Gateway not found")
		}

		*gateway = *pip

		return nil
	}
}

func testAccCheckCloudStackPrivateGatewayAttributes(
	gateway *cloudstack.PrivateGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if gateway.Vpcid != CLOUDSTACK_VPC_1 {
			return fmt.Errorf("Bad VPC ID: %s", gateway.Vpcid)
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

		gateway, _, err := cs.Address.GetPrivateGatewayByID(rs.Primary.ID)
		if err == nil && gateway.id != "" {
			return fmt.Errorf("Private gateway %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

var testAccCloudStackPrivateGateway_basic = fmt.Sprintf(`
resource "cloudstack_vpc" "foobar" {
  name = "terraform-vpc"
  cidr = "%s"
  vpc_offering = "%s"
  zone = "%s"
}

resource "cloudstack_gateway" "foo" {
	gateway = "%s"
	ipaddress = "%s"
	netmask - "%s"
	network_offering = "%s"
  vpc_id = "${cloudstack_vpc.foobar.id}"
}`,
	CLOUDSTACK_VPC_CIDR_1,
	CLOUDSTACK_VPC_OFFERING,
	CLOUDSTACK_ZONE,
	CLOUDSTACK_PRIVGW_GATEWAY,
	CLOUDSTACK_PRIVGW_IPADDRESS,
	CLOUDSTACK_PRIVGW_NETMASK,
	CLOUDSTACK_PRIVGW_NETWORK_OFFERING)
