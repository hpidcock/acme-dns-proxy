package dns01

import "strings"

// ToFQDN converts the name into a fqdn appending a trailing dot.
func ToFQDN(name string) string {
	if strings.HasSuffix(name, ".") {
		return name
	}
	return name + "."
}

// UnFQDN converts the fqdn into a name removing the trailing dot.
func UnFQDN(name string) string {
	return strings.TrimRight(name, ".")
}
