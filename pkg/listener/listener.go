package listener

import (
	"context"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/hpidcock/acme-dns-proxy/pkg/proxy"
)

// Serve accepts connections on the net.Listener and handles requests.
func Serve(ctx context.Context, log *logrus.Logger, listener net.Listener, p proxy.Proxy) {
	server := &http.Server{
		Handler: newHTTPHandler(p),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
	err := server.Serve(listener)
	if err != nil {
		log.Error(err)
	}
}
