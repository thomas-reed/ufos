package client

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/thomas-reed/ufos/internal/api"
	"golang.org/x/term"
)

func (c *Client) HandleCreatePersona(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("register", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona to register")
	fs.StringVar(name, "n", "", "alias for --name")
	domain := fs.String("domain", "", "The domain of the UFOs server you want to register to")
	fs.StringVar(domain, "d", "", "alias for --domain")
	token := fs.String("token", "", "The registration token from running the 'new' command, (or getting the Initial Bootstrap Token from server logs)")
	fs.StringVar(token, "t", "", "alias for --token")
	if err := fs.Parse(cmd.Args); err != nil {
		return err
	}

	// If name wasn't in Args, prompt
	if *name == "" {
		n, err := getInput("your deisred persona name", true)
		if err != nil {
			return err
		}
		name = &n
	}
	// If domain wasn't in Args, prompt
	if *domain == "" {
		d, err := getInput("the UFOs domain", true)
		if err != nil {
			return err
		}
		formatDomain(&d)
		domain = &d
	}
	

	// If token wasn't in Args, prompt
	if *token == "" {
		t, err := getInput("registration token", true)
		if err != nil {
			return err
		}
		token = &t
	}

	// Get master password to decrypt vault
	fmt.Print("Enter master password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("Error reading password: %w", err)
	}
	defer clear(password)
	fmt.Println()

	// Check to see if the persona exists in the vault (making this a retry registration)
	err = c.GetPersonaFromVault(*name, password)
	if errors.Is(err, ErrPersonaNotFound) {
		// In case of an error, create the new persona and load it into the client
		err = c.AddPersonaToVault(*name, *domain, password)
		if err != nil {
			return fmt.Errorf("Error writing new persona to vault: %w", err)
		}
		err = c.GetPersonaFromVault(*name, password)
		if err != nil {
			return fmt.Errorf("Error retrieving new persona from vault: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("Error getting persona from vault: %w", err)
	}
	defer clear(c.ActivePersona.PrivateSigningKey)
	defer clear(c.ActivePersona.PrivateExchangeKey)
	defer clear(c.MasterKey)

	body := api.NewPersonaRequest{
		ID:          c.PersonaID,
		SigningKey:  c.ActivePersona.PublicSigningKey,
		ExchangeKey: c.ActivePersona.PublicExchangeKey,
	}
	url := serverScheme + *domain + api.RoutePersonas
	header := map[string]string{
		api.HeaderRegistration: *token,
	}
	res, status, err := ufoPublicRequest[api.CreatePersonaResponse](
		c,
		http.MethodPost,
		url,
		body,
		header,
	)
	if err != nil {
		return fmt.Errorf("Error creating persona: %w (%d)", err, status)
	}

	if res.ID == c.PersonaID {
		fmt.Printf("Persona '%s' has been registered on %s.\n", res.ID, *domain)
	}
	return nil
}
