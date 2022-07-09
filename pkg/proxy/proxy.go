package proxy

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"

	"github.com/hpidcock/acme-dns-proxy/pkg/dns"
)

// Proxy handles incoming request and calls the DNS provider API.
type Proxy struct {
	Log      *logrus.Logger
	Provider dns.Provider
	ACLs     ACLs
}

// Handle validates and authenticates a request. If everything is fine, the configured DNS provider API gets called.
func (p *Proxy) Handle(ctx context.Context, req *Request) error {
	log := p.Log.
		WithField("reqID", uuid.New().String()).
		WithField("action", req.Action)

	log.Info("request",
		"from", req.Remote.Address,
		"fqdn", req.Challenge.FQDN,
		"key_auth_value", req.Challenge.EncodedKeyAuth,
	)

	if err := req.Challenge.Validate(); err != nil {
		return errors.Trace(err)
	}

	rule, err := p.ACLs.Search(req.Challenge.FQDN)
	if err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	if !rule.CheckAuth(req.AuthToken) {
		return fmt.Errorf("access denied")
	}

	switch req.Action {
	case "present":
		err := p.Provider.Present(ctx, req.Challenge)
		if err != nil {
			return fmt.Errorf("add record failed: %w", err)
		}
	case "cleanup":
		err := p.Provider.Cleanup(ctx, req.Challenge)
		if err != nil {
			return fmt.Errorf("cleanup record failed: %w", err)
		}
	default:
		return fmt.Errorf("unknown action %q", req.Action)
	}

	return nil
}
