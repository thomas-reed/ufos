package client

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/thomas-reed/ufos/internal/api"
	"golang.org/x/term"
)

func (c *Client) HandleOrbitRemove(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	id := fs.String("id", "", "The Persona ID of the user you want to remove")
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
		return fmt.Errorf("Enter Persona ID of the user you wish to remove using '--id' or '-i'")
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

	// Remove satellite from orbit
	url := serverScheme + c.ActivePersona.BaseURL + api.RouteOrbit + "/" + *id
	_, status, err := ufoSignedRequest[api.EmptyResponse](c, http.MethodDelete, url, nil, nil)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("Unexpected response status code (%d) for Persona ID %s", status, *id)
	}

	fmt.Printf("Satellite %s has been removed from your orbit.\n", *id)
	return nil
}
