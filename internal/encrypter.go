package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/quickybrains/copirator/internal/log"
)

type Encryptor struct {
	cfg    EncryptionConfig
	logger log.Logger
}

func NewEncryptor(cfg EncryptionConfig, logger log.Logger) *Encryptor {
	return &Encryptor{
		cfg:    cfg,
		logger: logger.With("component", "encrypter"),
	}
}

func (e *Encryptor) Encrypt(src [][]byte) ([][]byte, error) {
	if !e.cfg.Enabled {
		e.logger.Info().Print("Encrypter is disabled")

		return src, nil
	}

	res := make([][]byte, 0, len(src))

	for _, data := range src {
		ciphertext, err := encryptGCM(data, []byte(e.cfg.Key))
		if err != nil {
			return nil, fmt.Errorf("encrypt gcm: %w", err)
		}

		res = append(res, []byte(base64.StdEncoding.EncodeToString(ciphertext)))
	}

	return res, nil
}

func (e *Encryptor) TestDecrypt(src [][]byte) ([][]byte, error) {
	res := make([][]byte, 0, len(src))
	for _, data := range src {
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return nil, fmt.Errorf("base64 decode: %w", err)
		}

		decrypted, err := decryptGCM(decoded, []byte(e.cfg.Key))
		if err != nil {
			return nil, fmt.Errorf("decrypt gcm: %w", err)
		}

		res = append(res, decrypted)
	}

	return res, nil
}

func encryptGCM(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptGCM(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
