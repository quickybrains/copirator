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
				Path: "./testdata/book.pdf",
			},
			{
				Path: "./testdata/pic.jpg",
			},
		},
	}

	testCmpr := NewCompressor(cfg, zerolog.NewNoop(nil))

	data, err := testCmpr.Compress()
	require.NoError(t, err)
	require.Len(t, data, 1)

	dataR := bytes.NewReader(data[0])

	r, err := zip.NewReader(dataR, int64(len(data)))
	require.NoError(t, err)

	_, err = r.Open("book.pdf")
	require.NoError(t, err)

	_, err = r.Open("pic.jpg")
	require.NoError(t, err)
}
