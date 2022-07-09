package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
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

	defaultConfigFile string = "/etc/acmep.d/config.hcl"
)

func main() {
	var configFile string
	var doInstall bool
	var noExec bool
	flag.BoolVar(&doInstall, "install", false, "installs systemd service")
	flag.BoolVar(&noExec, "no-exec", false, "")
	flag.StringVar(&configFile, "config", defaultConfigFile, "config file")
	flag.Parse()

	log := logrus.New()

	if doInstall {
		err := install(log, noExec)
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

func install(log *logrus.Logger, noExec bool) error {
	self, err := filepath.Abs(os.Args[0])
	if err != nil {
		return err
	}
	if os.Getuid() != 0 {
		if noExec {
			return fmt.Errorf("must be run as root")
		}
		cmd := exec.Command("/usr/bin/sudo", self, "--install", "--no-exec")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	err = ioutil.WriteFile("/etc/systemd/system/acmep.service", []byte(serviceFile), 0777)
	if err != nil {
		return errors.Annotate(err, "creating systemd service")
	}
	err = os.MkdirAll("/etc/acmep.d/", 0755)
	if err != nil {
		return errors.Annotate(err, "creating acmep.d folder")
	}
	_, err = os.Stat(defaultConfigFile)
	if errors.Is(err, os.ErrNotExist) {
		answers := struct {
			Host            string `survey:"host"`
			CloudflareToken string `survey:"cloudflare_token"`
		}{}
		err = survey.Ask(initQuestions, &answers)
		if err != nil {
			return errors.Annotate(err, "survey failed")
		}
		configStr := fmt.Sprintf(`
server {
	listen_address = ":https"
	certmagic %q {
	}
}
provider "cloudflare" {
	api_token = %q
}`[1:], answers.Host, answers.CloudflareToken)
		err = ioutil.WriteFile(defaultConfigFile, []byte(configStr), 0644)
		if err != nil {
			return errors.Annotatef(err, "writing default config %s", defaultConfigFile)
		}
	} else if err != nil {
		return errors.Trace(err)
	}
	cfg, err := config.ParseFile(defaultConfigFile)
	if err != nil {
		return errors.Annotatef(err, "failed to parse config: %s", defaultConfigFile)
	}
	if cfg.Server.CertMagic != nil {
		provider, err := dns.NewProviderFromConfig(&cfg.Provider, dns.DefaultZoneResolver)
		if err != nil {
			return errors.Annotate(err, "invalid provider")
		}
		certmagic.DefaultACME.DNS01Solver = &certmagic.DNS01Solver{
			DNSProvider: provider.Underlying(),
		}
		certmagic.DefaultACME.Agreed = true
		cmCfg := certmagic.NewDefault()
		err = cmCfg.ManageSync(context.Background(), []string{cfg.Server.CertMagic.Host})
		if err != nil {
			return errors.Annotatef(err, "certmagic listen for host %s", cfg.Server.CertMagic.Host)
		}
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

var initQuestions = []*survey.Question{
	{
		Name:     "host",
		Prompt:   &survey.Input{Message: "What is the DNS name for this acmep instance?"},
		Validate: survey.Required,
	},
	{
		Name:     "cloudflare_token",
		Prompt:   &survey.Input{Message: "What is your cloudflare api token?"},
		Validate: survey.Required,
	},
}
