package proxy

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/matthiasng/acme-dns-proxy/dns"
	"go.uber.org/zap"
)

// Proxy handles incoming request and calls the DNS provider API.
type Proxy struct {
	Logger      *zap.Logger
	Provider    dns.Provider
	AccessRules AccessRules
}

// Handle validates and authenticates a request. If everything is fine, the configured DNS provider API gets called.
func (p *Proxy) Handle(req *Request) error {
	log := p.Logger.With(
		zap.String("reqID", uuid.New().String()),
		zap.String("action", req.Action),
	).Sugar()

	log.Infow("new request",
		"from", req.Remote.Addr,
		"fqdn", req.Challenge.FQDN,
		"key_auth_value", req.Challenge.EncodedKeyAuth,
	)

	if err := validateRequest(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	rule := p.AccessRules.Search(req.Challenge.FQDN)
	if rule == nil {
		err := fmt.Errorf(`no access rule for "%s" found`, req.Challenge.FQDN)
		log.Debug(err)
		return err
	}

	log.Debugw("access rule matched",
		"pattern", rule.Pattern,
	)

	if !rule.CheckAuth(req.AuthToken) {
		log.Debug("access denied")
		return fmt.Errorf("access denied")
	}

	err := callProvider(p.Provider, req, log)
	if err != nil {
		log.Debug(err)
		return err
	}

	log.Debug("DNS API called")
	return nil
}

func validateRequest(req *Request) error {
	if req.Action != "present" && req.Action != "cleanup" {
		return errors.New("invalid request: unknown action")
	}

	if len(req.Challenge.FQDN) == 0 {
		return errors.New("invalid request: fqdn not set")
	}

	if len(req.Challenge.EncodedKeyAuth) == 0 {
		return errors.New("invalid request: key auth value not set")
	}

	return nil
}

func callProvider(provider dns.Provider, req *Request, log *zap.SugaredLogger) error {
	if req.Action == "present" {
		return callPresent(provider, req, log)
	}

	return callCleanup(provider, req, log)
}

func callPresent(provider dns.Provider, req *Request, log *zap.SugaredLogger) error {
	err := provider.Present(req.Challenge)
	if err != nil {
		return fmt.Errorf(`add record failed: %w`, err)
	}

	return nil
}

func callCleanup(provider dns.Provider, req *Request, log *zap.SugaredLogger) error {
	err := provider.CleanUp(req.Challenge)
	if err != nil {
		return fmt.Errorf(`delete record failed: %w`, err)
	}

	return nil
}
