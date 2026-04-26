package client

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/thomas-reed/ufos/internal/crypto"
)

func (c *Client) Sign(
	r *http.Request,
	timestamp int64,
	requiresBodyHash bool,
) error {
	bodyHash := ""
	if requiresBodyHash {
		var body []byte
		var err error

		if r.Body != nil {
			body, err = io.ReadAll(r.Body)
			if err != nil {
				return fmt.Errorf("couldn't read request body: %w", err)
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		bodyHash = crypto.HashAndBase64(body)
	}

	timestampStr := fmt.Sprintf("%d", timestamp)

	payload := fmt.Sprintf(
		"%s|%s|%s|%s|%s",
		r.Method,
		r.URL.Path,
		c.PersonaID,
		timestampStr,
		bodyHash,
	)

	signature := crypto.SignRequest(
		ed25519.PrivateKey(c.ActivePersona.PrivateSigningKey),
		[]byte(payload),
	)

	r.Header.Set("X-UFO-Persona", c.PersonaID)
	r.Header.Set("X-UFO-Timestamp", fmt.Sprintf("%d", timestamp))
	r.Header.Set("X-UFO-Signature", base64.StdEncoding.EncodeToString(signature))

	return nil
}
