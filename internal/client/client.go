package client

import (
	"net/http"
	"time"
)

const (
	clientTimeout = 10 * time.Second
	ServerScheme  = "http://" // just for now (dev) - need to figure out letsEncrypt certs
)

type Client struct {
	HTTPClient    *http.Client // for sending the requests
	ActivePersona *Persona     // name, baseURL, private key
	PersonaID     string       // derived from persona's public key and name
	MasterKey     []byte       // derived from persona's private key and ID for file encryption
}

// NewClient initializes a new UFO client with standard defaults.
func NewClient() (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{
			Timeout: clientTimeout,
		},
	}

	return &c, nil
}
