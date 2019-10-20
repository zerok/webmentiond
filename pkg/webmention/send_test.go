package webmention_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/webmention"
)

func TestSendMention(t *testing.T) {
	t.Run("validate-request", func(t *testing.T) {
		ctx := context.Background()
		var req *http.Request
		var reqBody []byte
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req = r
			reqBody, _ = ioutil.ReadAll(r.Body)
			r.Body.Close()
			w.WriteHeader(201)
		}))
		sender := webmention.NewSender(func(c *webmention.SenderConfiguration) {
			c.HTTPClient = srv.Client()
		})
		err := sender.Send(ctx, srv.URL, webmention.Mention{
			Source: "https://source.com",
			Target: "https://target.com",
		})
		defer req.Body.Close()
		require.NoError(t, err)
		require.Equal(t, http.MethodPost, req.Method)
		require.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
		data, err := url.ParseQuery(string(reqBody))
		require.NoError(t, err)
		require.NotEmpty(t, data.Get("source"))
		require.NotEmpty(t, data.Get("target"))
	})

	t.Run("error-on-non-success-code", func(t *testing.T) {
		ctx := context.Background()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(400)
		}))
		sender := webmention.NewSender(func(c *webmention.SenderConfiguration) {
			c.HTTPClient = srv.Client()
		})
		err := sender.Send(ctx, srv.URL, webmention.Mention{
			Source: "https://source.com",
			Target: "https://target.com",
		})
		require.Error(t, err)
	})
}
