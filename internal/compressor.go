package internal

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/quickybrains/copirator/internal/log"
)

var ErrAllFilesSkipped error = errors.New("All files have been skipped")

type Compressor struct {
	cfg         Config
	slashSymbol string

	logger log.Logger
}

func NewCompressor(cfg Config, logger log.Logger) *Compressor {
	c := &Compressor{
		cfg:         cfg,
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

func (c *Compressor) Compress() ([][]byte, error) {
	files, err := c.collectFileInfo()
	if err != nil {
		return nil, fmt.Errorf("collect file info: %w", err)
	}

	if len(files) == 0 {
		c.logger.Info().Print("All files have been skipped")

		return nil, ErrAllFilesSkipped
	}

	chunks := c.devideOnChunks(files)

	data, err := c.compressChunks(chunks)
	if err != nil {
		return nil, fmt.Errorf("compress chunks: %w", err)
	}

	return data, nil
}

type zipFileInfo struct {
	archivePath string
	inputPath   string
	sizeMb      int64
}

type zipChunk []zipFileInfo

func (c *Compressor) compressChunks(input []zipChunk) ([][]byte, error) {
	res := make([][]byte, 0, len(input))
	for _, nextZipChunk := range input {
		// We start new compression, reset
		buff := new(bytes.Buffer)
		w := zip.NewWriter(buff)

		for _, nextZipFile := range nextZipChunk {
			zf, err := w.Create(nextZipFile.archivePath)
			if err != nil {
				return nil, fmt.Errorf("zip file create %s: %w", nextZipFile.archivePath, err)
			}

			err = copyFileToWriter(nextZipFile.inputPath, zf)
			if err != nil {
				return nil, fmt.Errorf("copy file to writer: %w", err)
			}
		}

		res = append(res, buff.Bytes())
	}

	return res, nil
}

func (c *Compressor) devideOnChunks(files []zipFileInfo) []zipChunk {
	res := make([]zipChunk, 0, 10)
	currChunk := make(zipChunk, 0, 10)

	var currSize int64
	for _, nextFile := range files {
		if nextFile.sizeMb > c.cfg.Compression.GetSizeLimitMb() {
			c.logger.Info().
				String("filePath", nextFile.inputPath).
				Int64("fileSizeMb", nextFile.sizeMb).
				Print("Skip file due to its large size")

			continue
		}

		currSize += nextFile.sizeMb
		if currSize > c.cfg.Compression.GetSizeLimitMb() {
			c.logger.Info().
				String("firstFilePath", currChunk[0].inputPath).
				Int64("firstFileSizeMb", currChunk[0].sizeMb).
				String("lastFilePath", currChunk[len(currChunk)-1].inputPath).
				Int64("lastFileSizeMb", currChunk[len(currChunk)-1].sizeMb).
				Int64("fileLimit", c.cfg.Compression.GetSizeLimitMb()).
				Print("Reached file limit, generated chunk")

			// Copying curr chunk content
			chunkCopy := make([]zipFileInfo, len(currChunk))
			copy(chunkCopy, currChunk)
			res = append(res, chunkCopy)

			// Reset curr chunk
			currChunk = currChunk[:0]

			// Not forgetting about current file
			currChunk = append(currChunk, nextFile)
			currSize = nextFile.sizeMb
			continue
		}

		currChunk = append(currChunk, nextFile)
	}

	if len(currChunk) > 0 {
		res = append(res, currChunk)
	}

	c.logger.Info().
		Int("chunksLen", len(res)).
		Int64("fileLimit", c.cfg.Compression.GetSizeLimitMb()).
		Print("Generated chunks")

	return res
}

func (c *Compressor) collectFileInfo() ([]zipFileInfo, error) {
	res := make([]zipFileInfo, 0, 100)
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

			res, err = c.traverseDir(dirName, file.Path, c.slashSymbol, res)
			if err != nil {
				return nil, fmt.Errorf("traverse dir: %w", err)
			}

			continue
		}

		outFileName, err := getFileName(file.OutputDir, file.Path, c.slashSymbol)
		if err != nil {
			return nil, fmt.Errorf("get file name: %w", err)
		}

		fSizeMb := float64(fInfo.Size()) / (1024 * 1024)

		fileInfo := zipFileInfo{
			archivePath: outFileName,
			inputPath:   file.Path,
			sizeMb:      int64(fSizeMb),
		}

		res = append(res, fileInfo)
	}

	c.logger.Info().Int("filesLen", len(res)).Print("Collected files info")

	return res, nil
}

func getFileName(outDir string, filePath string, slashSym string) (string, error) {
	idx := strings.LastIndex(filePath, slashSym)
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

func (c *Compressor) traverseDir(zipPath, dirPath, slashSym string, info []zipFileInfo) ([]zipFileInfo, error) {
	c.logger.Info().String("path", dirPath).Print("Traversing directory")

	dirL, err := os.ReadDir(dirPath)
	if err != nil {
		return info, fmt.Errorf("os read dir: %w", err)
	}

	for _, nextL := range dirL {
		nextDirPath := dirPath + slashSym + nextL.Name()
		nextZipPath := zipPath + slashSym + nextL.Name()
		if nextL.IsDir() {
			c.logger.Info().String("path", nextZipPath).Print("Traversing next subdirectory")

			info, err = c.traverseDir(nextZipPath, nextDirPath, slashSym, info)
			if err != nil {
				return info, fmt.Errorf("traverse dir: %w", err)
			}

			continue
		}

		ext, err := getExtension(nextDirPath)
		if err != nil {
			return info, fmt.Errorf("get extension: %w", err)
		}

		if slices.Contains(c.cfg.Filter.Extension, ext) {
			c.logger.Info().
				String("ext", ext).
				String("filePath", nextDirPath).
				Print("Skipping file due to extension filter")
			continue
		}

		fInfo, err := os.Stat(nextDirPath)
		if err != nil {
			return info, fmt.Errorf("os stat: %w", err)
		}

		fSizeMb := float64(fInfo.Size()) / (1024 * 1024)

		fileInfo := zipFileInfo{
			archivePath: nextZipPath,
			inputPath:   nextDirPath,
			sizeMb:      int64(fSizeMb),
		}

		info = append(info, fileInfo)
	}

	return info, nil
}

func getExtension(name string) (string, error) {
	idx := strings.LastIndex(name, ".")
	if idx == -1 {
		return "", fmt.Errorf("file extension is required, actual %s", name)
	}

	if idx+1 >= len(name) {
		return "", fmt.Errorf("file extension is expected after dot, actual %s", name)
	}

	return name[idx+1:], nil
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
