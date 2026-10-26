package internal

import (
	"encoding/base64"
	"fmt"
)

type Decryptor struct {
	cfg DecryptionConfig
}

func NewDecryptor(cfg DecryptionConfig) *Decryptor {
	return &Decryptor{
		cfg: cfg,
	}
}

func (e *Decryptor) Decrypt(src []byte) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(src))
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	decrypted, err := decryptGCM(decoded, []byte(e.cfg.Key))
	if err != nil {
		return nil, fmt.Errorf("decrypt gcm: %w", err)
	}

	return decrypted, nil
}
