package crypto

import (
	"crypto/ed25519"
	"crypto/sha3"
	"encoding/base64"

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

func DerivePersonaID(publicKey ed25519.PublicKey, personaName string) string {
	hash := derive(publicKey, "UFOS-PERSONA-ID-V1", personaName)
	return base64.URLEncoding.EncodeToString(hash)
}

func DeriveMasterKey(privateKey ed25519.PrivateKey, personaID string) []byte {
	return derive(privateKey, "UFOS-MASTER-KEY-V1", personaID)
}

func DeriveWrappingKey(masterKey []byte, personaID string) []byte {
	return derive(masterKey, "UFOS-WRAPPING-KEY-V1", personaID)
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

func HashAndBase64(payload []byte) string {
	h := sha3.Sum256(payload)
	return base64.StdEncoding.EncodeToString(h[:])
}
