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
	link, err := Resolve(ctx, "https://t.co/JqumM1uaVE")
	require.NoError(t, err)
	require.Equal(t, "https://zerokspot.com/weblog/2022/03/25/pogo-podman-executor-gitlab/", link)
}
