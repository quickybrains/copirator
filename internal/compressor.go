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

func (c *Compressor) Compress() ([]byte, error) {
	// We start new compression, reset
	c.buff.Reset()

	w := zip.NewWriter(c.buff)

	for _, file := range c.cfg.Files {
		fInfo, err := os.Stat(file.Path)
		if err != nil {
			return nil, fmt.Errorf("os stat: %w", err)
		}

		if fInfo.IsDir() {
			dirName, err := c.getDirName(file.Path)
			if err != nil {
				return nil, fmt.Errorf("get dir name: %w", err)
			}

			err = c.traverseDir(dirName, file.Path, c.slashSymbol, w)
			if err != nil {
				return nil, fmt.Errorf("traverse dir: %w", err)
			}

			continue
		}

		outFileName, err := c.getFileName(file.OutputDir, file.Path)
		if err != nil {
			return nil, fmt.Errorf("get file name: %w", err)
		}

		zf, err := w.Create(outFileName)
		if err != nil {
			return nil, fmt.Errorf("zip file create %s: %w", file.OutputDir, err)
		}

		err = copyFileToWriter(file.Path, zf)
		if err != nil {
			return nil, fmt.Errorf("copy file to writer: %w", err)
		}

		c.logger.Info().String("name", outFileName).Print("Compressed file")
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

	// If dir is empty we don't want leading slash in result
	if outDir != "" {
		idx += 1
	}

	return outDir + filePath[idx:], nil
}

func (c *Compressor) getDirName(dirPath string) (string, error) {
	idx := strings.LastIndex(dirPath, c.slashSymbol)
	if idx == -1 {
		return "", fmt.Errorf("path to file required, instead get %s", dirPath)
	}

	if idx == len(dirPath)-1 {
		return "", fmt.Errorf("file path should contain name of the file")
	}

	return dirPath[idx+1:], nil
}

func (c *Compressor) traverseDir(zipPath, dirPath, slashSym string, w *zip.Writer) error {
	c.logger.Info().String("path", dirPath).Print("Traversing directory")

	dirL, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("os read dir: %w", err)
	}

	for _, nextL := range dirL {
		nextDirPath := dirPath + slashSym + nextL.Name()
		nextZipPath := zipPath + slashSym + nextL.Name()
		if nextL.IsDir() {
			c.logger.Info().String("path", nextZipPath).Print("Traversing next subdirectory")

			err = c.traverseDir(nextZipPath, nextDirPath, slashSym, w)
			if err != nil {
				return fmt.Errorf("traverse dir: %w", err)
			}

			continue
		}

		wr, err := w.Create(nextZipPath)
		if err != nil {
			return fmt.Errorf("create zip file writer: %w", err)
		}

		err = copyFileToWriter(nextDirPath, wr)
		if err != nil {
			return fmt.Errorf("copy file to writer: %w", err)
		}

		c.logger.Info().String("path", nextZipPath).Print("Compressed file")
	}

	return nil
}

func copyFileToWriter(filePath string, w io.Writer) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("os open: %w", err)
	}

	_, err = io.Copy(w, f)
	if err != nil {
		return fmt.Errorf("compress file %s: %w", filePath, err)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("file %s close: %w", filePath, err)
	}

	return nil
}
