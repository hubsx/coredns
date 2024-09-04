package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/libdns/libdns"
	mvccpb "go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strings"
	"time"
)

func getSkyDNSData(cli *clientv3.Client, prefix string) ([]*mvccpb.KeyValue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	return resp.Kvs, nil
}

func connectEtcd(endpoints []string) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil
}

type SkyDNSRecord struct {
	Host string `json:"host"`
	TTL  int    `json:"ttl"`
	Text string `json:"text"`
}

func convertToLibdnsRecords(kvs []*mvccpb.KeyValue) ([]libdns.Record, error) {
	var records []libdns.Record

	for _, kv := range kvs {
		var skyRecord SkyDNSRecord
		if kv.Value == nil || len(kv.Value) < 1 {
			continue
		}
		err := json.Unmarshal(kv.Value, &skyRecord)
		if err != nil {
			return nil, err
		}
		myKey := string(kv.Key)
		skyRecord.TTL = 1800
		//校验typer
		typer := ""
		val := ""
		if strings.HasSuffix(skyRecord.Host, ".dev") {
			typer = "CNAME"
			val = skyRecord.Host
		} else if skyRecord.Host != "" {
			typer = "A"
			val = skyRecord.Host

		} else if skyRecord.Text != "" {
			typer = "TXT"
			val = skyRecord.Text
		} else {
			continue
		}

		// Extract the DNS record type and name from the key or value
		record := libdns.Record{
			Type:  typer,                    // Example: You might need to determine the type from the data.
			Name:  extractDomainName(myKey), // This function should extract the name from the key.
			Value: val,
			TTL:   time.Duration(skyRecord.TTL) * time.Second,
		}
		records = append(records, record)
	}

	return records, nil
}

func extractDomainName(key string) string {
	// Implement your logic to extract domain name from the key.
	parts := strings.Split(strings.TrimPrefix(key, "/skydns/dev/hubx/"), "/")
	fmt.Println(parts[0])
	return parts[0] // Placeholder
}
