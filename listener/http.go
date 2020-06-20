package listener

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/matthiasng/dns-challenge-proxy/proxy"
)

func notFound(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func badRequest(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func forbidden(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

func internalServerError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func NewHTTP(addr string) Listener {
	return &httpListener{
		addr: addr,
	}
}

type httpListener struct {
	addr   string
	server *http.Server
}

func (h *httpListener) ListenAndServe(p proxy.Proxy) error {
	h.server = &http.Server{
		Addr:    h.addr,
		Handler: newHTTPHandler(p),
	}

	err := h.server.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (h *httpListener) Shutdown() error {
	return h.server.Shutdown(context.Background())
}

func newHTTPHandler(p proxy.Proxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			methodNotAllowed(w)
			return
		}

		action := strings.ToLower(strings.Trim(r.URL.Path, "/"))
		if action != "present" && action != "cleanup" {
			notFound(w)
			return
		}

		req, err := parseHTTPRequest(r)
		if err != nil {
			badRequest(w, err)
			return
		}

		req.Action = action

		err = p.Handle(req)
		if err != nil {
			// We dont care about the reason.
			forbidden(w)
			return
		}

		response, err := json.Marshal(struct {
			FQDN  string
			Value string
		}{req.Payload.TxtFQDN, req.Payload.TxtValue})

		if err != nil {
			internalServerError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func parseHTTPRequest(httpReq *http.Request) (*proxy.Request, error) {
	payload := map[string]string{}

	err := json.NewDecoder(httpReq.Body).Decode(&payload)
	if err != nil {
		return nil, fmt.Errorf("cannot parse request body: %w", err)
	}

	var fqdn string
	var value string
	if !isLegoRawRequest(payload) {
		fqdn = payload["fqdn"]
		value = payload["value"]
	} else {
		fqdn, value = dns01.GetRecord(payload["domain"], payload["keyAuth"])
	}

	token, err := convertBasicAuthToToken(httpReq)
	if err != nil {
		return nil, err
	}

	req := proxy.Request{
		Payload: proxy.Payload{
			TxtFQDN:  dns01.ToFqdn(fqdn),
			TxtValue: value,
		},
		Client: proxy.Client{
			RemoteAddr: httpReq.RemoteAddr,
			Name:       httpReq.UserAgent(),
		},
		Token: token,
	}

	return &req, nil
}

func convertBasicAuthToToken(req *http.Request) (string, error) {
	username, password, ok := req.BasicAuth()
	if !ok {
		return "", fmt.Errorf(`invalid basic auth header`)
	}

	in := []byte(fmt.Sprintf("%s:%s", username, password))
	hash := sha256.Sum256(in)
	return hex.EncodeToString(hash[:]), nil
}

func isLegoRawRequest(data map[string]string) bool {
	if _, ok := data["domain"]; !ok {
		return false
	}
	if _, ok := data["keyAuth"]; !ok {
		return false
	}
	if _, ok := data["token"]; !ok {
		return false
	}

	return true
}
