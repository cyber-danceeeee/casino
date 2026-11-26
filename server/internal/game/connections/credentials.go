package connections

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"

	"casino/internal/config"
)

func credentialCipher() (cipher.AEAD, error) {
	root, err := base64.StdEncoding.DecodeString(config.Get().AdminAuth.MFAEncryptionKey)
	if err != nil || len(root) != 32 {
		return nil, errors.New("provider credential encryption is not configured")
	}
	derive := hmac.New(sha256.New, root)
	derive.Write([]byte("casino:provider-credentials:v1"))
	block, err := aes.NewCipher(derive.Sum(nil))
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
func seal(code, adapter string, values map[string]string) (string, error) {
	a, err := credentialCipher()
	if err != nil {
		return "", err
	}
	plain, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, a.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(a.Seal(nonce, nonce, plain, []byte(code+":"+adapter))), nil
}
func open(code, adapter, value string) (map[string]string, error) {
	if value == "" {
		return map[string]string{}, nil
	}
	a, err := credentialCipher()
	if err != nil {
		return nil, err
	}
	data, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil || len(data) < a.NonceSize() {
		return nil, errors.New("provider credentials could not be decrypted")
	}
	plain, err := a.Open(nil, data[:a.NonceSize()], data[a.NonceSize():], []byte(code+":"+adapter))
	if err != nil {
		return nil, errors.New("provider credentials could not be decrypted")
	}
	out := map[string]string{}
	if json.Unmarshal(plain, &out) != nil {
		return nil, errors.New("provider credentials could not be decrypted")
	}
	return out, nil
}
