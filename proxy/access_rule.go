package proxy

import (
	"fmt"
	"strings"

	"github.com/matthiasng/acme-dns-proxy/config"
)

// AccessRule defines a pattern and the corresponding auth key
type AccessRule struct {
	Pattern *Pattern
	Token   string
}

// NewAccessRulesFromConfig creates AccessRules from configuration
func NewAccessRulesFromConfig(accessRulesCfg config.AccessRules) (AccessRules, error) {
	if len(accessRulesCfg) == 0 {
		return nil, fmt.Errorf("error loading access rules: no access rules defined")
	}

	createRule := func(ruleCfg config.AccessRule) (*AccessRule, error) {
		pattern, err := CompilePattern(ruleCfg.Pattern)
		if err != nil {
			return nil, fmt.Errorf(`invalid pattern: "%s", error: %w`, ruleCfg.Pattern, err)
		}

		if len(ruleCfg.Token) == 0 {
			return nil, fmt.Errorf("'token' not specified")
		}

		return &AccessRule{
			Pattern: pattern,
			Token:   ruleCfg.Token,
		}, nil
	}

	rules := AccessRules{}
	for _, ruleCfg := range accessRulesCfg {
		rule, err := createRule(ruleCfg)
		if err != nil {
			return nil, fmt.Errorf("error loading access rules %s: %w", ruleCfg.Pattern, err)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// CheckAuth validate a given token against the AccessRule token
func (a *AccessRule) CheckAuth(token string) bool {
	return token == a.Token
}

// AccessRules is a list of AccessRules
type AccessRules []*AccessRule

// Search for a access rule by FQDN
func (a AccessRules) Search(fqdn string) *AccessRule {
	for _, rule := range a {
		domain := strings.TrimRight(fqdn, ".")
		if rule.Pattern.Match(domain) {
			return rule
		}
	}

	return nil
}
