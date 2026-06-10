package internal

import (
	"testing"

	"github.com/quickybrains/copirator/internal/log/zerolog"
	"github.com/stretchr/testify/require"
)

func TestCompressor(t *testing.T) {
	cfg := Config{
		Files: []FileConfig{
			{
				Path: "./testdata",
			},
		},
	}

	testCmpr := NewCompressor(cfg, zerolog.NewNoop(nil))

	_, err := testCmpr.Compress()
	require.ErrorIs(t, err, ErrAllFilesSkipped)
}
