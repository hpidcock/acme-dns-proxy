# ACME DNS Proxy

Proxy to secure ACME DNS challenges.

Most DNS providers do not offer a way to restrict access only to TXT records or to a specific domain. DigitalOcean for example only offers API tokens with full cloud access.

This creates a security issue if you use multipe host with `acme.sh` or `lego`, for example, because you have to distribute your API key among the host.

With `ACME DNS Proxy` you can control which client has access to which domains without storing your DNS Provider API keys on the client.

## Example configuration

```hcl
server {
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
```

## TODO

- Rewrite README.md
- Rewrite all unit tests
- Re-add in all libdns providers

Original project by [matthiasng](https://github.com/matthiasng/acme-dns-proxy)