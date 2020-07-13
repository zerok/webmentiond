package server

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/zerok/webmentiond/pkg/mailer"
	"github.com/zerok/webmentiond/pkg/policies"
)

// RequestPolicy functions allow you to mark incoming requests as allowed or
// denied.
type RequestPolicy func(*http.Request) bool

// RequestPolicyAllowHost creates a policy that allows only requests targeted at
// specific hosts.
func RequestPolicyAllowHost(hosts ...string) RequestPolicy {
	return func(r *http.Request) bool {
		if r.URL == nil {
			return false
		}
		rh := r.URL.Host
		switch r.URL.Scheme {
		case "http":
			rh = strings.TrimSuffix(rh, ":80")
		case "https":
			rh = strings.TrimSuffix(rh, ":443")
		}
		for _, host := range hosts {
			if rh == host {
				return true
			}
		}
		return false
	}
}

type ReceiverConfiguration struct {
	TargetPolicy RequestPolicy
}

type SenderConfiguration struct{}

type AuthConfiguration struct {
	AdminEmails []string
	JWTSecret   string
	JWTTTL      time.Duration
}

type Configuration struct {
	Context                     context.Context
	Database                    *sql.DB
	MigrationsFolder            string
	Receiver                    ReceiverConfiguration
	Sender                      SenderConfiguration
	Auth                        AuthConfiguration
	MailFrom                    string
	Mailer                      mailer.Mailer
	AllowedOrigins              []string
	PublicURL                   string
	UIPath                      string
	VerificationTimeoutDuration time.Duration
	VerificationMaxRedirects    int
	NotifyOnVerification        bool
	Policies                    *policies.Registry
	PolicyLoader                policies.Loader
}

type Configurator func(c *Configuration)
