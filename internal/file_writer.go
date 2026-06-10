package internal

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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

func (fw *FileWriter) Write(data [][]byte) error {
	if !fw.cfg.Enabled {
		return nil
	}

	for idx, nextData := range data {
		fileName := nextFileName(fw.cfg.Path, idx+1)

		f, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("os create %s: %w", fileName, err)
		}

		fw.logger.Info().
			Int("fileIdx", idx+1).
			String("path", fileName).
			Print("Writing backup to file")

		_, err = f.Write(nextData)
		if err != nil {
			return fmt.Errorf("file write %s: %w", fileName, err)
		}

		err = f.Close()
		if err != nil {
			return fmt.Errorf("file close %s: %w", fileName, err)
		}
	}

	return nil
}

func nextFileName(name string, idx int) string {
	var suffix string

	dotIdx := strings.LastIndex(name, ".")
	if dotIdx != -1 {
		suffix = name[dotIdx:]
		name = name[:dotIdx]

	}

	name += "_" + strconv.FormatInt(int64(idx), 10)

	return name + suffix
}
