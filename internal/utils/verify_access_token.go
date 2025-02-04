package utils

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/square/go-jose/v3"
	"golang.org/x/crypto/hkdf"
)

type CreateAccessTokenUtil struct{}

func NewCreateAccessTokenUtil() *CreateAccessTokenUtil {
	return &CreateAccessTokenUtil{}
}

func (b *CreateAccessTokenUtil) DecodeToken(token string) (map[string]interface{}, error) {
	encryptionKey, err := getDerivedEncryptionKey([]byte(os.Getenv("SECRET_JWT")), "")
	if err != nil {
		return nil, err
	}

	payload, err := decodeToken(token, encryptionKey)
	if err != nil {
		return nil, err
	}

	if err := validateClaims(payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func getDerivedEncryptionKey(keyMaterial []byte, salt string) ([]byte, error) {
	info := []byte("NextAuth.js Generated Encryption Key")
	if salt != "" {
		info = []byte(fmt.Sprintf("NextAuth.js Generated Encryption Key (%s)", salt))
	}
	h := hkdf.New(sha256.New, keyMaterial, []byte(salt), info)
	key := make([]byte, 32)
	if _, err := io.ReadFull(h, key); err != nil {
		return nil, err
	}
	return key, nil
}

func decodeToken(tokenStr string, encryptionKey []byte) (map[string]interface{}, error) {
	jweObject, err := jose.ParseEncrypted(tokenStr)
	if err != nil {
		return nil, err
	}
	decrypted, err := jweObject.Decrypt(encryptionKey)
	if err != nil {
		return nil, err
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(decrypted, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func validateClaims(payload map[string]interface{}) error {
	now := time.Now().Unix()

	if exp, ok := payload["exp"].(float64); ok {
		if now > int64(exp) {
			return errors.New("token expirado")
		}
	}

	if iat, ok := payload["iat"].(float64); ok {
		if now < int64(iat) {
			return errors.New("token não é válido ainda")
		}
	}

	return nil
}
