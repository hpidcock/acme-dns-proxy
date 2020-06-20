package main

import (
	"os"
	"os/signal"

	"github.com/matthiasng/dns-challenge-proxy/config"
	"github.com/matthiasng/dns-challenge-proxy/listener"
	"github.com/matthiasng/dns-challenge-proxy/proxy"

	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func main() {
	// #todo flags
	cfgFilename := "./config.yml"

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Sugar().Infow("load config", "from", cfgFilename)
	loader := config.NewFileLoader(afero.NewOsFs(), cfgFilename)

	cfg, err := loader.Load()
	if err != nil {
		logger.Panic("Failed to load config", zap.Error(err))
	}

	provider, err := proxy.NewLegoProviderFromConfig(&cfg.Provider)
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
		logger.Sugar().Infow("start listening", "on", cfg.Server.Addr)
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
