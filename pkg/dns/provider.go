package dns

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/juju/errors"
	"github.com/libdns/cloudflare"
	"github.com/libdns/libdns"
	"github.com/matthiasng/libdnsfactory"

	"github.com/hpidcock/acme-dns-proxy/pkg/config"
	"github.com/hpidcock/acme-dns-proxy/pkg/dns01"
)

// Provider calls the DNS provider API
type Provider interface {
	Present(ctx context.Context, c Challenge) error
	Cleanup(ctx context.Context, c Challenge) error
	Underlying() interface {
		libdns.RecordGetter
		libdns.RecordAppender
		libdns.RecordSetter
		libdns.RecordDeleter
	}
}

// NewProviderFromConfig creates a new provider from a config.Provider instance
func NewProviderFromConfig(cfg *config.Provider, resolver ZoneResolver) (Provider, error) {
	if len(cfg.Type) == 0 {
		return nil, fmt.Errorf("error initializing provider: provider type not specified")
	}

	p := &provider{
		pendingRecords: map[string]pendingRecord{},
	}

	switch cfg.Type {
	case "cloudflare":
		var c struct {
			APIToken string `hcl:"api_token"`
		}
		err := gohcl.DecodeBody(cfg.Remain, nil, &c)
		if err != nil {
			return nil, errors.Trace(err)
		}
		p.provider = &cloudflare.Provider{
			APIToken: c.APIToken,
		}
	default:
		return nil, fmt.Errorf("unsupported provider %q", cfg.Type)
	}

	return p, nil
}

// NewProvider creates a new provider
func NewProvider(p libdnsfactory.Provider, resolver ZoneResolver) (Provider, error) {
	return &provider{
		provider:       p,
		zoneResolver:   resolver,
		pendingRecords: map[string]pendingRecord{},
	}, nil
}

type provider struct {
	provider     libdnsfactory.Provider
	zoneResolver ZoneResolver

	pendingRecords      map[string]pendingRecord
	pendingRecordsMutex sync.Mutex
}

type pendingRecord struct {
	recordID string
	zone     string
}

func (l *provider) Present(ctx context.Context, c Challenge) error {
	zone, err := l.zoneResolver(c.FQDN)
	if err != nil {
		return fmt.Errorf("failed to append record: %w", err)
	}

	recordName := dns01.UnFQDN(dns01.RemoveZoneFromFQDN(dns01.TXTRecordName(c.FQDN), zone))
	record := libdns.Record{
		Type:  "TXT",
		Name:  recordName,
		Value: c.EncodedKeyAuth,
		TTL:   60 * time.Second, // TODO: config
	}

	records, err := l.provider.AppendRecords(ctx, zone, []libdns.Record{record})
	if err != nil {
		return fmt.Errorf("failed to append record: %w", err)
	}

	l.pendingRecordsMutex.Lock()
	defer l.pendingRecordsMutex.Unlock()
	l.pendingRecords[c.EncodedKeyAuth] = pendingRecord{
		recordID: records[0].ID,
		zone:     zone,
	}
	return nil
}

func (l *provider) Cleanup(ctx context.Context, c Challenge) error {
	pendingRecord, err := l.popPendingRecordID(c.EncodedKeyAuth)
	if err != nil {
		return fmt.Errorf("failed to cleanup record: %w", err)
	}

	_, err = l.provider.DeleteRecords(ctx, pendingRecord.zone, []libdns.Record{{ID: pendingRecord.recordID}})
	if err != nil {
		return fmt.Errorf("failed to cleanup record: %w", err)
	}

	return nil
}

func (l *provider) Underlying() interface {
	libdns.RecordGetter
	libdns.RecordAppender
	libdns.RecordSetter
	libdns.RecordDeleter
} {
	return l.provider
}

func (l *provider) popPendingRecordID(encKeyAuth string) (*pendingRecord, error) {
	l.pendingRecordsMutex.Lock()
	defer l.pendingRecordsMutex.Unlock()

	if pr, ok := l.pendingRecords[encKeyAuth]; ok {
		delete(l.pendingRecords, encKeyAuth)
		return &pr, nil
	}

	return nil, fmt.Errorf("no pending record found [encKeyAuth: %s]", encKeyAuth)
}
