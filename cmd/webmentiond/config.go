package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

func validateConfig(cfg *viper.Viper) error {
	t := &failState{}
	requireConfigSet(t, cfg, "email.host", "EMAIL_HOST", "")
	requireConfigSet(t, cfg, "email.port", "EMAIL_PORT", "")
	requireConfigSet(t, cfg, "email.from", "EMAIL_FROM", "")
	requireConfigStringSliceSet(t, cfg, "server.auth_admin_emails", "", "--auth-admin-emails")
	requireConfigStringSliceSet(t, cfg, "server.allowed_target_domains", "", "--allowed-target-domains")
	requireConfigSet(t, cfg, "server.auth_jwt_secret", "", "--auth-jwt-secret")
	requireConfigSet(t, cfg, "database.path", "", "--database")
	return t.Error()
}

func requireConfigSet(fs *failState, cfg *viper.Viper, field string, env string, flag string) {
	if fs.Failed() {
		return
	}
	val := cfg.GetString(field)
	if val == "" {
		fs.FailFor(field, env, flag)
	}
}
func requireConfigStringSliceSet(fs *failState, cfg *viper.Viper, field string, env string, flag string) {
	if fs.Failed() {
		return
	}
	val := cfg.GetStringSlice(field)
	if val == nil || len(val) == 0 {
		fs.FailFor(field, env, flag)
	}
}

type failState struct {
	err error
}

func (fs *failState) Failed() bool {
	return fs.err != nil
}
func (fs *failState) Error() error {
	return fs.err
}
func (fs *failState) Fail(err error) {
	fs.err = err
}
func (fs *failState) FailFor(field string, env string, flag string) {
	msg := fmt.Sprintf("%s not set. Please set this using the configuration file", field)
	if env != "" {
		msg += fmt.Sprintf(" or the %s environment variable", env)
	}
	if flag != "" {
		msg += fmt.Sprintf(" or the %s flag", flag)
	}
	msg += "."
	fs.Fail(fmt.Errorf(msg))
}
