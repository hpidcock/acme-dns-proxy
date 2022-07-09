package config

import (
	"github.com/hashicorp/hcl/v2"
)

type Config struct {
	Server   Server   `hcl:"server,block"`
	Provider Provider `hcl:"provider,block"`
	ACLs     []ACL    `hcl:"acl,block"`
}

type Server struct {
	ListenAddress string     `hcl:"listen_addr,optional"`
	CertMagic     *CertMagic `hcl:"certmagic,block"`
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
