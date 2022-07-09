package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/caddyserver/certmagic"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"

	"github.com/hpidcock/acme-dns-proxy/pkg/config"
	"github.com/hpidcock/acme-dns-proxy/pkg/dns"
	"github.com/hpidcock/acme-dns-proxy/pkg/listener"
	"github.com/hpidcock/acme-dns-proxy/pkg/proxy"
)

const (
	restartErr errors.ConstError = "restart acmep"
)

func main() {
	var configFile string
	var doInstall bool
	flag.BoolVar(&doInstall, "install", false, "installs systemd service")
	flag.StringVar(&configFile, "config", "/etc/acmep.d/config.hcl", "config file")
	flag.Parse()

	log := logrus.New()

	if doInstall {
		err := install(log)
		if err != nil {
			log.Panic(err)
		}
		return
	}

	for {
		err := cmd(log, configFile)
		if errors.Is(err, restartErr) {
			log.Info("reloading")
			continue
		} else if err != nil {
			log.Panic(err)
		}
		log.Info("shutdown")
		break
	}
}

func install(log *logrus.Logger) error {
	if os.Getuid() != 0 {
		return fmt.Errorf("must be root")
	}
	err := ioutil.WriteFile("/etc/systemd/system/acmep.service", []byte(serviceFile), 0777)
	if err != nil {
		return errors.Annotate(err, "creating systemd service")
	}
	return nil
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

	server := http.Server{
		Addr: cfg.Server.ListenAddress,
	}
	if cfg.Server.CertMagic != nil {
		certmagic.DefaultACME.DNS01Solver = &certmagic.DNS01Solver{
			DNSProvider: provider.Underlying(),
		}
		certmagic.DefaultACME.Agreed = true
		cmCfg := certmagic.NewDefault()
		err := cmCfg.ManageSync(ctx, []string{cfg.Server.CertMagic.Host})
		if err != nil {
			return errors.Annotatef(err, "certmagic listen for host %s", cfg.Server.CertMagic.Host)
		}
		server.TLSConfig = cmCfg.TLSConfig()
		server.TLSConfig.NextProtos = append([]string{"h2", "http/1.1"}, server.TLSConfig.NextProtos...)
	}

	done := make(chan any)
	go listener.Serve(ctx, log, server, proxy, done)
	defer func() { <-done }()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)
	sig := <-signalChan
	if sig == syscall.SIGHUP {
		return restartErr
	}
	return nil
}

const serviceFile = `[Unit]
Description=ACME DNS Proxy server
After=network.target auditd.service

[Service]
ExecStart=/usr/local/bin/acmep
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process
Restart=on-failure
RestartPreventExitStatus=255
Type=notify

[Install]
WantedBy=multi-user.target
Alias=acmep.service
`
