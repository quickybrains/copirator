package internal

import (
	"fmt"
	"os"

	"github.com/quickybrains/copirator/internal/log"
)

type FileWriter struct {
	cfg    FileOutputConfig
	logger log.Logger
}

func NewFileWriter(cfg FileOutputConfig, logger log.Logger) *FileWriter {
	return &FileWriter{
		cfg:    cfg,
		logger: logger.With("component", "file_writer"),
	}
}

func (fw *FileWriter) Write(data []byte) error {
	if !fw.cfg.Enabled {
		return nil
	}

	f, err := os.Create(fw.cfg.Path)
	if err != nil {
		return fmt.Errorf("os create %s: %w", fw.cfg.Path, err)
	}

	fw.logger.Info().String("path", fw.cfg.Path).Print("Writing backup to file")

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("file write %s: %w", fw.cfg.Path, err)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("file close %s: %w", fw.cfg.Path, err)
	}

	return nil
}
