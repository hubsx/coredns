package provider

import (
	"context"
	"fmt"
	"testing"
)

var pp = &Provider{
	APIUrl: "http://hg.hubx.dev:2379",
	Prefix: "/skydns",
}

func TestProvider_GetRecords(t *testing.T) {
	data, err := pp.GetRecords(context.Background(), "hubx.dev")
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func TestProvider_Etcd(t *testing.T) {
	client, err := connectEtcd([]string{pp.APIUrl})
	if err != nil {
		panic(err)
	}
	resp, err := client.Get(context.Background(), "/skydns/dev/hubx/cd")
	if err != nil {
		panic(err)
	}
	for _, n := range resp.Kvs {
		fmt.Println(n)
	}

}
