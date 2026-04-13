package client

import (
	"flag"
	"fmt"
	"os"

	"github.com/thomas-reed/ufos/internal/api"
	"golang.org/x/term"
)

func (c *Client) HandleOrbitInfo(cmd Command) error {
  // Set up flags and parse
	fs := flag.NewFlagSet("orbit details", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	id := fs.String("id", "", "The Persona ID of the user you want to inspect")
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
		return fmt.Errorf("Enter Persona ID of the user you wish to inspect with '--id' or '-i'")
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
    
  // Get satellite data and print
	url := c.ActivePersona.BaseURL + api.RouteOrbit + "/" + *id
	sat, _, err := ufoSignedRequest[api.Satellite](c, "GET", url, nil, nil)
    
  if err = c.printSatelliteDetails(sat); err != nil {
		return err
	}
	return nil
}