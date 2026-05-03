package internal

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/quickybrains/copirator/internal/log"
)

type Compressor struct {
	cfg         Config
	buff        *bytes.Buffer
	slashSymbol string

	logger log.Logger
}

func NewCompressor(cfg Config, logger log.Logger) *Compressor {
	c := &Compressor{
		cfg:         cfg,
		buff:        new(bytes.Buffer),
		slashSymbol: "/",

		logger: logger.With("component", "compressor"),
	}

	os := runtime.GOOS

	switch os {
	case "windows":
		c.slashSymbol = "\\"
	default:
		c.slashSymbol = "/"
	}

	return c
}

// TODO: add logging
func (c *Compressor) Compress() ([]byte, error) {
	// We start new compression, reset
	c.buff.Reset()

	w := zip.NewWriter(c.buff)

	for _, file := range c.cfg.Files {
		outFileName, err := c.getFileName(file.OutputDir, file.Path)
		if err != nil {
			return nil, fmt.Errorf("get file name: %w", err)
		}

		zf, err := w.Create(outFileName)
		if err != nil {
			return nil, fmt.Errorf("zip file create %s: %w", file.OutputDir, err)
		}

		f, err := os.Open(file.Path)
		if err != nil {
			return nil, fmt.Errorf("os open: %w", err)
		}

		_, err = io.Copy(zf, f)
		if err != nil {
			return nil, fmt.Errorf("compress file %s: %w", file.Path, err)
		}

		err = f.Close()
		if err != nil {
			return nil, fmt.Errorf("file %s close: %w", file.Path, err)
		}

		c.logger.Info().String("zip_file", outFileName).Print("Compressed file")
	}

	err := w.Close()
	if err != nil {
		return nil, fmt.Errorf("zip close: %w", err)
	}

	c.logger.Info().Print("Compression successful")

	return c.buff.Bytes(), nil
}

func (c *Compressor) getFileName(outDir string, filePath string) (string, error) {
	idx := strings.LastIndex(filePath, c.slashSymbol)
	if idx == -1 {
		return "", fmt.Errorf("path to file required, instead get %s", filePath)
	}

	if idx == len(filePath)-1 {
		return "", fmt.Errorf("file path should contain name of the file")
	}

	return outDir + filePath[idx+1:], nil
}
