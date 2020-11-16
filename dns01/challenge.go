package dns01

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// TXTRecordName returns the name of the TXT record to create for solving the dns-01 challenge.
func TXTRecordName(domain string) string {
	return "_acme-challenge." + domain
}

// EncodeKeyAuthorization encodes a key authorization value to be used in a TXT record.
func EncodeKeyAuthorization(keyAuth string) string {
	h := sha256.Sum256([]byte(keyAuth))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// FQDNFromTXTRecordName returns the FQDN for a TXT record name
func FQDNFromTXTRecordName(name string) string {
	return ToFqdn(strings.TrimPrefix(name, "_acme-challenge."))
}

// RemoveZoneFromFqdn removes the zone from an FQDN and return the sub domain
func RemoveZoneFromFqdn(fqdn, zone string) string {
	return strings.TrimSuffix(fqdn, zone)
}
