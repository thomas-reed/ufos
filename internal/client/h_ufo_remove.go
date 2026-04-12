package client

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/crypto"
	"github.com/thomas-reed/ufos/internal/objects"
	"golang.org/x/term"
)

func (c *Client) HandleRemoveUFO(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("info", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	id := fs.String("id", "", "The id of the UFO you want to remove")
	fs.StringVar(id, "i", "", "alias for --id")

	if err := fs.Parse(cmd.Args); err != nil {
		return err
	}

	// If name wasn't in Args, prompt
	if *name == "" {
		n, err := getInput("your persona name")
		if err != nil {
			return err
		}
		name = &n
	}

	// If id wasn't in Args, error out
	if *id == "" {
		return fmt.Errorf("Enter id of UFO you wish to remove using '--id' or '-i'")
	}

	// get master password to decrypt vault, find persona
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

	// Generate salt for prefix hashing
	searchSalt := crypto.DeriveSearchSalt(c.MasterKey, c.PersonaID)
	defer clear(searchSalt)

	// Get UFO Metadata to get the plaintext Prefix
	ufoUrl := c.ActivePersona.BaseURL + api.RouteUFOs + "/" + *id
	ufoRes, _, err := ufoSignedRequest[api.UFOMetadataFromHeader](c, http.MethodHead, ufoUrl, nil, nil)

	metadataBytes, err := base64.StdEncoding.DecodeString(string(ufoRes.MetadataBlob))
	if err != nil {
		return fmt.Errorf("Error decoding metadata: %w", err)
	}
	metadata, err := crypto.Decrypt(c.MasterKey, metadataBytes)
	if err != nil {
		return err
	}

	var ufoMeta objects.ObjectMetadata
	if err = json.Unmarshal(metadata, &ufoMeta); err != nil {
		return fmt.Errorf("Error unmarshalling metadata: %w", err)
	}

	// Build list of IDs to remove
	idList := []string{*id}

	// Desired UFO is a folder - alert the user for recursive removal
	if ufoMeta.SizeBytes < 0 {
		ufoConfirm, err := getInput("your persona name")
		if err != nil {
			return err
		}
		
		if *id != ufoConfirm {
			return fmt.Errorf("Confirmation ID does not match! Cancelling request")
		}
		// Add child IDs to the list
		children, err := c.getRecursiveIDs(ufoRes, searchSalt)
		if err != nil {
			return fmt.Errorf("Error getting child IDs: %w", err)
		}
		idList = append(idList, children...)
	}

	for _, ufoID := range idList {
		fmt.Printf("Removing UFO %s... ", ufoID)
		url := c.ActivePersona.BaseURL + api.RouteUFOs + "/" + ufoID
		_, statusCode, err := ufoSignedRequest[api.EmptyResponse](c, http.MethodDelete, url, nil, nil)
		if err != nil {
			return err
		}
		if statusCode != http.StatusNoContent {
			return fmt.Errorf("Unexpected response status code (%d) for UFO ID %s", statusCode, ufoID)
		}
		fmt.Println("done")
	}

	return nil
}
