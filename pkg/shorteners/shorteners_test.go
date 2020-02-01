package shorteners

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	ctx := context.Background()
	_, err := Resolve(ctx, "")
	require.Error(t, err, "An empty URL should trigger an error.")
}

func TestTwitterResolver(t *testing.T) {
	ctx := context.Background()
	link, err := Resolve(ctx, "https://t.co/mEnq1oJX3Q?amp=1")
	require.NoError(t, err)
	require.Equal(t, "https://resource-types.concourse-ci.org/", link)
}
