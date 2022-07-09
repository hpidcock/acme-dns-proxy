package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

func ParseFile(filename string) (*Config, error) {
	cfg := &Config{}
	err := hclsimple.DecodeFile(filename, nil, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", filename, err)
	}
	return cfg, nil
}

func Parse(config string) (*Config, error) {
	cfg := &Config{}
	err := hclsimple.Decode("config.hcl", []byte(config), nil, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}
