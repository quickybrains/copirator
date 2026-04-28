package internal

import (
	"time"

	"github.com/quickybrains/copirator/internal/log/zerolog"
)

type LifecycleConfig struct {
	SingleRun      bool          `yaml:"single"`
	BackupInterval time.Duration `yaml:"interval"`
}

type FileConfig struct {
	Path      string `yaml:"path"`
	OutputDir string `yaml:"output_dir"`
}

type CompressionConfig struct {
	Type  string `yaml:"type"`  // "zip", "tar.gz"
	Ratio int    `yaml:"ratio"` // (1-100)
}

type EmailOutputConfig struct {
	Enabled     bool          `yaml:"enabled"`
	FileName    string        `yaml:"file_name"`
	Addr        string        `yaml:"addr"`
	ToAddr      string        `yaml:"to_addr"`
	Subj        string        `yaml:"subj"`
	DialTimeout time.Duration `yaml:"timeout"`
}

type FileOutputConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type OutputConfig struct {
	File  FileOutputConfig  `yaml:"file"`
	Email EmailOutputConfig `yaml:"email"`
}

type LogConfig struct {
	Enabled  bool   `yaml:"enabled"`
	FilePath string `yaml:"path"`
	// other settings for rotation
}

type Config struct {
	// Lifecycle control
	Lifecycle LifecycleConfig `yaml:"lifycycle"`
	// Paths to files to backup
	Files []FileConfig `yaml:"files"`
	// Compression settings
	Compression CompressionConfig `yaml:"compression"`
	// Output settings
	Output OutputConfig `yaml:"output"`
	// Logging settings

	Log zerolog.Config `yaml:"log"`
}
