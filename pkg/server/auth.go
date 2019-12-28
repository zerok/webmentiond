package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/xid"
)

// requireAuthMiddleware ensures that a request contains a valid JWT
// before allowing it to pass through.
func (srv *Server) requireAuthMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		c, err := r.Cookie("token")
		if err != nil {
			srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusUnauthorized})
			return
		}
		if c == nil {
			srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusUnauthorized})
			return
		}
		token, err := jwt.Parse(c.Value, func(token *jwt.Token) (interface{}, error) {
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

		for _, e := range srv.cfg.Auth.AdminEmails {
			if e == claims["sub"] {
				handler.ServeHTTP(w, r)
				return
			}
		}
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusForbidden})
	})
}

func (srv *Server) sendAuthenticationMail(ctx context.Context, email string, token string) error {
	return srv.mailer.SendMail(ctx, "horst@zerokspot.com", []string{email}, "[webmentiond] Login token", token)
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

// handleAuthenticate accepts a login code and generates a token
// attached to the response as cookie.
func (srv *Server) handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authtoken := r.URL.Query().Get("token")
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
	claims := &jwt.StandardClaims{
		Issuer:    "webmentiond",
		Subject:   matchingMail,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 2).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedToken, err := token.SignedString([]byte(srv.cfg.Auth.JWTSecret))
	if err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
	}
	cookie := http.Cookie{
		Name:  "token",
		Value: signedToken,
	}
	http.SetCookie(w, &cookie)
	delete(srv.validToken, matchingMail)
}
