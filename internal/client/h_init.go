package client

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/term"
)

func HandleInit(cmd Command) error {
	fmt.Println("Welcome to UFOs!")
	fmt.Println("(U)nidentifiable (F)ile/(O)bject (s)torage")
	fmt.Println()
	fmt.Println("Creating new Vault..")

	n, err := getInput("your desired persona name", true)
	if err != nil {
		return err
	}

	url, err := getInput("the UFOs server URL", true)
	if err != nil {
		return err
	}

	fmt.Printf("Enter master password: ")
	p, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("Error reading password: %w", err)
	}
	defer clear(p)

	fmt.Printf("Confirm master password: ")
	pc, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("Error reading password confirmation: %w", err)
	}
	defer clear(pc)

	if !bytes.Equal(p, pc) {
		return fmt.Errorf("Password and password confirmation do not match!")
	}

	return CreateNewVault(n, url, p)
}
