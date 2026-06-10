package client

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/thomas-reed/ufos/internal/api"
	"golang.org/x/term"
)

func (c *Client) HandleInit(cmd Command) error {
	// Check if a vault exists and error out immediately
	vaultPath, err := getVaultFilepath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(vaultPath); err == nil {
		return fmt.Errorf("Vault already exists")
	}

	fmt.Println("Welcome to UFOs!")
	fmt.Println("(U)nidentifiable (F)ile/(O)bject (s)torage")
	fmt.Println()
	fmt.Println("Creating new Vault..")

	n, err := getInput("your desired persona name", true)
	if err != nil {
		return err
	}

	fmt.Print("Enter desired vault password: ")
	p, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("Error reading password: %w", err)
	}
	defer clear(p)
	fmt.Println()

	fmt.Print("Confirm vault password: ")
	pc, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("Error reading password confirmation: %w", err)
	}
	defer clear(pc)
	fmt.Println()

	if !bytes.Equal(p, pc) {
		return fmt.Errorf("Password and password confirmation do not match!")
	}

	d, err := getInput("the UFOs domain", true)
	if err != nil {
		return err
	}
	formatDomain(&d)

	if err = CreateNewVault(vaultPath, n, d, p); err != nil {
		return err
	}

	t, err := getInput("registration token", true)
	if err != nil {
		return err
	}

	if err = c.GetPersonaFromVault(n, p); err != nil {
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

	url := serverScheme + d + api.RoutePersonas
	header := map[string]string{
		api.HeaderRegistration: t,
	}
	res, status, err := ufoPublicRequest[api.CreatePersonaResponse](
		c,
		http.MethodPost,
		url,
		body,
		header,
	)
	if err != nil {
		return fmt.Errorf("Error creating persona: %w, (%d)", err, status)
	}

	if res.ID == c.PersonaID {
		fmt.Printf("Persona '%s' has been registered on %s.\n", res.ID, d)
	}
	return nil
}
