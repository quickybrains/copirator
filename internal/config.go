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

type FilterConfig struct {
	Extension []string `yaml:"ext"`
}

const defaultSizeLimitMb = 15

type CompressionConfig struct {
	SizeLimitMb int64  `yaml:"limit_mb"`
	Type        string `yaml:"type"`  // "zip", "tar.gz"
	Ratio       int    `yaml:"ratio"` // (1-100)
}

func (cc CompressionConfig) GetSizeLimitMb() int64 {
	if cc.SizeLimitMb <= 0 {
		return defaultSizeLimitMb
	}

	return cc.SizeLimitMb
}

type EmailOutputConfig struct {
	Enabled     bool          `yaml:"enabled"`
	FileName    string        `yaml:"file_name"`
	Addr        string        `yaml:"addr"`
	ToAddr      string        `yaml:"to_addr"`
	Subj        string        `yaml:"subj"`
	Username    string        `yaml:"username"`
	ApiKey      string        `yaml:"api_key"`
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
	// Files to filter
	Filter FilterConfig
	// Compression settings
	Compression CompressionConfig `yaml:"compression"`
	// Output settings
	Output OutputConfig `yaml:"output"`
	// Logging settings

	Log zerolog.Config `yaml:"log"`
}
