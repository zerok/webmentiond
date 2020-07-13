package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logger zerolog.Logger
var cfg *viper.Viper
var configFilePath string

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
			if configFilePath != "" {
				cfg.SetConfigFile(configFilePath)
				if err := cfg.ReadInConfig(); err != nil {
					logger.Fatal().Err(err).Msg("Failed to read configuration file.")
				}
			}
			if verbose {
				logger = logger.Level(zerolog.DebugLevel)
			}
		},
	}
	rootCmd.PersistentFlags().StringVar(&configFilePath, "config-file", "", "Path to a configuration file")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")
	cfg.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	return newBaseCommand(rootCmd)
}

func init() {
	cfg = viper.New()
}

func main() {
	if err := buildCmd().Execute(); err != nil {
		logger.Fatal().Msg(err.Error())
	}
}
