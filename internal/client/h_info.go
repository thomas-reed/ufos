package client

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/thomas-reed/ufos/internal/api"
	"golang.org/x/term"
)

func (c *Client) HandleUFOInfo(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("info", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	id := fs.String("id", "", "The id of the file you want the detailed info for")
	fs.StringVar(id, "i", "", "alias for --id")

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

	// If id wasn't in Args, error out
	if *id == "" {
		return fmt.Errorf("Enter id of UFO you wish to retrieve using '--id' or '-i'")
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

	// Get UFO Metadata
	ufoUrl := c.ActivePersona.BaseURL + api.RouteUFOs + "/" + *id
	ufoRes, _, err := ufoSignedRequest[api.UFOMetadataFromHeader](c, http.MethodHead, ufoUrl, nil, nil)

	if err = c.printUFODetails(ufoRes); err != nil {
		return err
	}
	return nil
}
