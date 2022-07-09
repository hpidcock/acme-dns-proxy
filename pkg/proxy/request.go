package proxy

import (
	"github.com/hpidcock/acme-dns-proxy/pkg/dns"
)

// Request holds information about the request.
type Request struct {
	Action    string        // Action for the current request. Can be present or cleanup
	AuthToken string        // AuthToken for the current request
	Challenge dns.Challenge // Challenge for the current request
	Remote    Remote        // Remote information for the current request
}

// Remote holds information about the remote client.
type Remote struct {
	Address string // Address is the address of the client
	Name    string // Name is the name of the client. This depends on the client and can be for example go/lego for lego
}
