package main

import (
	"crypto/hkdf"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	jose "github.com/square/go-jose/v3"
)

// getDerivedEncryptionKey deriva a chave usando HKDF com SHA-256
func getDerivedEncryptionKey(secret, salt string) ([]byte, error) {
	// Para A256GCM precisamos de uma chave de 32 bytes
	hkdf := hkdf.New(sha256.New, []byte(secret), []byte(salt), nil)
	key := make([]byte, 32)
	if _, err := io.ReadFull(hkdf, key); err != nil {
		return nil, err
	}
	return key, nil
}

// decodeToken realiza a decodificação e validação do token JWE
func decodeToken(token, secret, salt string) (map[string]interface{}, error) {
	// Deriva a chave de encriptação
	key, err := getDerivedEncryptionKey(secret, salt)
	if err != nil {
		return nil, fmt.Errorf("erro ao derivar chave: %v", err)
	}

	// Faz o parse do token criptografado (JWE)
	jwe, err := jose.ParseEncrypted(token)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear token: %v", err)
	}

	// Decripta o token usando a chave derivada
	decrypted, err := jwe.Decrypt(key)
	if err != nil {
		return nil, fmt.Errorf("erro ao decriptar token: %v", err)
	}

	// Decodifica os claims (payload JSON)
	var claims map[string]interface{}
	if err := json.Unmarshal(decrypted, &claims); err != nil {
		return nil, fmt.Errorf("erro ao decodificar payload: %v", err)
	}

	// Valida o claim de expiração ("exp")
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, errors.New("token expirado")
		}
	} else {
		return nil, errors.New("claim de expiração não encontrado")
	}

	// Você pode adicionar validações extras, como 'iat', 'jti', etc.
	return claims, nil
}

func main() {
	// Exemplo de uso:
	token := "seu-token-jwe-aqui" // substitua pelo token gerado
	secret := "seu-segredo"
	salt := "" // conforme o código original, salt vazio indica token de sessão

	claims, err := decodeToken(token, secret, salt)
	if err != nil {
		log.Fatalf("Erro ao decodificar token: %v", err)
	}

	fmt.Printf("Token decodificado e válido: %+v\n", claims)
}
