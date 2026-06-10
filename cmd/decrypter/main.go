package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/quickybrains/copirator/internal/log/zerolog"
	"github.com/stretchr/testify/assert/yaml"

	copirator "github.com/quickybrains/copirator/internal"
)

func main() {
	cfg, err := readConfig()
	if err != nil {
		log.Fatalf("can't read config, err=%v", err.Error())
	}

	logger := zerolog.New(cfg.Log).With("component", "decryptor")

	decryptor := copirator.NewDecryptor(cfg.Decryption)

	var slashSymbol string
	switch runtime.GOOS {
	case "windows":
		slashSymbol = "\\"
	default:
		slashSymbol = "/"
	}

	for _, nextDir := range cfg.Decryption.Files {
		fInfo, err := os.Stat(nextDir.Path)
		if err != nil {
			logger.Err(err).
				String("path", nextDir.Path).
				Print("Couldn't open dir")
			os.Exit(1)
		}

		if !fInfo.IsDir() {
			logger.Err(err).
				String("path", nextDir.Path).
				Print("Directory is required in path")
			os.Exit(1)
		}

		dirEntry, err := os.ReadDir(nextDir.Path)
		if err != nil {
			logger.Err(err).
				String("path", nextDir.Path).
				Print("Failed to read dir")
			os.Exit(1)
		}

		for _, nextFile := range dirEntry {
			basePath := nextDir.Path + slashSymbol
			filePath := basePath + nextFile.Name()

			fileName, err := getFileName(filePath, slashSymbol)
			if err != nil {
				logger.Err(err).
					String("path", filePath).
					Print("Couldn't get file name for decryption")
				os.Exit(1)
			}

			outPath := basePath + fileName + ".zip"

			outF, err := os.Create(outPath)
			if err != nil {
				logger.Err(err).
					String("path", outPath).
					Print("Failed to create output file")
				os.Exit(1)
			}

			inF, err := os.Open(filePath)
			if err != nil {
				logger.Err(err).
					String("path", filePath).
					Print("Failed to create input file")
				os.Exit(1)
			}

			data, err := io.ReadAll(inF)
			if err != nil {
				logger.Err(err).
					String("path", filePath).
					Print("Failed read input file")
				os.Exit(1)
			}

			decrypted, err := decryptor.Decrypt(data)
			if err != nil {
				logger.Err(err).
					String("path", filePath).
					Print("Failed to decrypt input file")
				os.Exit(1)
			}

			_, err = io.Copy(outF, bytes.NewReader(decrypted))
			if err != nil {
				logger.Err(err).
					String("path", filePath).
					String("outPath", outPath).
					Print("Failed to write output")
				os.Exit(1)
			}

			err = inF.Close()
			if err != nil {
				logger.Err(err).
					String("path", filePath).
					Print("Failed to properly close input file")
			}

			err = outF.Close()
			if err != nil {
				logger.Err(err).
					String("path", outPath).
					Print("Failed to properly close output file")
			}

			logger.Info().
				String("path", filePath).
				String("outPath", outPath).
				Print("File decrypted")
		}
	}
}

func readConfig() (copirator.Config, error) {
	cfgPath := os.Getenv("CONFIG_FILE_PATH")
	if cfgPath == "" {
		cfgPath = "config_decrypt.yaml"
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

func getFileName(src, slashSym string) (string, error) {
	slashIdx := strings.LastIndex(src, slashSym)
	if slashIdx == -1 {
		return "", fmt.Errorf("file path is incorrect, no slash symbol, path: %s", src)
	}

	dotIdx := strings.LastIndex(src, ".")
	if dotIdx == -1 {
		return "", fmt.Errorf("file extension is required, file name: %s", src)
	}

	return src[slashIdx+1 : dotIdx], nil
}
