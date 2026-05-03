package client

import "fmt"

type Command struct {
	Name string
	Args []string
}

type CommandInfo struct {
	Handler func(cmd Command) error
	Help    string
}

type Commands struct {
	Registry map[string]CommandInfo
}

func (c *Commands) Register(name string, handler func(Command) error, help string) {
	c.Registry[name] = CommandInfo{
		Handler: handler,
		Help:    help,
	}
}

func (c *Commands) Run(cmd Command) error {
	cmdInfo, found := c.Registry[cmd.Name]
	if !found {
		return fmt.Errorf("%s not found in list of commands", cmd.Name)
	}
	return cmdInfo.Handler(cmd)
}
