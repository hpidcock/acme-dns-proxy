package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"

	"github.com/caddyserver/certmagic"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"

	"github.com/hpidcock/acme-dns-proxy/pkg/config"
	"github.com/hpidcock/acme-dns-proxy/pkg/dns"
	"github.com/hpidcock/acme-dns-proxy/pkg/listener"
	"github.com/hpidcock/acme-dns-proxy/pkg/proxy"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "/etc/acmep.d/config.hcl", "config file")
	flag.Parse()

	log := logrus.New()
	err := cmd(log, configFile)
	if err != nil {
		log.Panic(err)
	}
}

func cmd(log *logrus.Logger, configFile string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Infof("parse config %s", configFile)
	cfg, err := config.ParseFile(configFile)
	if err != nil {
		return errors.Annotatef(err, "failed to parse config: %s", configFile)
	}

	provider, err := dns.NewProviderFromConfig(&cfg.Provider, dns.DefaultZoneResolver)
	if err != nil {
		return errors.Annotate(err, "invalid provider")
	}

	acls, err := proxy.NewACLsFromConfig(cfg.ACLs)
	if err != nil {
		return errors.Annotate(err, "invalid acls")
	}

	proxy := proxy.Proxy{
		Log:      log,
		Provider: provider,
		ACLs:     acls,
	}

	var l net.Listener
	if cfg.Server.CertMagic != nil {
		certmagic.DefaultACME.DNS01Solver = &certmagic.DNS01Solver{
			DNSProvider: provider.Underlying(),
		}
		l, err = certmagic.Listen([]string{cfg.Server.CertMagic.Host})
		if err != nil {
			return errors.Annotatef(err, "certmagic listen for host %s", cfg.Server.CertMagic.Host)
		}
	} else {
		l, err = net.Listen("tcp", cfg.Server.ListenAddress)
		if err != nil {
			return errors.Annotatef(err, "net listen on %s", cfg.Server.ListenAddress)
		}
	}

	go listener.Serve(ctx, log, l, proxy)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	return nil
}
