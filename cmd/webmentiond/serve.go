package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"net/smtp"
	"path/filepath"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	var verificationMaxRedirects int
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
			mailHost := cfg.GetString("email.host")
			mailPort := cfg.GetString("email.port")
			mailUser := cfg.GetString("email.user")
			mailPassword := cfg.GetString("email.password")
			mailFrom := cfg.GetString("email.from")
			mailNoTLS := cfg.GetBool("email.no_tls")
			mailUseStartTLS := cfg.GetBool("email.use_starttls")
			var mailAuth smtp.Auth
			if mailUser != "" {
				mailAuth = smtp.PlainAuth("", mailUser, mailPassword, mailHost)
			}
			mailConfigs := []mailer.DefaultMailerConfigurator{}
			mailConfigs = append(mailConfigs, mailer.WithTLS(!mailNoTLS))
			mailConfigs = append(mailConfigs, mailer.WithStartTLS(mailUseStartTLS))
			mailConfigs = append(mailConfigs, mailer.WithTLSConfig(&tls.Config{ServerName: mailHost}))

			m := mailer.New(fmt.Sprintf("%s:%s", mailHost, mailPort), mailAuth, mailConfigs...)

			allowedTargetDomains := cfg.GetStringSlice("server.allowed_target_domains")
			allowedOrigins := cfg.GetStringSlice("server.allowed_origin")
			dbpath := cfg.GetString("database.path")
			if dbpath == "" {
				return fmt.Errorf("no database specified")
			}
			migrationsFolder := cfg.GetString("database.migrations")
			migrationsFolder, err = filepath.Abs(migrationsFolder)
			if err != nil {
				return err
			}
			authAdminEmails := cfg.GetStringSlice("server.auth_admin_emails")
			authJWTSecret := cfg.GetString("server.auth_jwt_secret")
			uiPath := cfg.GetString("server.ui_path")
			uiPath, err = filepath.Abs(uiPath)
			if err != nil {
				return err
			}

			if err := validateConfig(cfg); err != nil {
				return fmt.Errorf("configuration invalid: %w", err)
			}

			db, err := sql.Open("sqlite3", dbpath)
			if err != nil {
				return fmt.Errorf("failed to open %s: %w", dbpath, err)
			}
			pol := policies.NewRegistry(policies.DEFAULT)
			policyLoader := &dbPolicyLoader{db: db}
			defer db.Close()
			metricsAddr := cfg.GetString("server.metrics_addr")
			exposeMetrics := metricsAddr == addr
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
				c.PublicURL = cfg.GetString("server.public_url")
				c.UIPath = uiPath
				c.NotifyOnVerification = notify
				c.Policies = pol
				c.PolicyLoader = policyLoader
				c.VerificationMaxRedirects = verificationMaxRedirects
				c.ExposeMetrics = exposeMetrics
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
			// Start the metrics exporter on a separate listener:
			if metricsAddr != "" && !exposeMetrics {
				logger.Info().Msgf("Exposing metrics  via %s...", metricsAddr)
				startMetricsExporter(ctx, metricsAddr)
			}
			if metricsAddr == "" {
				logger.Info().Msgf("To export metrics set --metrics-addr")
			}
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

	cfg.BindEnv("email.host", "MAIL_HOST")
	cfg.BindEnv("email.port", "MAIL_PORT")
	cfg.BindEnv("email.user", "MAIL_USER")
	cfg.BindEnv("email.password", "MAIL_PASSWORD")
	cfg.BindEnv("email.from", "MAIL_FROM")
	cfg.BindEnv("email.no_tls", "MAIL_NO_TLS")
	cfg.BindEnv("email.use_starttls", "MAIL_USE_STARTTLS")
	cfg.BindEnv("server.auth_jwt_secret", "SERVER_AUTH_JWT_SECRET")

	serveCmd.Flags().String("database", "./webmentiond.sqlite", "Path to a SQLite database file")
	cfg.BindPFlag("database.path", serveCmd.Flags().Lookup("database"))
	serveCmd.Flags().String("database-migrations", "./pkg/server/migrations", "Path to the database migrations")
	cfg.BindPFlag("database.migrations", serveCmd.Flags().Lookup("database-migrations"))

	serveCmd.Flags().String("addr", "127.0.0.1:8080", "Address to listen on for HTTP requests")
	serveCmd.Flags().String("metrics-addr", "", "Address where metrics are exposed")
	cfg.BindPFlag("server.addr", serveCmd.Flags().Lookup("addr"))
	cfg.BindPFlag("server.metrics_addr", serveCmd.Flags().Lookup("metrics-addr"))

	serveCmd.Flags().String("public-url", "http://127.0.0.1:8080", "URL used as base for generating links")
	cfg.BindPFlag("server.public_url", serveCmd.Flags().Lookup("public-url"))

	serveCmd.Flags().StringSlice("allowed-target-domains", []string{}, "Domain name that are accepted as targets")
	cfg.BindPFlag("server.allowed_target_domains", serveCmd.Flags().Lookup("allowed-target-domains"))
	serveCmd.Flags().StringSlice("allowed-origin", []string{}, "Domain name that is allowed to contact the API (CORS)")
	cfg.BindPFlag("server.allowed_origin", serveCmd.Flags().Lookup("allowed-origin"))
	serveCmd.Flags().String("ui-path", "./frontend", "Path which should be served as /ui/")
	cfg.BindPFlag("server.ui_path", serveCmd.Flags().Lookup("ui-path"))

	serveCmd.Flags().String("auth-jwt-secret", "", "Secret used to sign and verify JWTs generated by the server")
	cfg.BindPFlag("server.auth_jwt_secret", serveCmd.Flags().Lookup("auth-jwt-secret"))
	serveCmd.Flags().DurationVar(&tokenTTL, "auth-jwt-ttl", time.Hour*24*7, "TTL of the generated JWTs")
	cfg.BindPFlag("server.auth_jwt_ttl", serveCmd.Flags().Lookup("auth-jwt-ttl"))
	serveCmd.Flags().StringSlice("auth-admin-emails", []string{}, "All e-mail addresses that can gain admin-access")
	cfg.BindPFlag("server.auth_admin_emails", serveCmd.Flags().Lookup("auth-admin-emails"))
	serveCmd.Flags().DurationVar(&verificationTimeoutDur, "verification-timeout", time.Second*30, "Wait at least this time before re-verifying a source")
	serveCmd.Flags().IntVar(&verificationMaxRedirects, "verification-max-redirects", 10, "Number of redirects allowed during verification")
	cfg.BindPFlag("verification.timeout", serveCmd.Flags().Lookup("verification-timeout"))

	serveCmd.Flags().BoolVar(&notify, "send-notifications", false, "Send email notifications about new/updated webmentions")
	cfg.BindPFlag("notifications.enabled", serveCmd.Flags().Lookup("send-notifications"))

	return newBaseCommand(serveCmd)
}

func startMetricsExporter(ctx context.Context, addr string) {
	srv := http.Server{}
	srv.Addr = addr
	r := chi.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
	srv.Handler = r
	go srv.ListenAndServe()
}
