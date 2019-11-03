package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/server"
)

func TestRequestPolicyAllowHost(t *testing.T) {
	policy := server.RequestPolicyAllowHost("domain.com")
	require.True(t, policy(httptest.NewRequest(http.MethodGet, "http://domain.com/lala", nil)))
	require.True(t, policy(httptest.NewRequest(http.MethodGet, "http://domain.com:80/lala", nil)))
	require.True(t, policy(httptest.NewRequest(http.MethodGet, "https://domain.com:443/lala", nil)))
	require.False(t, policy(httptest.NewRequest(http.MethodGet, "http://other-domain.com/lala", nil)))
	require.False(t, policy(httptest.NewRequest(http.MethodGet, "/", nil)))
}
