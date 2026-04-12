package client

import (
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

func (c *Client) HandleSearch(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("search", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	tagList := fs.String("tags", "", "The tags you wish to search for. Surround multiple tags with quotes, separated by commas")
	fs.StringVar(tagList, "t", "", "alias for --tags")
	prefix := fs.String("prefix", "", "The prefix (folder) you'd like to list. Default is '/' (root)")
	fs.StringVar(prefix, "p", "", "alias for --prefix")

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

	// If tags is empty, error out
	if *tagList == "" {
		return fmt.Errorf("Include search terms using '--tags' or '-t'")
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

	// Get list of hashed Tags
	queryValues := url.Values{}
	// if prefix exists, format it and add to query params
	if *prefix != "" {
		formatPrefix(prefix)
		hashedPrefix := crypto.HashTag(searchSalt, *prefix)
		queryValues.Add("prefix", hashedPrefix)
	}
	tags := strings.Split(*tagList, ",")
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		hashedTag := crypto.HashTag(searchSalt, strings.ToLower(tag))
		queryValues.Add("tag", hashedTag)
	}

	// Send the request to search for UFOs
	url := c.ActivePersona.BaseURL + api.RouteUFOs + "?" + queryValues.Encode()
	listRes, _, err := ufoSignedRequest[[]api.UFO](
		c,
		http.MethodGet,
		url,
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Error getting UFOs: %w", err)
	}
	if len(listRes) == 0 {
		fmt.Println("0 results.")
		return nil
	}
	fmt.Printf("%d UFOs found:\n", len(listRes))
	if err = c.printUFOList(listRes); err != nil {
		return err
	}
	return nil
}
