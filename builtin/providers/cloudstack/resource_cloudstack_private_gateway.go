package cloudstack

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

func resourceCloudStackPrivateGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudStackPrivateGatewayCreate,
		Read:   resourceCloudStackPrivateGatewayRead,
		Update: resourceCloudStackPrivateGatewayUpdate,
		Delete: resourceCloudStackPrivateGatewayDelete,

		Schema: map[string]*schema.Schema{
			"gateway": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Computed: true,
				ForceNew: true,
			},

			"ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Computed: true,
				ForceNew: true,
			},

			"netmask": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Computed: true,
				ForceNew: true,
			},

			"vlan": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"acl_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"network_offering": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},


			"physical_network_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"source_nat_supported": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

		},
	}
}

func resourceCloudStackPrivateGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	ipaddress := d.Get("ip_address").(string)

	// Create a new parameter struct
	p := cs.VPC.NewCreatePrivateGatewayParams(
		d.Get("gateway").(string),
		ipaddress,
		d.Get("netmask").(string),
		d.Get("vlan").(string),
		d.Get("vpc_id").(string)
	)

	// Retrieve the network_offering ID
	networkofferingid, e := retrieveID(cs, "network_offering", d.Get("network_offering").(string))
	if e != nil {
		return e.Error()
	}
	p.SetNetworkofferingid(networkofferingid)

	// Since we're in a VPC, check if we want to assiciate an ACL list
	aclid, ok := d.GetOk("acl_id")
	if ok {
		// Set the acl ID
		p.SetAclid(aclid.(string))
	}

	// Check if the phyical_network ID is given
	physicalnetworkid, ok := d.GetOk("physical_network_id")
	if ok {
		// Set the phyical_network ID
		p.SetPhysicalnetworkid(physicalnetworkid.(string))
	}

	// Check if the sourcenatsupported is set
	sourcenatsupported, ok := d.GetOk("source_nat_supported")
	if ok {
		// Set sourcenatsupported
		p.SetSourcenatsupported(sourcenatsupported.(bool))
	}

	// Create the new network
	r, err := cs.VPC.CreatePrivateGateway(p)
	if err != nil {
		return fmt.Errorf("Error creating private gateway %s: %s", ipaddress, err)
	}

	d.SetId(r.Id)

	return resourceCloudStackPrivateGatewayRead(d, meta)
}

func resourceCloudStackPrivateGatewayRead(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	// Get the virtual machine details
	n, count, err := cs.VPC.GetPrivateGatewayByID(
		d.Id()
	)
	if err != nil {
		if count == 0 {
			log.Printf(
				"[DEBUG] PrivateGateway %s does no longer exist", d.Get("ipaddress").(string))
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("gateway", n.Gateway)
	d.Set("ipaddress", n.Ipaddress)
	d.Set("netmask", n.Netmask)
	d.Set("gateway", n.Gateway)
	d.Set("vpc_id", n.Vpcid)
	d.Set("acl_id", n.Aclid)
	setValueOrID(d, "network_offering", n.Networkofferingname, n.Networkofferingid)
	d.Set("physical_network_id", n.Physicalnetworkid)
	d.Set("source_nat_supported", n.Sourcenatsupported)

	return nil
}

func resourceCloudStackPrivateGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	// Replace the ACL if the ID has changed
	if d.HasChange("acl_id") {
		aclid, ok := d.GetOk("acl_id")
		if !ok {
			return fmt.Errorf("Replacing the ACL requires a valid ACL ID")
		}

		p := cs.NetworkACL.NewReplaceNetworkACLListParams(aclid.(string))
		p.SetGatewayid(d.Id())

		_, err := cs.NetworkACL.ReplaceNetworkACLList(p)
		if err != nil {
			return fmt.Errorf("Error replacing ACL: %s", err)
		}
	}

	return resourceCloudStackNetworkRead(d, meta)
}

func resourceCloudStackPrivateGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	// Create a new parameter struct
	p := cs.VPC.NewDeletePrivateGatewayParams(d.Id())

	// Delete the network
	_, err := cs.VPC.DeletePrivateGateway(p)
	if err != nil {
		// This is a very poor way to be told the ID does no longer exist :(
		if strings.Contains(err.Error(), fmt.Sprintf(
			"Invalid parameter id value=%s due to incorrect long value format, "+
				"or entity does not exist", d.Id())) {
			return nil
		}

		return fmt.Errorf("Error deleting private gateway %s: %s", d.Get("ipaddress").(string), err)
	}
	return nil
}
