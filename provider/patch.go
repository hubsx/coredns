package provider

import (
	"encoding/json"
	"fmt"
	"github.com/libdns/libdns"
	"strings"
)

func (p *Provider) buildEtcdKey(zone string, record libdns.Record) string {
	// Example: Construct etcd key based on zone and record name
	// e.g., "/skydns/com/example/www/A"
	if record.Type == "TXT" {
		return fmt.Sprintf("%s/%s/%s", p.Prefix, reverseDomain(zone), record.Name)
	}
	return fmt.Sprintf("%s/%s/%s/%s", p.Prefix, reverseDomain(zone), record.Name, record.Type)
}

func (p *Provider) buildEtcdValue(record libdns.Record) (string, error) {

	skyRecord := SkyDNSRecord{
		Host: record.Value,
		TTL:  int(record.TTL.Seconds()),
		// Add any other fields that are necessary for SkyDNS
	}
	if record.Type == "TXT" {
		skyRecord = SkyDNSRecord{
			Text: record.Value,
			TTL:  int(record.TTL.Seconds()),
			// Add any other fields that are necessary for SkyDNS
		}
	}

	jsonValue, err := json.Marshal(skyRecord)
	if err != nil {
		return "", err
	}

	return string(jsonValue), nil
}

func reverseDomain(domain string) string {
	parts := strings.Split(domain, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, "/")
}
