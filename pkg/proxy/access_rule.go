package proxy

import (
	"fmt"
	"strings"

	"github.com/juju/errors"

	"github.com/hpidcock/acme-dns-proxy/pkg/config"
)

// ACL defines a pattern and the corresponding auth key
type ACL struct {
	Pattern Pattern
	Token   string
}

// NewACLsFromConfig creates ACLs from configuration
func NewACLsFromConfig(cfg []config.ACL) (ACLs, error) {
	if len(cfg) == 0 {
		return nil, fmt.Errorf("error loading access rules: no access rules defined")
	}

	createRule := func(ruleCfg config.ACL) (ACL, error) {
		pattern, err := CompilePattern(ruleCfg.Pattern)
		if err != nil {
			return ACL{}, fmt.Errorf("invalid pattern: %q, error: %w", ruleCfg.Pattern, err)
		}

		if len(ruleCfg.Token) == 0 {
			return ACL{}, fmt.Errorf("'token' not specified")
		}

		return ACL{
			Pattern: pattern,
			Token:   ruleCfg.Token,
		}, nil
	}

	rules := ACLs{}
	for _, ruleCfg := range cfg {
		rule, err := createRule(ruleCfg)
		if err != nil {
			return nil, fmt.Errorf("error loading access rules %s: %w", ruleCfg.Pattern, err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// CheckAuth validate a given token against the ACL token
func (a *ACL) CheckAuth(token string) bool {
	return token == a.Token
}

// ACLs is a list of ACLs
type ACLs []ACL

// Search for a access rule by FQDN
func (a ACLs) Search(fqdn string) (ACL, error) {
	for _, rule := range a {
		domain := strings.TrimRight(fqdn, ".")
		if rule.Pattern.Match(domain) {
			return rule, nil
		}
	}
	return ACL{}, errors.NotFoundf("acl for fqdn %q", fqdn)
}
