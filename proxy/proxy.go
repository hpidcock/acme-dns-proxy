package proxy

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-acme/lego/challenge/dns01"

	"go.uber.org/zap"
)

type context struct {
	Log *zap.SugaredLogger
}

// Proxy handles incoming request and calls the DNS provider API.
type Proxy struct {
	Logger      *zap.Logger
	Provider    Provider
	AccessRules AccessRules
	counter     RequestIDCounter
}

// Handle handels a request.
func (p *Proxy) Handle(req *Request) error {
	log := p.Logger.With(
		zap.Uint64("reqID", p.counter.Next()),
		zap.String("action", req.Action),
	).Sugar()

	log.Infow("new request",
		"from", req.Client.RemoteAddr,
		"TxtFQDN", req.Payload.TxtFQDN,
		"TxtValue", req.Payload.TxtValue,
	)

	if err := validateRequest(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	cleanFqdn := strings.TrimPrefix(req.Payload.TxtFQDN, "_acme-challenge.")
	rule := p.AccessRules.Search(cleanFqdn)
	if rule == nil {
		err := fmt.Errorf(`no access rule for "%s" found`, cleanFqdn)
		log.Debug(err)
		return err
	}

	log.Debugw("access rule found",
		"matchedPattern", rule.Pattern,
	)

	if !rule.CheckAuth(req.Token) {
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
		return errors.New("invalid action")
	}

	if len(req.Payload.TxtFQDN) == 0 {
		return errors.New("TXT record: FQDN not set")
	}

	if len(req.Payload.TxtValue) == 0 {
		return errors.New("TXT record: value not set")
	}

	return nil
}

func callProvider(provider Provider, req *Request, log *zap.SugaredLogger) error {
	domain := dns01.ToFqdn(strings.TrimPrefix(req.Payload.TxtFQDN, "_acme-challenge."))
	token := req.Payload.TxtValue
	fqdn := req.Payload.TxtFQDN
	value := req.Payload.TxtValue

	log.Debugw("Call DNS API",
		"domain", domain,
		"token", token,
		"fqdn", fqdn,
		"value", value,
	)

	var err error
	if req.Action == "present" {
		err = provider.Present(domain, token, fqdn, value)
	} else {
		err = provider.CleanUp(domain, token, fqdn, value)
	}

	if err != nil {
		return fmt.Errorf(`DNS api call failed: %w`, err)
	}

	return nil
}
