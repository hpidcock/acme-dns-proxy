package config

const (
	defaultProviderName = "default"
)

type Config struct {
	Server      Server      `yaml:"server"`
	Provider    Provider    `yaml:"provider"`
	AccessRules AccessRules `yaml:"access_rules"`
}

type Server struct {
	Addr string `yaml:"addr"`
}

type Provider struct {
	Type      string            `yaml:"type"`
	Variables map[string]string `yaml:"variables"`
}

type AccessRules []AccessRule
type AccessRule struct {
	Pattern string `yaml:"pattern"`
	Token   string `yaml:"token"`
}
