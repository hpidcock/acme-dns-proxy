package proxy

import (
	"fmt"

	"github.com/gobwas/glob"
)

// Pattern is a glob-like string for ACL
type Pattern struct {
	source string
	glob   glob.Glob
}

// CompilePattern creates a Pattern form a given string
func CompilePattern(source string) (Pattern, error) {
	g, err := glob.Compile(source)
	if err != nil {
		return Pattern{}, fmt.Errorf("invalid pattern: %q, error: %w", source, err)
	}
	return Pattern{
		source: source,
		glob:   g,
	}, nil
}

// MustCompilePattern creates a Pattern form a given string. Panics in case of an error
func MustCompilePattern(source string) Pattern {
	pattern, err := CompilePattern(source)
	if err != nil {
		panic(err)
	}
	return pattern
}

// Match the pattern with a given string
func (p *Pattern) Match(s string) bool {
	return p.glob.Match(s)
}

// String returns the pattern as string
func (p *Pattern) String() string {
	return p.source
}
