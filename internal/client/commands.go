package client

import "fmt"

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Registry map[string]func(cmd Command) error
}

func (c *Commands) Register(name string, f func(Command) error) {
	c.Registry[name] = f
}

func (c *Commands) Run(cmd Command) error {
	cmdFunc, found := c.Registry[cmd.Name]
	if !found {
		return fmt.Errorf("%s not found in list of commands", cmd.Name)
	}
	return cmdFunc(cmd)
}