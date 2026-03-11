package client

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/thomas-reed/ufos/internal/crypto"
)

func (c *Client) Sign(r *http.Request, timestamp int64) error {
	bodyHash := ""
	if !(r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/api/objects/")) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("Couldn't read request body: %w", err)
		}
		bodyHash = crypto.HashAndBase64(body)
		// Re-stuff the body so HandleCreateUFO can decode the JSON
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	payload := fmt.Sprintf(
		"%s|%s|%s|%d|%s",
		r.Method,
		r.URL.Path,
		c.PersonaID,
		timestamp,
		bodyHash,
	)

	signature := crypto.SignRequest(c.PersonaData.PrivateKey, []byte(payload))

	r.Header.Set("X-UFO-Persona", c.PersonaID)
	r.Header.Set("X-UFO-Timestamp", fmt.Sprintf("%d", timestamp))
	r.Header.Set("X-UFO-Signature", base64.StdEncoding.EncodeToString(signature))

	return nil
}
