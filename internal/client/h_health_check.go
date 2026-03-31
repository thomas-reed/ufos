package client

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/thomas-reed/ufos/internal/api"
)

func (c *Client) HandleHealthCheck(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("health", flag.ContinueOnError)

	domain := fs.String("domain", "", "The UFO domain to check")
	fs.StringVar(domain, "d", "", "alias for --domain")
	if err := fs.Parse(cmd.Args); err != nil {
		return err
	}

	// If domain wasn't in Args, prompt
	if *domain == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter domain > ")
		if !scanner.Scan() {
			return fmt.Errorf("Input interrupted!")
		}
		d := scanner.Text()
		domain = &d
	}

	url := *domain + api.RouteHealth
	_, status, err := ufoPublicRequest[struct{}](c, http.MethodGet, url, nil, nil)
	if err != nil {
		return err
	}

	switch status {
	case http.StatusOK:
		fmt.Println("Server is up!")
	default:
		fmt.Printf("Unexpected status code (%s)\n", status)
	}
	return nil
}
