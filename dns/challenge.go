package dns

// Challenge holds information about an ACME challenge.
type Challenge struct {
	EncodedKeyAuth string // encoded key authorization value
	FQDN           string // FQDN we want to verify
}
