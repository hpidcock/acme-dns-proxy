# ACME DNS Proxy

Proxy to secure ACME DNS challenges.

Most DNS providers do not offer a way to restrict access only to TXT records or to a specific domain. DigitalOcean for example only offers API tokens with full cloud access.

When someone gets the key, they not only have control over your entire domain, in the case of digitalocean they also have access to all cloud resources. Obviously you want to keep your token as secret as possible.

But when you are using ACME Client's like `acme.sh` or `lego`, you have to set the token somewhere on your client system.

With `ACME DNS Proxy` you can easily control which client has access to which domains without storing your DNS Provider API keys on the client

## Supported clients:
- [acmesh-official/acme.sh](https://github.com/acmesh-official/acme.sh) with [dns_acmeproxy](https://github.com/acmesh-official/acme.sh/wiki/dnsapi#78-use-acmeproxy-dns-api)
- [go-acme/lego](https://github.com/go-acme/lego) with [httpreq](https://go-acme.github.io/lego/dns/httpreq/)
- [certmagic](https://github.com/caddyserver/certmagic) with [httpreq](https://go-acme.github.io/lego/dns/httpreq/)
- Everything else that can send a HTTP request

## Configuration & Usage

The configuration consists of three main parts. `server`, `provider` and `access`

---

Under `server` you can configure common stuff like TLS and the address and port, the server listens to.
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
Please see the [TODO Link: Supported Providers](#) section for further information.

`variables:` \
Which `variables` are available depends on the `type`.
Please see the [TODO Link: Supported Providers](#) section for further information.

---
The `access` section specifies the domains for which a client can request a certificate.

```yaml
access:
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
- `*sub.matthiasng.com`: all subdomains, the current domain, and each subdomain of `matthiasng.com` starting with `*sub`

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
      type: digitalocean
      variables:
        DO_AUTH_TOKEN: {{ .Env.DO_AUTH_TOKEN }}
    ```
