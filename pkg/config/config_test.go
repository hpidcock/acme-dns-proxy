package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hpidcock/acme-dns-proxy/pkg/config"
)

func TestParseConfig(t *testing.T) {
	_, err := config.Parse(`
server {
	listen_addr = ":https"
	certmagic "acme.domain.example" {
	}
}
provider "cloudflare" {
	api_token = "my cloudflare api token"
}
acl "service-0.domain.example" {
	token = "secure token for service-0"
}
acl "*.sub.domain.example" {
	token = "secure token for all *.sub.domain.example"
}
`[1:])
	assert.NoError(t, err)
}
