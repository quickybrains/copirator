package internal

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/quickybrains/copirator/internal/log/zerolog"
	"github.com/stretchr/testify/require"
)

func TestCompressor(t *testing.T) {
	cfg := Config{
		Files: []FileConfig{
			{
				Path: "./testdata/encrypt",
			},
		},
	}

	testCmpr := NewCompressor(cfg, zerolog.NewNoop(nil))

	data, err := testCmpr.Compress()
	require.NoError(t, err)

	dataR := bytes.NewReader(data[0])

	r, err := zip.NewReader(dataR, int64(len(data[0])))
	require.NoError(t, err)

	_, err = r.Open("encrypt/book.epub")
	require.NoError(t, err)
}
