package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	copirator "github.com/quickybrains/copirator/internal"
	"github.com/quickybrains/copirator/internal/log/zerolog"
	"gopkg.in/yaml.v3"
)

func main() {
	cfg, err := readConfig()
	if err != nil {
		log.Fatalf("can't read config, err=%v", err.Error())
	}

	logger := zerolog.New(cfg.Log)

	app, err := copirator.StartNewApp(cfg, logger)
	if err != nil {
		log.Fatalf("can't init app, err=%v", err.Error())
	}

	if cfg.Lifecycle.SingleRun {
		logger.Info().Print("Single run finished, shutting down")

		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Print("Got exit signal, shutting down")

	app.Stop()
}

func readConfig() (copirator.Config, error) {
	cfgPath := os.Getenv("CONFIG_FILE_PATH")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

	f, err := os.Open(cfgPath)
	if err != nil {
		return copirator.Config{}, fmt.Errorf("os open %s: %w", cfgPath, err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return copirator.Config{}, fmt.Errorf("io readall: %w", err)
	}

	var cfg copirator.Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return copirator.Config{}, fmt.Errorf("yaml unmarshal: %w", err)
	}

	return cfg, nil
}
