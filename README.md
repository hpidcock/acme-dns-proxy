# ACME DNS Proxy

Proxy to secure ACME DNS challenges.

Most DNS providers do not offer a way to restrict access only to TXT records or to a specific domain. DigitalOcean for example only offers API tokens with full cloud access.

This creates a security issue if you use multipe host with `acme.sh` or `lego`, for example, because you have to distribute your API key among the host.

With `ACME DNS Proxy` you can control which client has access to which domains without storing your DNS Provider API keys on the client.

## Features
- Restrict ACME client access to specified (sub)domains
- CertMagic or self signed certificate for the proxy itself (TODO)
- Monitoring endpoint (prometheus) (TODO)
- "auto cleanup" DNS records (TODO)

## Supported clients:
- [acmesh-official/acme.sh](https://github.com/acmesh-official/acme.sh) with [dns_acmeproxy](https://github.com/acmesh-official/acme.sh/wiki/dnsapi#78-use-acmeproxy-dns-api)
- [go-acme/lego](https://github.com/go-acme/lego) with [httpreq](https://go-acme.github.io/lego/dns/httpreq/)
- [traefik/traefik](https://doc.traefik.io/traefik/https/acme/#providers) which uses [go-acme/lego](https://github.com/go-acme/lego)
- Everything else that can send a HTTP request

## Configuration & Usage

The configuration consists of three main parts. `server`, `provider` and `access`

---

Under `server` you can configure common stuff like TLS and the address, the server listens to.
```yaml
server:
  addr: ":8080"
```

---

The `provider` section configures the access to your DNS provider.
```yaml
provider:
  type: gcloud
  variables:
    GCE_PROJECT: some-project
    GCE_SERVICE_ACCOUNT: some-service-account
    GCE_SERVICE_ACCOUNT_FILE: /some-service-account-file.json
```

`type:` \
`type` specifies the DNS provider. 
`acme-dns-proxy` uses [libdns/libdns](https://github.com/libdns/libdns) to add and remove DNS records. Please see the list of [Supported Providers](https://github.com/matthiasng/libdnsfactory/blob/master/docs.md) section for further information.
All providers support 

`variables:` \
Which `variables` are available depends on the `type`.
Please see the list [Supported Providers](https://github.com/matthiasng/libdnsfactory/blob/master/docs.md) section for further information.

---
The `access_rules` section specifies the domains for which a client can request a certificate.

```yaml
access_rules:
  - pattern: "*.a.b.c.matthiasng.com"
    token: f9e5f6a00056b7913fea130aa31921ccae1b4cb152a12999d7751e667c016344
  - pattern: matthiasng.com
    token: f97b0d33302f348adf6ed887961156cc11b2436fd4699e7aa759becd8d96c7e3
  - pattern: "x.y.matthiasng.com"
    token: f71876a55b38a12a5da6ec1900a5cf09c7a06574726d42b3295614cc7f20b344
```

`pattern:` \
A glob pattern that must match the domain a client is allowed to verify.
- `matthiasng.com`: only allow `matthiasng.com`
- `*.sub.matthiasng.com`: allow all subdomains but not `sub.matthiasng.com`
- `*foo.matthiasng.com`: all subdomains, the current domain, and each subdomain of `matthiasng.com` starting with `*foo`

`token:` \
A token to verify the request.

For acme.sh and lego this must be the SHA 256 value of `<username>:<password>`.
This way we dont need an extra client plugin and you can integrate the proxy inside existing infrastructure easily.

```sh
echo -n <user>:<password> | sha256sum
```

---

The configration file supports [golang's template](https://golang.org/pkg/text/template/).
The following variables are available:
- `Env` \
    Contains all environment variables. \
    ```yaml
    provider:
      type: hetzner
      variables:
        AuthAPIToken: {{ .Env.HETZNER_AUTH_API_TOKEN }}
    ```

# TODOs
- https://github.com/rmbolger/Posh-ACME/wiki/List-of-Supported-DNS-Providers
- multiple providers
