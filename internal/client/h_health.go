package client

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
)

func (c *Client) HandleHealth(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("health", flag.ContinueOnError)

	domain := fs.String("domain", "", "The UFO domain to check")
	fs.StringVar(domain, "d", "", "alias for --domain")
	if err := fs.Parse(cmd.Args); err != nil {
		return err
	}

	// If domain wasn't in Args, prompt
	if *domain == "" {
		d, err := getInput("domain", true)
		if err != nil {
			return err
		}
		domain = &d
	}

	url := serverScheme + *domain + api.RouteHealth
	_, status, err := ufoPublicRequest[api.EmptyResponse](c, http.MethodGet, url, nil, nil)
	if err != nil {
		return err
	}

	switch status {
	case http.StatusOK:
		fmt.Printf("%s is up.\n", *domain)
	default:
		fmt.Printf("Unexpected status code (%d)\n", status)
	}
	return nil
}
