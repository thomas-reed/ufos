package client

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/thomas-reed/ufos/internal/api"
	"golang.org/x/term"
)

func (c *Client) HandleUFODetails(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	id := fs.String("id", "", "The id of the UFO you want the details for")
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
		return fmt.Errorf("Enter id of UFO you wish to retrieve using '--id' or '-i'")
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

	// Get UFO Metadata and print
	ufoUrl := serverScheme + c.ActivePersona.BaseURL + api.RouteUFOs + "/" + *id
	ufoRes, status, err := ufoSignedRequest[api.UFOMetadataFromHeader](c, http.MethodHead, ufoUrl, nil, nil)
	if err != nil {
		return fmt.Errorf("Error fetching ufo details: %w, (%d)", err, status)
	}
	metadataBytes, err := base64.StdEncoding.DecodeString(string(ufoRes.MetadataBlob))
	if err != nil {
		return fmt.Errorf("Error decoding metadata from header: %w", err)
	}

	if err = c.printUFODetails(metadataBytes); err != nil {
		return err
	}
	return nil
}
