package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/thomas-reed/ufos/internal/api"
	"golang.org/x/term"
)

func (c *Client) HandleCreatePersona(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("register", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona to register")
	fs.StringVar(name, "n", "", "alias for --name")
	token := fs.String("token", "", "The registration token from running the 'new' command, (or getting the Initial Bootstrap Token from server logs)")
	fs.StringVar(token, "t", "", "alias for --token")
	if err := fs.Parse(cmd.Args); err != nil {
		return err
	}

	// If name wasn't in Args, prompt
	if *name == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter desired persona name > ")
		if !scanner.Scan() {
			return fmt.Errorf("Input interrupted!")
		}
		n := scanner.Text()
		name = &n
	}
	// get Domain from name, or prompt
	splitName := strings.Split(*name, "@")
	domain := ""
	switch len(splitName) {
	case 1:
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter your desired UFOs domain > ")
		if !scanner.Scan() {
			return fmt.Errorf("Input interrupted!")
		}
		domain = scanner.Text()
	case 2:
		domain = splitName[1]
	default:
		return fmt.Errorf("Error parsing domain from given name")
	}

	// if token wasn't in Args, prompt
	if *token == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter registration token > ")
		if !scanner.Scan() {
			return fmt.Errorf("Input interrupted!")
		}
		t := scanner.Text()
		token = &t
	}

	// get master password to decrypt vault
	fmt.Printf("Enter master password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("Error reading password: %w", err)
	}
	defer clear(pw)

	// Check to see if the persona exists in the vault (making this a retry registration)
	err = c.GetPersonaFromVault(*name, pw)
	if errors.Is(err, ErrPersonaNotFound) {
		// In case of an error, create the new persona and load it into the client
		err = c.AddPersonaToVault(*name, domain, pw)
		if err != nil {
			return fmt.Errorf("Error writing new persona to vault: %w", err)
		}
		err = c.GetPersonaFromVault(*name, pw)
		if err != nil {
			return fmt.Errorf("Error retrieving new persona from vault: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("Error getting persona from vault: %w", err)
	}
	defer clear(c.ActivePersona.PrivateSigningKey)
	defer clear(c.ActivePersona.PrivateExchangeKey)
	defer clear(c.MasterKey)

	data, err := json.Marshal(api.NewPersonaRequest{
		ID:          c.PersonaID,
		SigningKey:  c.ActivePersona.PublicSigningKey,
		ExchangeKey: c.ActivePersona.PublicExchangeKey,
	})
	body := bytes.NewReader(data)
	url := domain + api.RouteRegister
	header := map[string]string{
		api.HeaderRegistration: *token,
	}
	res, _, err := ufoSignedRequest[api.CreatePersonaResponse](
		c,
		http.MethodPost,
		url,
		body,
		header,
	)
	if err != nil {
		return fmt.Errorf("Error creating persona: %w", err)
	}

	if res.ID == c.PersonaID {
		fmt.Printf("Persona '%s' has been registered on %s", res.ID, domain)
	}
	return nil
}
