package client

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"golang.org/x/term"
)

func Init(cmd Command) error {
	fmt.Println("Welcome to UFOs!")
	fmt.Println("(U)nidentifiable (F)ile/(O)bject (s)torage")
	fmt.Println()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Creating new Vault..")
	fmt.Print("Enter persona name > ")
	if !scanner.Scan() {
		return fmt.Errorf("Input interrupted!")
	}
	n := scanner.Text()

	fmt.Print("Enter UFOs URL > ")
	if !scanner.Scan() {
		return fmt.Errorf("Input interrupted!")
	}
	url := scanner.Text()

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
