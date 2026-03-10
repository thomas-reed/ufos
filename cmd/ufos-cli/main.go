package cmd

import (
	"log"

	"github.com/thomas-reed/ufos/internal/client"
)

func main() {
	// register cmds like in gator
	// read cmd line parameters, get cmd, parameters
	// get password using term.ReadPassword()
	personaName := ""
	password := []byte{}
	c, err := client.NewClient(password, personaName)
	if err != nil {
		log.Fatalf("Error initializing client: %v", err)
	}
	// handle cmd, sign request, send to server

	// as soon as necessary wipe sensitive data.  should i just clear entire client?
	clear(c.PersonaData.PrivateKey)
	clear(c.MasterKey)
}
