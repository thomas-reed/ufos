package crypto

import (
	"crypto/ecdh"
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

// For Identity/Signatures (Ed25519)
func GenerateSigningKeyPair() (pub, priv []byte, err error) {
	return ed25519.GenerateKey(rand.Reader)
}

// For Encryption/Exchange (X25519)
func GenerateExchangeKeyPair() (pub, priv []byte, err error) {
	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return key.PublicKey().Bytes(), key.Bytes(), nil
}

// For creating a shared secret for asymmetric encryption
func GenerateSharedSecret(privateKey, publicKey []byte) (secret []byte, err error) {
	privX, err := ecdh.X25519().NewPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	defer func() { privX = nil }() // Sever reference for security

	pubX, err := ecdh.X25519().NewPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	return privX.ECDH(pubX)
}
