package client

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/thomas-reed/ufos/internal/crypto"
)

func (c *Client) Sign(
	req *http.Request,
	timestamp, size int64,
	prefixHash, metadata []byte,
) error {
	prefixHashStr := base64.StdEncoding.EncodeToString(prefixHash)
	metadataStr := base64.StdEncoding.EncodeToString(metadata)

	payload := fmt.Sprintf(
		"%s|%s|%s|%d|%d|%s|%s",
		req.Method,
		req.URL.Path,
		c.PersonaID,
		timestamp,
		size,
		prefixHashStr,
		metadataStr,
	)

	signature := crypto.SignRequest(c.PersonaData.PrivateKey, []byte(payload))

	req.Header.Set("X-UFO-Persona", c.PersonaID)
	req.Header.Set("X-UFO-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-UFO-Size", fmt.Sprintf("%d", size))
	req.Header.Set("X-UFO-Prefix", prefixHashStr)
	req.Header.Set("X-UFO-Metadata", metadataStr)
	req.Header.Set("X-UFO-Signature", base64.StdEncoding.EncodeToString(signature))

	return nil
}
