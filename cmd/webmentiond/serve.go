package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/zerok/webmentiond/pkg/mailer"
	"github.com/zerok/webmentiond/pkg/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server for sending and receiving mentions",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := logger.WithContext(context.Background())
		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			return err
		}
		mailHost := os.Getenv("MAIL_HOST")
		mailPort := os.Getenv("MAIL_PORT")
		mailUser := os.Getenv("MAIL_USER")
		mailFrom := os.Getenv("MAIL_FROM")
		mailPassword := os.Getenv("MAIL_PASSWORD")
		m := mailer.New(fmt.Sprintf("%s:%s", mailHost, mailPort), smtp.PlainAuth("", mailUser, mailPassword, mailHost), &tls.Config{ServerName: mailHost})

		allowedTargetDomains, err := cmd.Flags().GetStringSlice("allowed-target-domains")
		if err != nil {
			return err
		}
		allowedOrigins, err := cmd.Flags().GetStringSlice("allowed-origin")
		if err != nil {
			return err
		}
		dbpath, err := cmd.Flags().GetString("database")
		if err != nil {
			return err
		}
		if dbpath == "" {
			return fmt.Errorf("no database specified")
		}
		migrationsFolder, err := cmd.Flags().GetString("database-migrations")
		if err != nil {
			return err
		}
		migrationsFolder, err = filepath.Abs(migrationsFolder)
		if err != nil {
			return err
		}
		authAdminEmails, err := cmd.Flags().GetStringSlice("auth-admin-emails")
		if err != nil {
			return err
		}
		authJWTSecret, err := cmd.Flags().GetString("auth-jwt-secret")
		if err != nil {
			return err
		}

		db, err := sql.Open("sqlite3", dbpath)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", dbpath, err)
		}
		defer db.Close()
		srv := server.New(func(c *server.Configuration) {
			c.Auth.JWTSecret = authJWTSecret
			c.Auth.AdminEmails = authAdminEmails
			c.Context = ctx
			c.Database = db
			c.MigrationsFolder = migrationsFolder
			c.Receiver.TargetPolicy = server.RequestPolicyAllowHost(allowedTargetDomains...)
			c.MailFrom = mailFrom
			c.Mailer = m
			c.AllowedOrigins = allowedOrigins
		})
		if err := srv.MigrateDatabase(ctx); err != nil {
			return err
		}
		httpSrv := http.Server{}
		httpSrv.Addr = addr
		httpSrv.Handler = srv
		srv.StartVerifier(ctx)
		logger.Info().Msgf("Listening on %s...", addr)
		return httpSrv.ListenAndServe()
	},
}

func init() {
	serveCmd.Flags().String("database", "./webmentiond.sqlite", "Path to a SQLite database file")
	serveCmd.Flags().String("addr", "127.0.0.1:8080", "Address to listen on for HTTP requests")
	serveCmd.Flags().String("database-migrations", "./pkg/server/migrations", "Path to the database migrations")
	serveCmd.Flags().StringSlice("allowed-target-domains", []string{}, "Domain name that are accepted as targets")
	serveCmd.Flags().StringSlice("allowed-origin", []string{}, "Domain name that is allowed to contact the API (CORS)")

	serveCmd.Flags().String("auth-jwt-secret", "", "Secret used to sign and verify JWTs generated by the server")
	serveCmd.Flags().StringSlice("auth-admin-emails", []string{}, "All e-mail addresses that can gain admin-access")

	rootCmd.AddCommand(serveCmd)
}
