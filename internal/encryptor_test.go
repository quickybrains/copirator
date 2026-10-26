package internal

import (
	"testing"

	"github.com/quickybrains/copirator/internal/log/zerolog"
	"github.com/stretchr/testify/require"
)

func TestEncryptor_Smoke(t *testing.T) {
	var test [][]byte = [][]byte{
		[]byte("Hello world!"),
	}

	cfg := EncryptionConfig{
		Enabled: true,
		Key:     "llllllllllllllllllllllllllllllll",
	}
	testEncryptor := NewEncryptor(cfg, zerolog.NewNoop(nil))

	data, err := testEncryptor.Encrypt(test)
	require.NoError(t, err)
	require.Len(t, data, 1)

	actual, err := testEncryptor.TestDecrypt(data)
	require.NoError(t, err)
	require.Len(t, actual, 1)

	require.Equal(t, string(test[0]), string(actual[0]))
}
