package mailer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type mailhogMessages struct {
	Total int `json:"total"`
}

func setupMailhog(t *testing.T) (string, string) {
	t.Helper()
	ctx := context.Background()
	smtpAddr := os.Getenv("MAILHOG_SMTP_ADDR")
	apiAddr := os.Getenv("MAILHOG_API_ADDR")
	if smtpAddr != "" && apiAddr != "" {
		return smtpAddr, apiAddr
	}

	cr := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mailhog/mailhog:latest",
			ExposedPorts: []string{"1025/tcp", "8025/tcp"},
			WaitingFor:   wait.ForListeningPort("1025/tcp"),
		},
		Started: true,
	}
	mailhog, err := testcontainers.GenericContainer(ctx, cr)
	require.NoError(t, err)
	t.Cleanup(func() {
		mailhog.Terminate(context.Background())
	})

	smtpAddr, err = mailhog.PortEndpoint(ctx, "1025/tcp", "")
	require.NoError(t, err)
	apiAddr, err = mailhog.PortEndpoint(ctx, "8025/tcp", "")
	require.NoError(t, err)
	return smtpAddr, apiAddr
}

func TestWithoutTLS(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	smtpEndpoint, apiEndpoint := setupMailhog(t)
	m := New(smtpEndpoint, nil, WithTLS(false))
	require.NoError(t, m.SendMail(ctx, "from@zerokspot.com", []string{"to@zerokspot.com"}, "Test email", "Some body content"))
	requireMailhogMessageTotal(t, apiEndpoint, 1)
}

func requireMailhogMessageTotal(t *testing.T, apiEndpoint string, total int) {
	t.Helper()
	c := http.Client{}
	resp, err := c.Get(fmt.Sprintf("http://%s/api/v2/messages", apiEndpoint))
	require.NoError(t, err)
	defer resp.Body.Close()
	var messages mailhogMessages
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&messages))
	require.Equal(t, total, messages.Total)
}
