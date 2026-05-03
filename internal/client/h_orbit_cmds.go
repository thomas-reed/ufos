package client

import (
	"fmt"
	"strings"
)

func (c *Client) HandleOrbitCmds(cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("orbit command requires a subcommand - options: add, update, list, details, remove)")
	}
	subcommand := strings.ToLower(cmd.Args[0])
	cmd.Name = cmd.Name + " " + subcommand
	cmd.Args = cmd.Args[1:]
	switch subcommand {
	case "add":
		return c.HandleOrbitAdd(cmd)
	case "update":
		return c.HandleOrbitUpdate(cmd)
	case "list":
		return c.HandleOrbitList(cmd)
	case "details":
		return c.HandleOrbitDetails(cmd)
	case "remove":
		return c.HandleOrbitRemove(cmd)
	default:
		return fmt.Errorf("Orbit subcommand '%s' not recognized", subcommand)
	}
}
