package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
)

const (
	keyLen   = 32
	nonceLen = 12
	saltLen  = 16
)

func generateRandom(bytesLength int) ([]byte, error) {
	key := make([]byte, bytesLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func GenerateKey() ([]byte, error) {
	key, err := generateRandom(keyLen)
	if err != nil {
		return nil, fmt.Errorf("Error generating key: %w", err)
	}
	return key, nil
}

func GenerateSalt() ([]byte, error) {
	salt, err := generateRandom(saltLen)
	if err != nil {
		return nil, fmt.Errorf("Error generating salt: %w", err)
	}
	return salt, nil
}

func GenerateNonce() ([]byte, error) {
	nonce, err := generateRandom(nonceLen)
	if err != nil {
		return nil, fmt.Errorf("Error generating salt: %w", err)
	}
	return nonce, nil
}

func GenerateAsymPrivateKey() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}
