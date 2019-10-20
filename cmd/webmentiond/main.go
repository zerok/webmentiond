package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var logger zerolog.Logger

var rootCmd = &cobra.Command{
	Use: "webmentiond",
	Run: func(cmd *cobra.Command, args []string) {
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
		if ok, _ := cmd.Flags().GetBool("verbose"); ok {
			logger = logger.Level(zerolog.DebugLevel)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().Bool("verbose", false, "Verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal().Msg(err.Error())
	}
}
