package client

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/thomas-reed/ufos/internal/api"
	"golang.org/x/term"
)

func (c *Client) HandleNewPersona(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("new", flag.ContinueOnError)

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

	// Send request and print result
	url := c.ActivePersona.BaseURL + api.RouteInit
	token, status, err := ufoSignedRequest[api.InitPersonaResponse](c, http.MethodPost, url, nil, nil)
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("Unexpected status code (%d)", status)
	}

	fmt.Printf("Generated Registration Token: %s\n", token.RegistrationToken)
	fmt.Println("This token is only valid for 15 minutes.")
	fmt.Println("Use 'ufos register' command to create a new persona for yourself, or share it with your friend to let them use this server!")

	token = api.InitPersonaResponse{}
	return nil
}
