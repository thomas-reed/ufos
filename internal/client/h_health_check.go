package client

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

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
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = ServerScheme + url
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Error creating request %s", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error executing request %s", err)
	}
	defer res.Body.Close()

	// Possible responses
	switch res.StatusCode {
	case http.StatusOK:
		fmt.Println("Server is up!")
		return nil
	default:
		return fmt.Errorf("Unexpected response status: %d", res.StatusCode)
	}
}
