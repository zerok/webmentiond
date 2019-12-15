package server

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/zerok/webmentiond/pkg/mailer"
)

type RequestPolicy func(*http.Request) bool

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
}

type Configuration struct {
	Context          context.Context
	Database         *sql.DB
	MigrationsFolder string
	Receiver         ReceiverConfiguration
	Sender           SenderConfiguration
	Auth             AuthConfiguration
	MailFrom         string
	Mailer           mailer.Mailer
}

type Configurator func(c *Configuration)
