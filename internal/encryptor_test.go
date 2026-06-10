package internal

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
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

func TestDecryptor_Smoke(t *testing.T) {
	path := "./testdata/decrypt/test_2_2026-06-10 20_30_08.txt"
	outputPath := "./testdata/decrypt/test_2_result.zip"

	f, err := os.Open(path)
	require.NoError(t, err)

	data, err := io.ReadAll(f)
	require.NoError(t, err)

	cfg := EncryptionConfig{
		Enabled: true,
		Key:     "ponka14ponka14ponka14ponka14ponk",
	}
	testEncryptor := NewEncryptor(cfg, zerolog.NewNoop(nil))

	decrypted, err := testEncryptor.TestDecrypt([][]byte{data})
	require.NoError(t, err)

	outF, err := os.Create(outputPath)
	require.NoError(t, err)

	dataR := bytes.NewReader(decrypted[0])

	_, err = io.Copy(outF, dataR)
	require.NoError(t, err)

	r, err := zip.NewReader(dataR, int64(len(decrypted[0])))
	require.NoError(t, err)

	_, err = r.Open("testdata/documents/cv.xlsx")
	require.NoError(t, err)
}
