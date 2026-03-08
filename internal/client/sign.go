package client

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/thomas-reed/ufos/internal/crypto"
)

func (c *Client) Sign(req *http.Request, timestamp int64) error {
	payload := fmt.Sprintf("%s|%s|%s|%d", req.Method, req.URL.Path, c.PersonaID, timestamp)

	signature := crypto.SignRequest(c.PersonaData.PrivateKey, []byte(payload))

	req.Header.Set("X-UFO-Persona", c.PersonaID)
	req.Header.Set("X-UFO-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-UFO-Signature", base64.StdEncoding.EncodeToString(signature))

	return nil
}
