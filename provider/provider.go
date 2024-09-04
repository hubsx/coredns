// Package easydns implements a DNS record management client compatible
// with the libdns interfaces for EasyDNS.
// See https://cp.easydns.com/manage/security/ to manage Token and Key information
// for your account.
package provider

import (
	"context"
	"fmt"
	"github.com/libdns/libdns"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
	"time"
)

// Provider facilitates DNS record manipulation with EasyDNS.
type Provider struct {
	// EasyDNS API Token (required)
	APIToken string `json:"api_token,omitempty"`
	// EasyDNS API Key (required)
	APIKey string `json:"api_key,omitempty"`
	// EasyDNS API URL (defaults to https://rest.easydns.net)
	APIUrl string `json:"api_url,omitempty"`

	Prefix string
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	log.Println("Get Records for zone:", zone)

	// Remove trailing dot from zone if present
	zone = strings.TrimSuffix(zone, ".")
	client, err := connectEtcd([]string{p.APIUrl})
	if err != nil {
		return nil, err
	}
	kvs, err := getSkyDNSData(client, p.Prefix)
	if err != nil {

		return nil, err
	}
	return convertToLibdnsRecords(kvs)
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	log.Println("Append Record(s) to zone:", zone)
	// Remove trailing dot from zone if present
	zone = strings.TrimSuffix(zone, ".")
	fmt.Println("append records")
	for _, v := range records {
		fmt.Println(v)
	}

	var appendedRecords []libdns.Record
	client, err := connectEtcd([]string{p.APIUrl})
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		key := p.buildEtcdKey(zone, record)
		value, err := p.buildEtcdValue(record)
		fmt.Printf("use set %s=%s\n", key, value)
		if err != nil {
			return nil, err
		}

		// Use a transaction to ensure atomic operation
		txn := client.Txn(ctx)
		txn = txn.If(clientv3.Compare(clientv3.Version(key), "=", 0)).
			Then(clientv3.OpPut(key, value)).
			Else(clientv3.OpGet(key))

		// Commit the transaction
		txnResp, err := txn.Commit()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if txnResp.Succeeded {
			appendedRecords = append(appendedRecords, record)
			fmt.Println("success")
			time.Sleep(5 * time.Second)
		} else {
			log.Printf("Record %s already exists, skipping", record.Name)
		}
	}

	return appendedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	log.Println("Update Record(s) in zone:", zone)
	fmt.Println("set records")

	for _, v := range records {
		fmt.Println(v)
	}
	// Remove trailing dot from zone if present
	zone = strings.TrimSuffix(zone, ".")

	var setRecords []libdns.Record
	client, err := connectEtcd([]string{p.APIUrl})
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		key := p.buildEtcdKey(zone, record)
		value, err := p.buildEtcdValue(record)
		if err != nil {
			return nil, err
		}

		// Store the record in etcd (this will overwrite existing records)
		_, err = client.Put(ctx, key, value)
		if err != nil {
			return nil, err
		}

		setRecords = append(setRecords, record)
	}

	return setRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	log.Println("Delete Record(s) from zone:", zone)
	fmt.Println("delete records")

	for _, v := range records {
		fmt.Println(v)
	}
	// Remove trailing dot from zone if present
	zone = strings.TrimSuffix(zone, ".")

	var deletedRecords []libdns.Record
	client, err := connectEtcd([]string{p.APIUrl})
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		//key := p.buildEtcdKey(zone, record)

		// Delete the record from etcd
		//_, err := client.Delete(ctx, key)
		//if err != nil {
		//	return nil, err
		//}

		deletedRecords = append(deletedRecords, record)
	}

	return deletedRecords, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)

func (p *Provider) getApiUrl() string {
	if p.APIUrl != "" {
		return p.APIUrl
	}
	return "https://rest.easydns.net"
}
