package client

import (
	"net/http"
	"time"
)

const (
	clientTimeout = 10 * time.Second
)

type Client struct {
	HTTPClient  *http.Client // for sending the requests
	VaultPath   string       // location of the local persona data
	PersonaData Persona      // name, baseURL, private key
	PersonaID   string       // derived from persona's public key and name
	MasterKey   []byte       // derived from persona's private key and ID for file encryption
}

// NewClient initializes a new UFO client with standard defaults.
func NewClient(password []byte, personaName string) (*Client, error) {
	client := Client{
		HTTPClient: &http.Client{
			Timeout: clientTimeout,
		},
		PersonaData: Persona{
			Name: personaName,
		},
	}
	err := client.getVaultFilepath()
	if err != nil {
		return nil, err
	}
	return &client, nil
}
