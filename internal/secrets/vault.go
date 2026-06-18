package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const prefix = "enc:v1:"

type Vault struct {
	key []byte
}

func Open() (*Vault, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	appDir := filepath.Join(dir, "RestDeck")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return nil, err
	}
	keyPath := filepath.Join(appDir, "secret.key")
	key, err := os.ReadFile(keyPath)
	if errors.Is(err, os.ErrNotExist) {
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		if err := os.WriteFile(keyPath, key, 0o600); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid RestDeck local secret key length")
	}
	return &Vault{key: key}, nil
}

func (v *Vault) Seal(value string) (string, error) {
	if value == "" || strings.HasPrefix(value, prefix) {
		return value, nil
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(value), nil)
	payload := append(nonce, ciphertext...)
	return prefix + base64.RawURLEncoding.EncodeToString(payload), nil
}

func (v *Vault) Open(value string) (string, error) {
	if value == "" || !strings.HasPrefix(value, prefix) {
		return value, nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(value, prefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", fmt.Errorf("encrypted value is too short")
	}
	nonce := payload[:gcm.NonceSize()]
	ciphertext := payload[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
