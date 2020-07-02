package main

import "github.com/spf13/cobra"

func newConfigCmd() Command {
	cmd := &cobra.Command{
		Use: "config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	initCmd := &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			if configFilePath == "" {
				cfg.SetConfigFile("webmentiond.yaml")
			}
			return cfg.WriteConfig()
		},
	}

	cmd.AddCommand(initCmd)
	return newBaseCommand(cmd)
}
