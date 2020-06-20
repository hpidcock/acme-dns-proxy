package listener

import (
	"github.com/matthiasng/dns-challenge-proxy/proxy"
)

// Listener handles incomding requests
type Listener interface {
	ListenAndServe(proxy.Proxy) error
	Shutdown() error
}
