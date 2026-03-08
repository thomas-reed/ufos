package crypto

import (
	"crypto/ed25519"
)

// SignRequest creates a digital signature for a given payload.
func SignRequest(priv ed25519.PrivateKey, payload []byte) []byte {
	return ed25519.Sign(priv, payload)
}

// VerifyRequest confirms that a signature is valid for a given payload.
func VerifyRequest(pub ed25519.PublicKey, payload, sig []byte) bool {
	return ed25519.Verify(pub, payload, sig)
}
