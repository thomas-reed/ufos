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
	Registry: make(map[string]client.CommandInfo),
}

cmds.Register("help", cmds.PrintHelp, "List available commands")
cmds.Register("health", c.HandleHealth, "Check server health at the given --domain (-d)")
cmds.Register("init", c.HandleInit, "Create a new vault and register your first persona")
cmds.Register("new", c.HandleNewPersona, "Get a new registration token from the server to create a new persona")
cmds.Register("register", c.HandleCreatePersona, "Register a persona with the server")
cmds.Register("list", c.HandleList, "List UFOs with the given --prefix (-p)")
cmds.Register("search", c.HandleSearch, "Search for UFOs by --tag (-t), and limit to a given --prefix (-p)")
cmds.Register("upload", c.HandleUploadUFO, "Upload a new UFO")
cmds.Register("download", c.HandleDownloadUFO, "Download one of your own UFOs")
cmds.Register("fetch", c.HandleFetchUFO, "Fetch a shared UFO from another host")
cmds.Register("remove", c.HandleRemoveUFO, "Remove a UFO")
cmds.Register("update", c.HandleUpdateUFO, "Update a UFO")
cmds.Register("details", c.HandleUFODetails, "Show UFO metadata")
cmds.Register("orbit", c.HandleOrbitCmds, "Manage your orbit. Subcommands: add, update, list, details, remove")

	// parse cmd line arguments
	if len(os.Args) < 2 {
		log.Fatalln("Command error: Too few arguments.  Usage: ufos <command> [args...]")
	}
	cmdName := strings.ToLower(os.Args[1])
	cmdArgs := os.Args[2:]

	// run given command
	if err = cmds.Run(client.Command{Name: cmdName, Args: cmdArgs}); err != nil {
		log.Fatalf("System malfunction: Error running %s\n%s\n", cmdName, err)
	}

	os.Exit(0)
}
