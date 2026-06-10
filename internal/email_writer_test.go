package internal

import (
	"testing"

	"github.com/quickybrains/copirator/internal/log/zerolog"
	"github.com/stretchr/testify/require"
)

func TestEmailWriter_Smoke(t *testing.T) {
	// Turn on for test
	cfg := EmailOutputConfig{
		Enabled: true,
		Addr:    "smtp.gmail.com",
		// Need to use temp email
		ToAddr: "quickybrains@gmail.com",
		// Creds
		Username: "quickybrains@gmail.com",
		ApiKey:   "izfo mode afjy agjw",
		Subj:     "Backup",
		FileName: "backup.zip",
	}

	testEmailWriter, err := NewEmailWriter(cfg, zerolog.NewNoop(nil))
	require.NoError(t, err)

	fileCfg := FileOutputConfig{
		Enabled: true,
		Path:    "./testdata/result.zip",
	}
	testFileWriter := NewFileWriter(fileCfg, zerolog.NewNoop(nil))

	cmprCfg := Config{
		Compression: CompressionConfig{
			SizeLimitMb: 15,
		},
		Files: []FileConfig{
			{
				Path: "./testdata",
			},
		},
	}

	testCmpr := NewCompressor(cmprCfg, zerolog.NewNoop(nil))

	testData, err := testCmpr.Compress()
	require.NoError(t, err)

	err = testFileWriter.Write(testData)
	require.NoError(t, err)

	err = testEmailWriter.Write(testData)
	require.NoError(t, err)
}
