package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/mailer"
)

func TestAccessKeyAuthentication(t *testing.T) {
	secret := "12345"
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
	ctx := logger.WithContext(context.Background())
	srv := New(func(c *Configuration) {
		c.Auth.AdminAccessKeys = map[string]string{
			"test-key": "ci",
		}
		c.Auth.JWTTTL = time.Hour * 24 * 7
		c.Auth.JWTSecret = secret
	})

	t.Run("valid-key", func(t *testing.T) {
		w := httptest.NewRecorder()
		params := url.Values{}
		params.Add("key", "test-key")
		r := httptest.NewRequest(http.MethodPost, "/authenticate/access-key", bytes.NewBufferString(params.Encode()))
		r.Header.Set("Content-type", "application/x-www-form-urlencoded")
		srv.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.Equal(t, w.Header().Get("Content-Type"), "application/jwt")
		jot := w.Body.Bytes()
		require.NotEmpty(t, jot)

		// Let's check that the generated token is actually valid for roughly an hour
		token, err := jwt.Parse(string(jot), func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, err)
		require.NotNil(t, token)
		claims, ok := token.Claims.(jwt.MapClaims)
		require.True(t, ok)
		require.Equal(t, "key:ci", claims["sub"])
		require.Equal(t, "webmentiond", claims["iss"])
		exp := time.Unix(int64(claims["exp"].(float64)), 0)
		now := time.Now()
		require.True(t, now.Before(exp), "token is not valid *now*")
		require.True(t, now.Add(time.Minute*70).After(exp), "token is valid beyond one hour")

		// Now try to log in with the given token
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(jot)))
		srv.requireAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})).ServeHTTP(w, r.WithContext(ctx))
		require.Equal(t, w.Code, http.StatusOK)
	})

	t.Run("invalid-key", func(t *testing.T) {
		w := httptest.NewRecorder()
		params := url.Values{}
		params.Add("key", "invalid-key")
		r := httptest.NewRequest(http.MethodPost, "/authenticate/access-key", bytes.NewBufferString(params.Encode()))
		r.Header.Set("Content-type", "application/x-www-form-urlencoded")
		srv.ServeHTTP(w, r)
		require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})
}

func TestAuthentication(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
	ctx := logger.WithContext(context.Background())
	m := mailer.NewDummy()
	srv := New(func(c *Configuration) {
		c.Mailer = m
		c.Auth.AdminEmails = []string{"valid"}
		c.Auth.JWTTTL = time.Hour * 24 * 7
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/request-login", bytes.NewBufferString(""))
	r.Header.Set("Content-type", "application/x-www-form-urlencoded")
	srv.ServeHTTP(w, r)
	require.Equal(t, w.Code, http.StatusBadRequest)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/request-login", bytes.NewBufferString("email=invalid"))
	r.Header.Set("Content-type", "application/x-www-form-urlencoded")
	srv.ServeHTTP(w, r)
	require.Equal(t, w.Code, http.StatusOK)
	require.Len(t, m.Messages, 0)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/request-login", bytes.NewBufferString("email=valid"))
	r.Header.Set("Content-type", "application/x-www-form-urlencoded")
	srv.ServeHTTP(w, r)
	require.Equal(t, w.Code, http.StatusOK)
	require.Len(t, m.Messages, 1)
	// The email should contain the link to log in and the authentication token
	// on a seperate line:
	token := srv.validToken["valid"]
	require.Equal(t, fmt.Sprintf("/ui/#/authenticate/%s\n\n%s", token, token), m.Messages[0].Body)

	require.NotEmpty(t, token)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/authenticate?token="+token, nil)
	srv.ServeHTTP(w, r)
	require.Equal(t, w.Code, http.StatusOK)
	require.Equal(t, w.Header().Get("Content-Type"), "application/jwt")
	jot := w.Body.Bytes()
	require.NotEmpty(t, jot)

	// If we run against the auth-middleware without a token, we
	// should get a 401 status:
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/", nil)
	srv.requireAuthMiddleware(nil).ServeHTTP(w, r)
	require.Equal(t, w.Code, http.StatusUnauthorized)

	// If we add a broken jwt, the middleware should fail with a
	// 401 status as if no token were present:
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Authorization", "Bearer BROKEN")
	srv.requireAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})).ServeHTTP(w, r.WithContext(ctx))
	require.Equal(t, w.Code, http.StatusUnauthorized)

	// If we add the correct token, the middleware should let us
	// pass:
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(jot)))
	srv.requireAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})).ServeHTTP(w, r.WithContext(ctx))
	require.Equal(t, w.Code, http.StatusOK)

	// Let's make sure that the bearer token has the expected expiration time:
	jwt.Parse(string(jot), func(token *jwt.Token) (interface{}, error) {
		claims, _ := token.Claims.(jwt.MapClaims)
		rawExp := claims["exp"]
		if exp, ok := rawExp.(float64); ok {
			expTime := time.Unix(int64(exp), 0)
			fmt.Println(expTime)
			require.InDelta(t, expTime.Unix(), time.Now().Add(time.Hour*24*7).Unix(), 5)
		}
		return []byte{}, nil
	})
}
