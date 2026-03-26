package crypto

import (
	"crypto/hmac"
	"crypto/sha3"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"os"

	"golang.org/x/crypto/argon2"
)

const argonKeyLen = 32

// Private helper to handle the common derivation logic
func derive(key []byte, domain string, context string) []byte {
	h := sha3.New256()
	h.Write(key)
	h.Write([]byte(domain))
	h.Write([]byte(context))
	return h.Sum(nil)
}

func DerivePersonaID(publicSigningKey []byte, personaName string) string {
	hash := derive(publicSigningKey, "UFOS-PERSONA-ID-V1", personaName)
	return base64.URLEncoding.EncodeToString(hash)
}

func DeriveMasterKey(privateExchangeKey []byte, personaID string) []byte {
	return derive(privateExchangeKey, "UFOS-MASTER-KEY-V1", personaID)
}

func DeriveWrappingKey(baseKey []byte, personaID string) []byte {
	return derive(baseKey, "UFOS-WRAPPING-KEY-V1", personaID)
}

func DeriveSearchSalt(masterKey []byte, personaID string) []byte {
	return derive(masterKey, "UFOS-SEARCH-SALT-V1", personaID)
}

// Derives the vault key from the given password and salt using Argon2id
// Client needs to convert password string to []byte prior to calling (and be sure to clear() it!)
// Need to provide the (t)ime, (m)emory, and (p)arallelism values for Argon2 algorithm
func DeriveVaultKey(password, salt []byte, t, m uint32, p uint8) (key []byte, err error) {
	return argon2.IDKey(password, salt, t, m, p, argonKeyLen), nil
}

// Creates the vault key from the given password using Argon2id
// Client needs to convert password string to []byte prior to calling (and be sure to clear() it!)
// Returns the key and the salt used to generate it.
// Need to provide the (t)ime, (m)emory, and (p)arallelism values for Argon2 algorithm
func CreateVaultKey(password []byte, t, m uint32, p uint8) (key, salt []byte, err error) {
	salt, err = GenerateSalt()
	if err != nil {
		return nil, nil, err
	}
	return argon2.IDKey(password, salt, t, m, p, argonKeyLen), salt, nil
}

// Just a basic hash and base64 encode
func HashAndBase64(payload []byte) string {
	h := sha3.Sum256(payload)
	return base64.StdEncoding.EncodeToString(h[:])
}

// Returns the hmac of a given tag and salt
func HashTag(salt []byte, tag string) string {
	h := hmac.New(func() hash.Hash {
		return sha3.New256()
	}, salt)
	h.Write([]byte(tag))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Returns hash of a file for verifying integrity after download/decrypt
func HashFile(file *os.File) ([]byte, error) {
	h := sha3.New256()
	if _, err := io.Copy(h, file); err != nil {
		return nil, fmt.Errorf("Error copying file for hash")
	}

	file.Seek(0, io.SeekStart) // reset pointer prior to upload
	return h.Sum(nil), nil
}
