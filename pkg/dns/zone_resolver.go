package dns

import (
	"github.com/hpidcock/acme-dns-proxy/pkg/dns01"
)

// ZoneResolver resolve the zone from an FQDN
type ZoneResolver = func(string) (string, error)

// DefaultZoneResolver determines the authoritative zone for the given fqdn by recursing
// up the domain labels until the nameserver returns a SOA record in the answer section.
func DefaultZoneResolver(fqdn string) (string, error) {
	return dns01.FindZoneByFQDN(fqdn, dns01.RecursiveNameservers(nil)) // TODO: nameserver config
}
