package policies_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/policies"
)

func TestDefaultPolicy(t *testing.T) {
	reg := policies.NewRegistry(policies.REJECT)
	require.Equal(t, policies.REJECT, reg.DetermineForURL("https://domain.com"))

	reg = policies.NewRegistry(policies.APPROVE)
	require.Equal(t, policies.APPROVE, reg.DetermineForURL("https://domain.com"))
}

func TestMatchingPolicy(t *testing.T) {
	reg := policies.NewRegistry(policies.REJECT)
	reg.AddPolicy("domain.com", policies.APPROVE, 1)
	require.Equal(t, policies.REJECT, reg.DetermineForURL("https://domain2.com"))
	require.Equal(t, policies.APPROVE, reg.DetermineForURL("https://domain.com"))
}

func TestConflictingPolicies(t *testing.T) {
	reg := policies.NewRegistry(policies.DEFAULT)
	reg.AddPolicy("^https://domain.com", policies.REJECT, 10)
	reg.AddPolicy("^https://domain.com", policies.APPROVE, 1)
	require.Equal(t, policies.APPROVE, reg.DetermineForURL("https://domain.com"))
	require.Equal(t, policies.DEFAULT, reg.DetermineForURL("http://domain.com"))
}

func TestLoad(t *testing.T) {
	reg := policies.NewRegistry(policies.REJECT)
	require.NoError(t, reg.Load(context.Background(), policies.StaticLoader([]policies.URLPolicy{
		{
			URLPattern: regexp.MustCompile("domain.com"),
			Policy:     policies.REJECT,
			Weight:     10,
		},
		{
			URLPattern: regexp.MustCompile("domain.com"),
			Policy:     policies.APPROVE,
			Weight:     1,
		},
	})))
	// If the loader returns nil, then the existing policies are kept
	require.NoError(t, reg.Load(context.Background(), policies.StaticLoader(nil)))
	require.Equal(t, policies.APPROVE, reg.DetermineForURL("https://domain.com"))
}
