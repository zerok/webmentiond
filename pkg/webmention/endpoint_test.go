package webmention_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/webmention"
)

func TestDiscoverEndpoint(t *testing.T) {
	t.Run("discover link header", func(t *testing.T) {
		ctx := context.Background()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Link", "</endpoint/>; rel=\"webmention\"")
		}))
		disc := webmention.NewEndpointDiscoverer(func(c *webmention.EndpointDiscoveryConfiguration) {
			c.HTTPClient = srv.Client()
		})
		discovered, err := disc.DiscoverEndpoint(ctx, srv.URL)
		require.NoError(t, err)
		require.Equal(t, srv.URL+"/endpoint/", discovered)
	})

	t.Run("discover link header w/o quotes", func(t *testing.T) {
		ctx := context.Background()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Link", "</endpoint/>; rel=webmention")
		}))
		disc := webmention.NewEndpointDiscoverer(func(c *webmention.EndpointDiscoveryConfiguration) {
			c.HTTPClient = srv.Client()
		})
		discovered, err := disc.DiscoverEndpoint(ctx, srv.URL)
		require.NoError(t, err)
		require.Equal(t, srv.URL+"/endpoint/", discovered)
	})

	t.Run("discover <link>", func(t *testing.T) {
		ctx := context.Background()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			fmt.Fprintf(w, "<html><head><link rel=\"webmention\" href=\"/endpoint/\"></head><body></body></html>")
		}))
		disc := webmention.NewEndpointDiscoverer(func(c *webmention.EndpointDiscoveryConfiguration) {
			c.HTTPClient = srv.Client()
		})
		discovered, err := disc.DiscoverEndpoint(ctx, srv.URL)
		require.NoError(t, err)
		require.Equal(t, srv.URL+"/endpoint/", discovered)
	})

	t.Run("discover <link/>", func(t *testing.T) {
		ctx := context.Background()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			fmt.Fprintf(w, "<html><head><link rel=\"webmention\" href=\"/endpoint/\" /></head><body></body></html>")
		}))
		disc := webmention.NewEndpointDiscoverer(func(c *webmention.EndpointDiscoveryConfiguration) {
			c.HTTPClient = srv.Client()
		})
		discovered, err := disc.DiscoverEndpoint(ctx, srv.URL)
		require.NoError(t, err)
		require.Equal(t, srv.URL+"/endpoint/", discovered)
	})

	t.Run("discover <link/> with absolute URL", func(t *testing.T) {
		ctx := context.Background()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			fmt.Fprintf(w, "<html><head><link rel=\"webmention\" href=\"https://absolute-endpoint.com/endpoint/\" /></head><body></body></html>")
		}))
		disc := webmention.NewEndpointDiscoverer(func(c *webmention.EndpointDiscoveryConfiguration) {
			c.HTTPClient = srv.Client()
		})
		discovered, err := disc.DiscoverEndpoint(ctx, srv.URL)
		require.NoError(t, err)
		require.Equal(t, "https://absolute-endpoint.com/endpoint/", discovered)
	})

	t.Run("discover <a>", func(t *testing.T) {
		ctx := context.Background()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			fmt.Fprintf(w, "<html><head><link rel=\"not-webmention\" href=\"/endpoint/\"></head><body><a href=\"/endpoint/\" rel=\"webmention\">mention!</a></body></html>")
		}))
		disc := webmention.NewEndpointDiscoverer(func(c *webmention.EndpointDiscoveryConfiguration) {
			c.HTTPClient = srv.Client()
		})
		discovered, err := disc.DiscoverEndpoint(ctx, srv.URL)
		require.NoError(t, err)
		require.Equal(t, srv.URL+"/endpoint/", discovered)
	})
}
