package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var logger zerolog.Logger

func newRootCmd() Command {
	var verbose bool
	var rootCmd = &cobra.Command{
		Use:           "webmentiond",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(zerolog.InfoLevel)
			if verbose {
				logger = logger.Level(zerolog.DebugLevel)
			}
		},
	}
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")
	return newBaseCommand(rootCmd)
}

func main() {
	if err := buildCmd().Execute(); err != nil {
		logger.Fatal().Msg(err.Error())
	}
}
