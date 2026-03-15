package main

import (
	"log"
	"os"
	"strings"

	"github.com/thomas-reed/ufos/internal/client"
)

func main() {
	c, err := client.NewClient()
	if err != nil {
		log.Fatalf("Error initializing client: %v", err)
	}
	defer clear(c.ActivePersona.PrivateKey)
	defer clear(c.MasterKey)
	// build command registry
	cmds := client.Commands{
		Registry: make(map[string]func(cmd client.Command) error),
	}
	cmds.Register("Init", client.Init)

	// parse cmd line arguments
	if len(os.Args) < 2 {
		log.Fatalln("Too few arguments.  Usage: ufos <command> [args...]")
	}
	cmdName := strings.ToLower(os.Args[1])
	cmdArgs := os.Args[2:]

	// run given command
	if err = cmds.Run(client.Command{Name: cmdName, Args: cmdArgs}); err != nil {
		log.Fatalf("Error running %s command: %s\n", cmdName, err)
	}

	os.Exit(0)
}
