package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/xid"
)

type authorizedContextKey struct{}

func isAuthorized(ctx context.Context) bool {
	return ctx.Value(authorizedContextKey{}) == true
}

// AuthorizeContext marks the current context as being authorized.
func AuthorizeContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, authorizedContextKey{}, true)
}

// requireAuthMiddleware ensures that a request contains a valid JWT
// before allowing it to pass through.
func (srv *Server) requireAuthMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if isAuthorized(ctx) {
			handler.ServeHTTP(w, r)
			return
		}
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusUnauthorized})
			return
		}
		token, err := jwt.Parse(strings.TrimPrefix(header, "Bearer "), func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid token-signing format found")
			}
			return []byte(srv.cfg.Auth.JWTSecret), nil
		})
		if err != nil {
			srv.sendError(ctx, w, &HTTPError{Err: err, StatusCode: http.StatusUnauthorized})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			srv.sendError(ctx, w, &HTTPError{Err: fmt.Errorf("unexpected token claims found"), StatusCode: http.StatusBadRequest})
			return
		}

		for _, k := range srv.cfg.Auth.AdminAccessKeys {
			if claims["sub"] == formatAccessKeySubject(k) {
				handler.ServeHTTP(w, r.WithContext(AuthorizeContext(ctx)))
				return
			}
		}

		for _, e := range srv.cfg.Auth.AdminEmails {
			if e == claims["sub"] {
				handler.ServeHTTP(w, r.WithContext(AuthorizeContext(ctx)))
				return
			}
		}
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusForbidden})
	})
}

func (srv *Server) sendAuthenticationMail(ctx context.Context, email string, token string) error {
	tokenURL := fmt.Sprintf("%s/ui/#/authenticate/%s", srv.cfg.PublicURL, token)
	message := fmt.Sprintf("%s\n\n%s", tokenURL, token)
	return srv.mailer.SendMail(ctx, srv.cfg.MailFrom, []string{email}, "[webmentiond] Login token", message)
}

// handleLogin just takes a user's email address and sends out an
// authentication link.
func (srv *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: err})
		return
	}
	email := r.FormValue("email")
	if email == "" {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("no email provided")})
		return
	}
	for _, e := range srv.cfg.Auth.AdminEmails {
		if e == email {
			token := xid.New().String()
			srv.validTokenMutex.Lock()
			srv.validToken[email] = token
			srv.validTokenMutex.Unlock()
			if err := srv.sendAuthenticationMail(ctx, email, token); err != nil {
				srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
				return
			}
		}
	}
}

func (srv *Server) handleAuthenticateWithAccessKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: err})
		return
	}
	key := r.FormValue("key")
	if key == "" {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("token key")})
		return
	}

	name, ok := srv.cfg.Auth.AdminAccessKeys[key]
	if !ok {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusUnauthorized, Err: fmt.Errorf("invalid key")})
		return
	}
	if err := srv.generateJWT(ctx, w, fmt.Sprintf("key:%s", name)); err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
}

func (srv *Server) generateJWT(ctx context.Context, w http.ResponseWriter, subject string) error {
	exp := time.Now().Add(srv.cfg.Auth.JWTTTL).Unix()
	claims := &jwt.StandardClaims{
		Issuer:    "webmentiond",
		Subject:   subject,
		ExpiresAt: exp,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedToken, err := token.SignedString([]byte(srv.cfg.Auth.JWTSecret))
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/jwt")
	w.Write([]byte(signedToken))
	return nil
}

// handleAuthenticate accepts a login code and generates a token
// attached to the response as cookie.
func (srv *Server) handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: err})
		return
	}
	authtoken := r.FormValue("token")
	if authtoken == "" {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("token missing")})
		return
	}
	var matchingMail string
	srv.validTokenMutex.Lock()
	defer srv.validTokenMutex.Unlock()
	for mail, storedToken := range srv.validToken {
		if storedToken == authtoken {
			matchingMail = mail
			break
		}
	}

	if matchingMail == "" {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("token invalid")})
		return
	}
	if err := srv.generateJWT(ctx, w, matchingMail); err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
	delete(srv.validToken, matchingMail)
}

func formatAccessKeySubject(keyName string) string {
	return fmt.Sprintf("key:%s", keyName)
}
