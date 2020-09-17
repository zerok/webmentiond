package main

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	cfg := viper.New()
	require.Error(t, validateConfig(cfg))

	// Setting the email configuration is not enough
	cfg.Set("email.host", "somehost")
	cfg.Set("email.port", "25")
	cfg.Set("email.from", "somefrom")
	require.Error(t, validateConfig(cfg))

	cfg.Set("server.auth_admin_emails", []string{"test@zerokspot.com"})
	require.Error(t, validateConfig(cfg))
	cfg.Set("server.auth_jwt_secret", "secret")
	require.Error(t, validateConfig(cfg))
	cfg.Set("database.path", "secret")
	require.Error(t, validateConfig(cfg))
	cfg.Set("server.allowed_target_domains", []string{"zerokspot.com"})

	// Now we're complete!
	require.NoError(t, validateConfig(cfg))
}
