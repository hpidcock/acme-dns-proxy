package listener

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hpidcock/acme-dns-proxy/pkg/dns"
	"github.com/hpidcock/acme-dns-proxy/pkg/dns01"
	"github.com/hpidcock/acme-dns-proxy/pkg/proxy"
)

func newHTTPHandler(p proxy.Proxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		default:
			p.Log.Errorf("method not allowed: %s %s", r.Method, r.URL.String())
			methodNotAllowed(w)
			return
		case "GET":
			if r.URL.Path == "/" {
				ok(w)
				return
			}
			p.Log.Errorf("not found: %s %s", r.Method, r.URL.String())
			notFound(w)
			return
		case "POST":
		}

		action := strings.ToLower(strings.Trim(r.URL.Path, "/"))
		if action != "present" && action != "cleanup" {
			p.Log.Errorf("not found: %s %s", r.Method, r.URL.String())
			notFound(w)
			return
		}

		req, err := parseHTTPRequest(r)
		if err != nil {
			p.Log.Errorf("bad request: %s %s %s", r.Method, r.URL.String(), err.Error())
			badRequest(w, err)
			return
		}

		req.Action = action

		err = p.Handle(r.Context(), req)
		if err != nil {
			// We dont want to expose information to unauthorized clients.
			// So we dont care about the reason and always respond with unauthorized.
			p.Log.Errorf("unauthorized: %s %s %s", r.Method, r.URL.String(), err.Error())
			unauthorized(w)
			return
		}

		// TODO: do we have to send a response ?
		response, err := json.Marshal(struct {
			FQDN  string
			Value string
		}{req.Challenge.FQDN, req.Challenge.FQDN})
		if err != nil {
			p.Log.Errorf("internal server error: %s %s %s", r.Method, r.URL.String(), err.Error())
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
	var keyAuth string
	if !isLegoRawRequest(payload) {
		fqdn = dns01.FQDNFromTXTRecordName(payload["fqdn"])
		keyAuth = payload["value"]
	} else {
		fqdn = dns01.ToFQDN(payload["domain"])
		keyAuth = dns01.EncodeKeyAuthorization(payload["keyAuth"])
	}

	token, err := convertBasicAuthToToken(httpReq)
	if err != nil {
		return nil, err
	}

	req := proxy.Request{
		Challenge: dns.Challenge{
			FQDN:           dns01.ToFQDN(fqdn),
			EncodedKeyAuth: keyAuth,
		},
		Remote: proxy.Remote{
			Address: httpReq.RemoteAddr,
			Name:    httpReq.UserAgent(),
		},
		AuthToken: token,
	}

	return &req, nil
}

func convertBasicAuthToToken(req *http.Request) (string, error) {
	username, password, ok := req.BasicAuth()
	if !ok {
		return "", fmt.Errorf("invalid basic auth header")
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
