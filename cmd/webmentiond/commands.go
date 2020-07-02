package main

import "github.com/spf13/cobra"

type Command interface {
	Execute() error
	Cmd() *cobra.Command
	AddCommand(name string, cmd Command)
}

type BaseCommand struct {
	*cobra.Command
	commands map[string]Command
}

func newBaseCommand(cmd *cobra.Command) *BaseCommand {
	return &BaseCommand{
		Command:  cmd,
		commands: make(map[string]Command),
	}
}

func (c *BaseCommand) Execute() error {
	return c.Command.Execute()
}

func (c *BaseCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *BaseCommand) AddCommand(name string, cmd Command) {
	c.commands[name] = cmd
	c.Command.AddCommand(cmd.Cmd())
}

func buildCmd() Command {
	root := newRootCmd()
	serveCmd := newServeCmd()
	sendCmd := newSendCmd()
	verifyCmd := newVerifyCmd()
	root.AddCommand("serve", serveCmd)
	root.AddCommand("send", sendCmd)
	root.AddCommand("verify", verifyCmd)
	root.AddCommand("config", newConfigCmd())
	return root
}
