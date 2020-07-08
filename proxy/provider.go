package proxy

import (
	"fmt"

	"github.com/matthiasng/acme-dns-proxy/config"
	"github.com/matthiasng/dns-provider-api/provider"
)

// Provider calls the DNS provider API
type Provider interface {
	Present(domain, token, fqdn, value string) error
	CleanUp(domain, token, fqdn, value string) error
}

type LegoProvider struct {
	impl provider.Provider
}

func NewLegoProviderFromConfig(providerCfg *config.Provider) (Provider, error) {
	if len(providerCfg.Type) == 0 {
		return nil, fmt.Errorf("error initializing provider: provider type not specified")
	}

	impl, err := provider.New(providerCfg.Type, providerCfg.Variables)
	if err != nil {
		return nil, fmt.Errorf("error initializing provider: %w", err)
	}

	return &LegoProvider{
		impl: impl,
	}, nil
}

func (p *LegoProvider) Present(domain, token, fqdn, value string) error {
	return p.impl.Present(domain, token, fqdn, value)
}

func (p *LegoProvider) CleanUp(domain, token, fqdn, value string) error {
	return p.impl.CleanUp(domain, token, fqdn, value)
}
