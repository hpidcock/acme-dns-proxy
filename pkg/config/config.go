package config

import (
	"github.com/hashicorp/hcl/v2"
)

type Config struct {
	Server   Server   `hcl:"server"`
	Provider Provider `hcl:"provider"`
	ACLs     []ACL    `hcl:"acl,block"`
}

type Server struct {
	ListenAddress string     `hcl:"listen_addr"`
	CertMagic     *CertMagic `hcl:"certmagic"`
}

type CertMagic struct {
	Host string `hcl:"host,label"`
}

type Provider struct {
	Type   string   `hcl:"type,label"`
	Remain hcl.Body `hcl:",remain"`
}

type ACL struct {
	Pattern string `hcl:"pattern,label"`
	Token   string `hcl:"token"`
}
