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
	fs := flag.NewFlagSet("download", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	id := fs.String("id", "", "The personaID and domain (<id>@<domain>) of the user you want to add to your orbit")
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
		return fmt.Errorf("Enter <personaID>@<domain> of user you wish to add to your orbit using '--id' or '-i'")
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

	// 2. Dissection: Split bob@domain.com

	// 3. Discovery Handshake (Foreign Server)
	url := domain + api.RouteRegister + "/" + userID // Or api.RoutePersonas
	keys, _, err := ufoPublicRequest[api.PersonaKeysResponse](c, "GET", url, nil, nil)
	if err != nil {
		return err
	}

	// 4. Persistence Ritual (Local Server)
	orbitUrl := c.ActivePersona.BaseURL + api.RouteOrbit
	req := api.SatelliteRequest{
		PersonaID:   keys.PersonaID,
		SigningKey:  keys.SigningKey,
		ExchangeKey: keys.ExchangeKey,
		// Optional: Metadata encrypted with MasterKey
	}

	_, _, err = ufoSignedRequest[api.Satellite](c, "POST", orbitUrl, req, nil)
	// ... handle ...
}