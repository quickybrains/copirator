package internal

import (
	"os"
	"testing"

	"github.com/quickybrains/copirator/internal/log/zerolog"
	"github.com/stretchr/testify/require"
)

func TestEmailWriter_Smoke(t *testing.T) {
	// Turn on for test
	cfg := EmailOutputConfig{
		Enabled: false,
		Addr:    "smtp.gmail.com",
		// Need to use temp email
		ToAddr:   "quickybrains@gmail.com",
		Subj:     "Backup",
		FileName: "backup.zip",
	}

	// Need to provide credentials for an existing email box for sending
	err := os.Setenv("EMAIL_USERNAME", "quickybrains@gmail.com")
	require.NoError(t, err)

	err = os.Setenv("EMAIL_PASSWORD", "izfo mode afjy agjw")
	require.NoError(t, err)

	testEmailWriter, err := NewEmailWriter(cfg, zerolog.NewNoop(nil))
	require.NoError(t, err)

	// testData := "Some simple message from an outer world.\nHello, Earth!"

	cmprCfg := Config{
		Files: []FileConfig{
			{
				Path: "./testdata/book.pdf",
			},
			{
				Path: "./testdata/pic.jpg",
			},
		},
	}

	testCmpr := NewCompressor(cmprCfg, zerolog.NewNoop(nil))

	testData, err := testCmpr.Compress()
	require.NoError(t, err)

	err = testEmailWriter.Write([]byte(testData))
	require.NoError(t, err)
}
