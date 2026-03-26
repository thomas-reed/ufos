package crypto

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
)

const (
	keyLen   = 32 // for 256-bit keys
	nonceLen = 12 // for AES-GCM
	saltLen  = 16 // for KDF
	ivLen    = 16 // for AES-CTR
)

func generateRandom(bytesLength int) ([]byte, error) {
	key := make([]byte, bytesLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("Error reading random bytes: %w", err)
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

func GenerateIV() ([]byte, error) {
	iv, err := generateRandom(ivLen)
	if err != nil {
		return nil, fmt.Errorf("Error generating initialization vector: %w", err)
	}
	return iv, nil
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
		return nil, nil, fmt.Errorf("Error generating asymmetric key pair: %w", err)
	}
	return key.PublicKey().Bytes(), key.Bytes(), nil
}

// For creating a shared secret for asymmetric encryption
func GenerateSharedSecret(privateKey, publicKey []byte) (secret []byte, err error) {
	privX, err := ecdh.X25519().NewPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("Error generating private key from []byte: %w", err)
	}
	defer func() { privX = nil }() // Sever reference for security

	pubX, err := ecdh.X25519().NewPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("Error generating public key from []byte: %w", err)
	}

	return privX.ECDH(pubX)
}
