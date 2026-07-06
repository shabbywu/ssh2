package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ssh2/utils"
)

const Prefix = "go1$"

func Encrypt(text string) (string, error) {
	if strings.HasPrefix(text, Prefix) {
		return text, nil
	}

	key, err := secretKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	payload := append(nonce, gcm.Seal(nil, nonce, []byte(text), nil)...)
	return Prefix + base64.StdEncoding.EncodeToString(payload), nil
}

func Decrypt(text string) (string, error) {
	if text == "" {
		return "", nil
	}
	if strings.HasPrefix(text, "crypt$") {
		return "", errors.New("legacy Python crypt$ credentials are not supported; re-import this credential with the Go CLI")
	}
	if !strings.HasPrefix(text, Prefix) {
		return text, nil
	}

	key, err := secretKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	encoded := strings.TrimPrefix(text, Prefix)
	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("invalid encrypted credential payload: %w", err)
	}
	if len(payload) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted credential payload: missing nonce")
	}

	nonce := payload[:gcm.NonceSize()]
	ciphertext := payload[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt credential: %w", err)
	}
	return string(plaintext), nil
}

func secretKey() ([]byte, error) {
	if value := os.Getenv("SSH2_SECRET_KEY"); value != "" {
		return normalizeKey(value), nil
	}

	if err := os.MkdirAll(utils.SSH2_HOME, 0700); err != nil {
		return nil, err
	}

	path := filepath.Join(utils.SSH2_HOME, "secret.key")
	if data, err := os.ReadFile(path); err == nil {
		return normalizeKey(strings.TrimSpace(string(data))), nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	value := base64.StdEncoding.EncodeToString(key)
	if err := os.WriteFile(path, []byte(value), 0600); err != nil {
		return nil, err
	}
	return key, nil
}

func normalizeKey(value string) []byte {
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil && len(decoded) == 32 {
		return decoded
	}
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}
