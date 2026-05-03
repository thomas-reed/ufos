package client

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/contacts"
	"github.com/thomas-reed/ufos/internal/crypto"
	"golang.org/x/term"
)

func (c *Client) HandleOrbitAdd(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")

	if err := fs.Parse(cmd.Args); err != nil {
		return err
	}

	// If name wasn't in Args, prompt
	if *name == "" {
		n, err := getInput("your persona name", true)
		if err != nil {
			return err
		}
		name = &n
	}

	// Get vault password to decrypt vault, find persona
	fmt.Print("Enter your vault password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("Error reading password: %w", err)
	}
	defer clear(password)
	fmt.Println()
	err = c.GetPersonaFromVault(*name, password)
	if err != nil {
		return err
	}
	defer clear(c.ActivePersona.PrivateSigningKey)
	defer clear(c.ActivePersona.PrivateExchangeKey)
	defer clear(c.MasterKey)

	// Build Contact metadata
	contact, err := buildContact(contacts.ContactMetadata{})
	if err != nil {
		return err
	}

	// Encrypt user metadata
	contactMetaBytes, err := json.Marshal(contact)
	if err != nil {
		return fmt.Errorf("Error marshalling contact metadata: %w", err)
	}
	metaBlob, err := crypto.Encrypt(c.MasterKey, contactMetaBytes, crypto.CryptoSuiteV1)
	if err != nil {
		return err
	}

	// Get public keys of user
	url := serverScheme + contact.Domain + api.RoutePersonas + "/" + contact.PersonaID
	keys, _, err := ufoPublicRequest[api.PersonaKeysResponse](c, http.MethodGet, url, nil, nil)
	if err != nil {
		return err
	}

	// Save user to orbit and send
	orbitUrl := c.ActivePersona.BaseURL + api.RouteOrbit
	req := api.OrbitMetadataRequest{
		PersonaID:   keys.PersonaID,
		SigningKey:  keys.SigningKey,
		ExchangeKey: keys.ExchangeKey,
		Metadata:    metaBlob,
	}

	sat, status, err := ufoSignedRequest[api.Satellite](c, "POST", orbitUrl, req, nil)
	if err != nil {
		return err
	}
	if status != http.StatusCreated && status != http.StatusOK {
		return fmt.Errorf("Unexpected status code (%d)", status)
	}

	fmt.Printf("%s has been added to your orbit.\n", sat.PersonaID)
	return nil
}
