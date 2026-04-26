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

	// build command registry
	cmds := client.Commands{
		Registry: make(map[string]func(cmd client.Command) error),
	}
	cmds.Register("health", c.HandleHealth)
	cmds.Register("init", c.HandleInit)
	cmds.Register("new", c.HandleNewPersona)
	cmds.Register("register", c.HandleCreatePersona)
	cmds.Register("list", c.HandleList)
	cmds.Register("search", c.HandleSearch)
	cmds.Register("upload", c.HandleUploadUFO)
	cmds.Register("download", c.HandleDownloadUFO)
	cmds.Register("remove", c.HandleRemoveUFO)
	cmds.Register("update", c.HandleUpdateUFO)
	cmds.Register("details", c.HandleUFODetails)
	cmds.Register("orbit list", c.HandleOrbitList)
	cmds.Register("orbit add", c.HandleOrbitAdd)
	cmds.Register("orbit remove", c.HandleOrbitRemove)
	cmds.Register("orbit details", c.HandleOrbitDetails)
	cmds.Register("orbit update", c.HandleOrbitUpdate)

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
