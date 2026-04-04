package client

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/crypto"
	"golang.org/x/term"
)

func (c *Client) HandleList(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("list", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	prefix := fs.String("prefix", "", "The prefix (folder) you'd like to list. Default is '/' (root)")
	fs.StringVar(prefix, "p", "", "alias for --prefix")

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

	// If prefix doesn't start with '/', prepend '/' - handles empty case
	if !strings.HasPrefix(*prefix, "/") {
		*prefix = "/" + *prefix
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

	// Generate salt for prefix hash
	searchSalt := crypto.DeriveSearchSalt(c.MasterKey, c.PersonaID)
	defer clear(searchSalt)

	// Get hashed Prefix
	hashedPrefix := crypto.HashTag(searchSalt, strings.ToLower(*prefix))
	queryValue := url.Values{}
	queryValue.Add("prefix", hashedPrefix)

	// Send the request to list UFOs
	url := c.ActivePersona.BaseURL + api.RouteUFOs + "?" + queryValue.Encode()
	listRes, _, err := ufoSignedRequest[[]api.UFOItem](
		c,
		http.MethodGet,
		url,
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Error getting UFO list: %w", err)
	}
	if len(listRes) == 0 {
		fmt.Println("0 results.")
		return nil
	}
	fmt.Printf("%d UFOs found in %s:\n", len(listRes), *prefix)
	if err = c.printUFOList(listRes); err != nil {
		return err
	}
	return nil
}
