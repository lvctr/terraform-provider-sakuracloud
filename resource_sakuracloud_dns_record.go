package sakuracloud

import (
	"fmt"

	"bytes"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/yamamoto-febc/libsacloud/api"
	"github.com/yamamoto-febc/libsacloud/sacloud"
	"strings"
)

func resourceSakuraCloudDNSRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceSakuraCloudDNSRecordCreate,
		Read:   resourceSakuraCloudDNSRecordRead,
		Delete: resourceSakuraCloudDNSRecordDelete,

		Schema: map[string]*schema.Schema{
			"dns_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringInWord(sacloud.AllowDNSTypes()),
			},

			"value": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ttl": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3600,
				ForceNew: true,
			},

			"priority": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSakuraCloudDNSRecordCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	dnsID := d.Get("dns_id").(string)

	sakuraMutexKV.Lock(dnsID)
	defer sakuraMutexKV.Unlock(dnsID)

	dns, err := client.DNS.Read(dnsID)
	if err != nil {
		return fmt.Errorf("Couldn't find SakuraCloud DNS resource: %s", err)
	}

	record := expandDNSRecord(d)

	if r := findRecordMatch(record, &dns.Settings.DNS.ResourceRecordSets); r != nil {
		return fmt.Errorf("Failed to create SakuraCloud DNS resource:Duplicate DNS record: %v", record)
	}

	dns.AddRecord(record)
	dns, err = client.DNS.Update(dnsID, dns)
	if err != nil {
		return fmt.Errorf("Failed to create SakuraCloud DNSRecord resource: %s", err)
	}

	d.SetId(dnsRecordIDHash(dnsID, record))
	return resourceSakuraCloudDNSRecordRead(d, meta)
}

func resourceSakuraCloudDNSRecordRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	dns, err := client.DNS.Read(d.Get("dns_id").(string))
	if err != nil {
		return fmt.Errorf("Couldn't find SakuraCloud DNS resource: %s", err)
	}

	record := expandDNSRecord(d)
	if r := findRecordMatch(record, &dns.Settings.DNS.ResourceRecordSets); r == nil {
		return fmt.Errorf("Couldn't find SakuraCloud DNSRecord resource: %v", record)
	}

	d.Set("name", record.Name)
	d.Set("type", record.Type)
	d.Set("value", record.RData)
	d.Set("ttl", record.TTL)

	if record.Type == "MX" {
		// ex. record.RData = "10 example.com."
		values := strings.SplitN(record.RData, " ", 2)
		d.Set("value", values[1])
		d.Set("priority", values[0])
	}

	return nil
}

func resourceSakuraCloudDNSRecordDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	dnsID := d.Get("dns_id").(string)

	sakuraMutexKV.Lock(dnsID)
	defer sakuraMutexKV.Unlock(dnsID)

	dns, err := client.DNS.Read(dnsID)
	if err != nil {
		return fmt.Errorf("Couldn't find SakuraCloud DNS resource: %s", err)
	}

	record := expandDNSRecord(d)
	records := dns.Settings.DNS.ResourceRecordSets
	dns.ClearRecords()

	for _, r := range records {
		if !isSameDNSRecord(&r, record) {
			dns.AddRecord(&r)
		}
	}

	dns, err = client.DNS.Update(dnsID, dns)
	if err != nil {
		return fmt.Errorf("Failed to delete SakuraCloud DNSRecord resource: %s", err)
	}

	d.SetId("")
	return nil
}

func findRecordMatch(r *sacloud.DNSRecordSet, records *[]sacloud.DNSRecordSet) *sacloud.DNSRecordSet {
	for _, record := range *records {

		if isSameDNSRecord(r, &record) {
			return &record
		}
	}
	return nil
}
func isSameDNSRecord(r1 *sacloud.DNSRecordSet, r2 *sacloud.DNSRecordSet) bool {
	return r1.Name == r2.Name && r1.RData == r2.RData && r1.TTL == r2.TTL && r1.Type == r2.Type
}

func dnsRecordIDHash(dns_id string, r *sacloud.DNSRecordSet) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", dns_id))
	buf.WriteString(fmt.Sprintf("%s-", r.Type))
	buf.WriteString(fmt.Sprintf("%s-", r.RData))
	buf.WriteString(fmt.Sprintf("%d-", r.TTL))
	buf.WriteString(fmt.Sprintf("%s-", r.Name))

	return fmt.Sprintf("dnsrecord-%d", hashcode.String(buf.String()))
}

func expandDNSRecord(d *schema.ResourceData) *sacloud.DNSRecordSet {
	var dns = sacloud.DNS{}
	t := d.Get("type").(string)
	if t == "MX" {
		pr := 10
		if p, ok := d.GetOk("priority"); ok {
			pr = p.(int)
		}
		return dns.CreateNewMXRecord(
			d.Get("name").(string),
			d.Get("value").(string),
			d.Get("ttl").(int),
			pr)
	} else {
		return dns.CreateNewRecord(
			d.Get("name").(string),
			d.Get("type").(string),
			d.Get("value").(string),
			d.Get("ttl").(int))

	}
}
