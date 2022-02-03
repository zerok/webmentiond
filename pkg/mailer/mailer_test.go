package mailer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

type mailhogMessages struct {
	Total int `json:"total"`
}

func TestWithoutTLS(t *testing.T) {
	ctx := context.Background()
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)
	res, err := pool.Run("mailhog/mailhog", "latest", []string{})
	require.NoError(t, err)
	defer pool.Purge(res)
	m := New(res.GetHostPort("1025/tcp"), nil, WithTLS(false))
	require.NoError(t, m.SendMail(ctx, "from@zerokspot.com", []string{"to@zerokspot.com"}, "Test email", "Some body content"))
	requireMailhogMessageTotal(t, res, 1)
}

func requireMailhogMessageTotal(t *testing.T, res *dockertest.Resource, total int) {
	t.Helper()
	c := http.Client{}
	resp, err := c.Get(fmt.Sprintf("http://%s/api/v2/messages", res.GetHostPort("8025/tcp")))
	require.NoError(t, err)
	defer resp.Body.Close()
	var messages mailhogMessages
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&messages))
	require.Equal(t, total, messages.Total)
}
