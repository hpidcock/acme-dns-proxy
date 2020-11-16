package dns01

import "strings"

// ToFqdn converts the name into a fqdn appending a trailing dot.
func ToFqdn(name string) string {
	if strings.HasSuffix(name, ".") {
		return name
	}
	return name + "."
}

// UnFqdn converts the fqdn into a name removing the trailing dot.
func UnFqdn(name string) string {
	return strings.TrimRight(name, ".")
}
