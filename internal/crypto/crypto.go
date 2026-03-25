package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

type CryptoSuite uint8

const (
	_             CryptoSuite = iota
	CryptoSuiteV1             // SHA3-256, AES-256-GCM
	// CryptoSuiteV2					// some future packages
)
const (
	CryptoMetadataV1Size = 17 // version (1 byte) + gcm.NonceSize (16 bytes)
)

// Returns a string of the CryptoSuite version info
func (c CryptoSuite) String() string {
	switch c {
	case CryptoSuiteV1:
		return "v1:SHA3-256+AES-256-GCM"
	default:
		return "unknown"
	}
}

// Encrypt encrypts the plaintext using the given CryptoSuite with the provided key.
// It returns the ciphertext prefixed with the unique nonce.
func Encrypt(key, plaintext []byte, version CryptoSuite) (ciphertext []byte, err error) {
	switch version {
	case CryptoSuiteV1:
		ciphertext, err = encryptV1(key, plaintext)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unsupported crypto suite version: %d", version)
	}
	return append([]byte{byte(version)}, ciphertext...), nil
}

// Encrypts the plaintext using cryptosuite V1 (AES-GCM) with the provided key.
// It returns the ciphertext prefixed with the unique nonce.
func encryptV1(key, plaintext []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, err := generateRandom(gcm.NonceSize())
	if err != nil {
		return nil, err
	}

	// Seal appends the authenticated ciphertext to the nonce
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts the ciphertext using AES-GCM with the provided key.
// It returns the plaintext
func Decrypt(key, ciphertext []byte) (plaintext []byte, err error) {
	if len(ciphertext) < 2 {
		return nil, fmt.Errorf("Ciphertext too short")
	}
	version := ciphertext[0]
	payload := ciphertext[1:]

	switch CryptoSuite(version) {
	case CryptoSuiteV1:
		return decryptV1(key, payload)
	default:
		return nil, fmt.Errorf("Unsupported crypto suite version: %d", version)
	}
}

// Decrypts the ciphertext using cryptosuite V1 (AES-GCM) with the provided key.
// It returns the plaintext
func decryptV1(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("Ciphertext too short for nonce size")
	}
	nonce := ciphertext[:nonceSize]
	payload := ciphertext[nonceSize:]

	return gcm.Open(nil, nonce, payload, nil)
}
