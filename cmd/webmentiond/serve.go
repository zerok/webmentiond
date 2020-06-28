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
	"regexp"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/zerok/webmentiond/pkg/mailer"
	"github.com/zerok/webmentiond/pkg/policies"
	"github.com/zerok/webmentiond/pkg/server"
)

type dbPolicyLoader struct {
	db *sql.DB
}

func (l *dbPolicyLoader) Load(ctx context.Context) ([]policies.URLPolicy, error) {
	result, err := l.db.QueryContext(ctx, "SELECT id, url_pattern, policy, weight FROM url_policies ORDER BY weight ASC")
	if err != nil {
		return nil, err
	}
	defer result.Close()
	pols := make([]policies.URLPolicy, 0, 10)
	for result.Next() {
		var id int
		var pat string
		var pol string
		var weight int
		if err := result.Scan(&id, &pat, &pol, &weight); err != nil {
			return nil, err
		}
		urlp, err := regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		pols = append(pols, policies.URLPolicy{
			ID:         id,
			URLPattern: urlp,
			Policy:     policies.Policy(pol),
			Weight:     weight,
		})
	}
	return pols, nil
}

func newServeCmd() Command {
	var tokenTTL time.Duration
	var verificationTimeoutDur time.Duration
	var notify bool
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
			uiPath, err := cmd.Flags().GetString("ui-path")
			if err != nil {
				return err
			}
			uiPath, err = filepath.Abs(uiPath)
			if err != nil {
				return err
			}

			db, err := sql.Open("sqlite3", dbpath)
			if err != nil {
				return fmt.Errorf("failed to open %s: %w", dbpath, err)
			}
			pol := policies.NewRegistry(policies.DEFAULT)
			policyLoader := &dbPolicyLoader{db: db}
			defer db.Close()
			srv := server.New(func(c *server.Configuration) {
				c.Auth.JWTSecret = authJWTSecret
				c.Auth.AdminEmails = authAdminEmails
				c.Auth.JWTTTL = tokenTTL
				c.Context = ctx
				c.Database = db
				c.MigrationsFolder = migrationsFolder
				c.Receiver.TargetPolicy = server.RequestPolicyAllowHost(allowedTargetDomains...)
				c.MailFrom = mailFrom
				c.Mailer = m
				c.AllowedOrigins = allowedOrigins
				c.PublicURL, _ = cmd.Flags().GetString("public-url")
				c.UIPath = uiPath
				c.NotifyOnVerification = notify
				c.Policies = pol
				c.PolicyLoader = policyLoader
			})
			if err := srv.MigrateDatabase(ctx); err != nil {
				return err
			}
			if err := pol.Load(ctx, policyLoader); err != nil {
				return err
			}
			go func() {
				ticker := time.NewTicker(time.Second * 20)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						break
					}
					if err := pol.Load(ctx, policyLoader); err != nil {
						logger.Error().Err(err).Msg("Failed to load policies")
					}
				}
			}()
			httpSrv := http.Server{}
			httpSrv.Addr = addr
			httpSrv.Handler = srv
			srv.StartVerifier(ctx)
			if err := srv.UpdateGlobalMetrics(ctx); err != nil {
				return err
			}
			logger.Info().Msgf("Listening on %s...", addr)
			return httpSrv.ListenAndServe()
		},
	}

	serveCmd.Flags().String("database", "./webmentiond.sqlite", "Path to a SQLite database file")
	serveCmd.Flags().String("addr", "127.0.0.1:8080", "Address to listen on for HTTP requests")
	serveCmd.Flags().String("public-url", "http://127.0.0.1:8080", "URL used as base for generating links")
	serveCmd.Flags().String("database-migrations", "./pkg/server/migrations", "Path to the database migrations")
	serveCmd.Flags().StringSlice("allowed-target-domains", []string{}, "Domain name that are accepted as targets")
	serveCmd.Flags().StringSlice("allowed-origin", []string{}, "Domain name that is allowed to contact the API (CORS)")
	serveCmd.Flags().String("ui-path", "./frontend", "Path which should be served as /ui/")

	serveCmd.Flags().String("auth-jwt-secret", "", "Secret used to sign and verify JWTs generated by the server")
	serveCmd.Flags().DurationVar(&tokenTTL, "auth-jwt-ttl", time.Hour*24*7, "TTL of the generated JWTs")
	serveCmd.Flags().StringSlice("auth-admin-emails", []string{}, "All e-mail addresses that can gain admin-access")
	serveCmd.Flags().DurationVar(&verificationTimeoutDur, "verification-timeout", time.Second*30, "Wait at least this time before re-verifying a source")

	serveCmd.Flags().BoolVar(&notify, "send-notifications", false, "Send email notifications about new/updated webmentions")

	return newBaseCommand(serveCmd)
}
