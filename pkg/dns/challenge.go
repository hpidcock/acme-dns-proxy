package dns

import (
	"github.com/juju/errors"
)

// Challenge holds information about an ACME challenge.
type Challenge struct {
	EncodedKeyAuth string // encoded key authorization value
	FQDN           string // FQDN we want to verify
}

func (c Challenge) Validate() error {
	if len(c.FQDN) == 0 {
		return errors.New("invalid request: fqdn not set")
	}
	if len(c.EncodedKeyAuth) == 0 {
		return errors.New("invalid request: key auth value not set")
	}
	return nil
}
