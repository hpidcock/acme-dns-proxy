package main

import (
	"os"
	"os/signal"

	"github.com/matthiasng/acme-dns-proxy/config"
	"github.com/matthiasng/acme-dns-proxy/dns"
	"github.com/matthiasng/acme-dns-proxy/listener"
	"github.com/matthiasng/acme-dns-proxy/proxy"

	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func main() {
	// #todo flags
	cfgFilename := "./config.yml"

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Sugar().Infow("load config", "file", cfgFilename)
	loader := config.NewFileLoader(afero.NewOsFs(), cfgFilename)

	cfg, err := loader.Load()
	if err != nil {
		logger.Panic("Failed to load config", zap.Error(err))
	}

	provider, err := dns.NewProviderFromConfig(&cfg.Provider, dns.DefaultZoneResolver)
	if err != nil {
		logger.Panic("Invalid access rules", zap.Error(err))
	}

	accessRules, err := proxy.NewAccessRulesFromConfig(cfg.AccessRules)
	if err != nil {
		logger.Panic("Invalid access rules", zap.Error(err))
	}

	proxy := proxy.Proxy{
		Logger:      logger,
		Provider:    provider,
		AccessRules: accessRules,
	}

	listener := listener.NewHTTP(cfg.Server.Addr)

	go func() {
		logger.Sugar().Infow("start listening",
			"on", cfg.Server.Addr,
		)
		if err := listener.ListenAndServe(proxy); err != nil {
			panic(err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	sig := <-signalChan

	logger.Sugar().Infow("signal received",
		"signal", sig,
	)

	if err := listener.Shutdown(); err != nil {
		panic(err)
	}

}
