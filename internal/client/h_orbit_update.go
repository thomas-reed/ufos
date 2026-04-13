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

func (c *Client) HandleOrbitUpdate(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("orbit update", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	id := fs.String("id", "", "The Persona ID of the user you want to update")
	fs.StringVar(id, "i", "", "alias for --id")

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

	// If id wasn't in Args, error out
	if *id == "" {
		return fmt.Errorf("Enter Persona ID of the user you wish to update with '--id' or '-i'")
	}

	// Get master password to decrypt vault, find persona
	fmt.Printf("Enter master password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("Error reading password: %w", err)
	}
	defer clear(password)
	err = c.GetPersonaFromVault(*name, password)
	if err != nil {
		return err
	}
	defer clear(c.ActivePersona.PrivateSigningKey)
	defer clear(c.ActivePersona.PrivateExchangeKey)
	defer clear(c.MasterKey)

	// Get existing satellite and decrypt metadata
	satUrl := c.ActivePersona.BaseURL + api.RouteOrbit + "/" + *id
	sat, _, err := ufoSignedRequest[api.Satellite](c, http.MethodGet, satUrl, nil, nil)
	satMetaBytes, err := crypto.Decrypt(c.MasterKey, sat.Metadata)
	if err != nil {
		return err
	}
	defer clear(satMetaBytes)
	var satMetadata contacts.ContactMetadata
	if err = json.Unmarshal(satMetaBytes, &satMetadata); err != nil {
		return fmt.Errorf("Error unmarshalling satellite metadata: %w", err)
	}

	// Build Contact metadata
	contact, err := buildContact(satMetadata)
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

	fmt.Printf("%s has been added to your orbit!", sat.PersonaID)
	return nil
}
