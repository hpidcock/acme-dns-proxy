package dns

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libdns/libdns"
	"github.com/matthiasng/acme-dns-proxy/config"
	"github.com/matthiasng/acme-dns-proxy/dns01"
	"github.com/matthiasng/libdnsfactory"
)

// Provider calls the DNS provider API
type Provider interface {
	Present(c Challenge) error
	CleanUp(c Challenge) error
}

// NewProviderFromConfig creates a new provider from a config.Provider instance
func NewProviderFromConfig(providerCfg *config.Provider, resolver ZoneResolver) (Provider, error) {
	if len(providerCfg.Type) == 0 {
		return nil, fmt.Errorf("error initializing provider: provider type not specified")
	}

	p, err := libdnsfactory.NewProvider(providerCfg.Type, providerCfg.Variables)
	if err != nil {
		return nil, fmt.Errorf("error initializing provider: %w", err)
	}

	return &libdnsProvider{
		provider:       p,
		pendingRecords: map[string]pendingRecord{},
	}, nil
}

// NewProvider creates a new provider
func NewProvider(p libdnsfactory.Provider, resolver ZoneResolver) (Provider, error) {
	return &libdnsProvider{
		provider:       p,
		zoneResolver:   resolver,
		pendingRecords: map[string]pendingRecord{},
	}, nil
}

// libdnsProvider implements the Provider interface with github.com/libdns/libdns
type libdnsProvider struct {
	provider     libdnsfactory.Provider
	zoneResolver ZoneResolver

	pendingRecords      map[string]pendingRecord
	pendingRecordsMutex sync.Mutex
}

type pendingRecord struct {
	recordID string
	zone     string
}

func (l *libdnsProvider) Present(c Challenge) error {
	// #todo test requests
	// 1. present mn1.com (1)
	// 2. present mn2.com
	// 3. present mn1.com (2)
	// 4. cleanup mn2.com
	// 6. cleanup mn1.com (1)
	// 6. cleanup mn1.com (2)

	zone, err := l.zoneResolver(c.FQDN)
	if err != nil {
		return fmt.Errorf("Failed to append record: %w", err)
	}

	// #todo log("found authorizive zone", "for" name, "zone", zone)

	recordName := dns01.UnFqdn(dns01.RemoveZoneFromFqdn(dns01.TXTRecordName(c.FQDN), zone))
	record := libdns.Record{
		Type:  "TXT",
		Name:  recordName,
		Value: c.EncodedKeyAuth,
		TTL:   60 * time.Second, // #todo config
	}

	records, err := l.provider.AppendRecords(context.TODO(), zone, []libdns.Record{record})
	if err != nil {
		return fmt.Errorf("Failed to append record: %w", err)
	}

	l.pendingRecordsMutex.Lock()
	defer l.pendingRecordsMutex.Unlock()

	l.pendingRecords[c.EncodedKeyAuth] = pendingRecord{
		recordID: records[0].ID,
		zone:     zone,
	}
	return nil
}

func (l *libdnsProvider) CleanUp(c Challenge) error {
	pendingRecord, err := l.popPendingRecordID(c.EncodedKeyAuth)
	if err != nil {
		return fmt.Errorf("failed to cleanup record: %w", err)
	}

	_, err = l.provider.DeleteRecords(context.TODO(), pendingRecord.zone, []libdns.Record{{ID: pendingRecord.recordID}})
	if err != nil {
		return fmt.Errorf("failed to cleanup record: %w", err)
	}

	return nil
}

func (l *libdnsProvider) popPendingRecordID(encKeyAuth string) (*pendingRecord, error) {
	l.pendingRecordsMutex.Lock()
	defer l.pendingRecordsMutex.Unlock()

	for k, v := range l.pendingRecords {
		if k == encKeyAuth {
			delete(l.pendingRecords, k)
			return &v, nil
		}
	}

	return nil, fmt.Errorf("no pending record found [encKeyAuth: %s]", encKeyAuth)
}
