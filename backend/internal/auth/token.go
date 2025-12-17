package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func NewToken() (string, error) {

	buff := make([]byte, 32)
	n, err := rand.Read(buff)
	if err != nil {
		return "", nil
	}
	if n != 32 {
		return "", fmt.Errorf("expected 32 random bytes, got %d", n)

	}

	token := base64.RawURLEncoding.EncodeToString(buff)

	return token, nil

}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
