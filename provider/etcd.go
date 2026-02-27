package etcd

import (
        "context"
        "encoding/json"
        "fmt"
        "strings"
        "time"

        "github.com/caddyserver/caddy/v2"
        "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
        "github.com/libdns/libdns"
        clientv3 "go.etcd.io/etcd/client/v3"
)

func init() {
        caddy.RegisterModule(Provider{})
}

type Provider struct {
        Endpoints []string `json:"endpoints"`
        Prefix    string   `json:"prefix"`
        client    *clientv3.Client
}

func (Provider) CaddyModule() caddy.ModuleInfo {
        return caddy.ModuleInfo{
                ID:  "dns.providers.etcd",
                New: func() caddy.Module { return new(Provider) },
        }
}

func (p *Provider) Provision(ctx caddy.Context) error {
        if len(p.Endpoints) == 0 {
                p.Endpoints = []string{"http://localhost:2379"}
        }
        if p.Prefix == "" {
                p.Prefix = "/skydns"
        }
        p.initClient()
        return nil
}

func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
        d.Next()
        for nesting := d.Nesting(); d.NextBlock(nesting); {
                switch d.Val() {
                case "endpoints":
                        p.Endpoints = d.RemainingArgs()
                case "prefix":
                        if !d.NextArg() {
                                return d.ArgErr()
                        }
                        p.Prefix = d.Val()
                }
        }
        return nil
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
        if p.client == nil {
                p.initClient()
        }

        for i := 0; i < len(records); i++ {
                rec := records[i].RR()
                key := p.dnstoKey(rec.Name + "." + zone)
                val, _ := json.Marshal(map[string]string{"host": rec.Data})

                _, err := p.client.Put(ctx, key, string(val))
                if err != nil {
                        return nil, err
                }
                records[i] = rec.RR()
        }
        return records, nil
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
        if p.client == nil {
                p.initClient()
        }
        for i := 0; i < len(records); i++ {
                rec := records[i].RR()
                p.client.Delete(ctx, p.dnstoKey(rec.Name+"."+zone))
        }
        return records, nil
}

func (p *Provider) dnstoKey(dns string) string {
        parts := strings.Split(strings.Trim(dns, "."), ".")
        for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
                parts[i], parts[j] = parts[j], parts[i]
        }
        return p.Prefix + "/" + strings.Join(parts, "/")
}

func (p *Provider) initClient() {
        var err error
        p.client, err = clientv3.New(clientv3.Config{
                Endpoints:   p.Endpoints,
                DialTimeout: 5 * time.Second,
        })
        if err != nil {
                fmt.Printf("Failed to connect to etcd: %v\n", err)
        }
}

func (p *Provider) ProviderName() string {
        return "etcd"
}

func (p *Provider) FindTxtRecord(domain string) (string, error) {
        if p.client == nil {
                p.initClient()
        }

        key := p.dnstoKey(domain)
        resp, err := p.client.Get(context.Background(), key)
        if err != nil {
                return "", err
        }

        if len(resp.Kvs) == 0 {
                return "", fmt.Errorf("no TXT record found for %s", domain)
        }

        var record map[string]string
        if err := json.Unmarshal(resp.Kvs[0].Value, &record); err != nil {
                return "", err
        }

        return record["host"], nil
}
