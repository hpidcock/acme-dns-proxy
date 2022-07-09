package listener

import (
	"context"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/hpidcock/acme-dns-proxy/pkg/proxy"
)

// Serve accepts connections and handles requests.
func Serve(ctx context.Context, log *logrus.Logger, defaultServer http.Server, p proxy.Proxy, done chan any) {
	defer close(done)
	var err error
	server := &defaultServer
	server.Handler = newHTTPHandler(p)
	server.BaseContext = func(l net.Listener) context.Context {
		return ctx
	}
	if server.TLSConfig != nil {
		err = server.ListenAndServeTLS("", "")
	} else {
		err = server.ListenAndServe()
	}
	if err != nil {
		log.Error(err)
	}
}
