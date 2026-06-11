package internal

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/quickybrains/copirator/internal/log"
)

type compressor interface {
	Compress() ([][]byte, error)
}

type encryptor interface {
	Encrypt(src [][]byte) ([][]byte, error)
}

type outputWriter interface {
	Write([][]byte) error
}

type App struct {
	cfg          Config
	basicWriters map[string]outputWriter
	safeWriters  map[string]outputWriter
	encryptor    encryptor
	wg           sync.WaitGroup
	stopCh       chan struct{}

	compressor compressor
	logger     log.Logger
}

func StartNewApp(cfg Config, logger log.Logger) (*App, error) {
	a := &App{
		cfg:          cfg,
		basicWriters: make(map[string]outputWriter),
		safeWriters:  make(map[string]outputWriter),

		logger: logger.With("component", "app"),
	}

	err := a.init()
	if err != nil {
		return nil, fmt.Errorf("app init: %w", err)
	}

	return a, nil
}

func (a *App) init() error {
	a.compressor = NewCompressor(a.cfg, a.logger)

	emailWriter, err := NewEmailWriter(a.cfg.Output.Email, a.logger)
	if err != nil {
		return fmt.Errorf("new email writer: %w", err)
	}

	a.basicWriters["file"] = NewFileWriter(a.cfg.Output.File, a.logger)
	a.safeWriters["email"] = emailWriter

	a.encryptor = NewEncryptor(a.cfg.Encryption, a.logger)

	if a.cfg.Lifecycle.SingleRun {
		err := a.runBackup()
		if err != nil {
			return fmt.Errorf("single run backup: %w", err)
		}

		return nil
	}

	a.wg.Go(a.backupLoop)

	return nil
}

func (a *App) Stop() {
	if a.cfg.Lifecycle.SingleRun {
		return
	}

	close(a.stopCh)
	a.wg.Wait()
}

func (a *App) backupLoop() {
	const defaultInterval = 24 * time.Hour

	interval := defaultInterval
	if a.cfg.Lifecycle.BackupInterval > 0 {
		interval = a.cfg.Lifecycle.BackupInterval
	}

	ticker := time.NewTicker(interval)
	tickCh := time.After(0)

	a.logger.
		Info().
		Int64("hours", int64(interval.Hours())).
		Int64("minutes", int64(interval.Minutes())).
		Print("Starting backup loop")

	for {
		select {
		case <-tickCh:
			tickCh = ticker.C

			a.logger.Info().Print("Running backup")

			err := a.runBackup()
			if err != nil {
				a.logger.Err(err).Print("Run backup error")
			}
		case <-a.stopCh:
			a.logger.Info().Print("Shutting down")
			return
		}
	}
}

func (a *App) runBackup() error {
	outdated, err := a.backupOutdated()
	if err != nil {
		return fmt.Errorf("backup outdated: %w", err)
	}

	if !outdated {
		a.logger.Info().String("path", a.cfg.Output.File.Path).Print("Backup is up-to-date, skip backup")

		return nil
	}

	a.logger.Info().String("path", a.cfg.Output.File.Path).Print("Backup is outdated, start backup")

	data, err := a.compressor.Compress()
	if err != nil {
		return fmt.Errorf("compress: %w", err)
	}

	for name, wr := range a.basicWriters {
		err = wr.Write(data)
		if err != nil {
			return fmt.Errorf("%s output write: %w", name, err)
		}
	}

	for name, wr := range a.safeWriters {
		data, err := a.encryptor.Encrypt(data)
		if err != nil {
			return fmt.Errorf("encrypt: %w", err)
		}

		err = wr.Write(data)
		if err != nil {
			return fmt.Errorf("%s output write: %w", name, err)
		}
	}

	return nil
}

func (a *App) backupOutdated() (bool, error) {
	outputPath := a.cfg.Output.File.Path
	f, err := os.Open(outputPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}

		return false, fmt.Errorf("os open: %w", err)
	}

	finfo, err := f.Stat()
	if err != nil {
		return false, fmt.Errorf("file stat: %w", err)
	}

	if time.Since(finfo.ModTime().Add(-1*time.Minute)) >= a.cfg.Lifecycle.BackupInterval {
		return true, nil
	}

	return false, nil
}
