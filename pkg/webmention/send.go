package webmention

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rs/zerolog"
)

// SenderConfiguration allows to inject a custom HTTP Client into the
// Sender.
type SenderConfiguration struct {
	HTTPClient *http.Client
}

// SenderConfigurator is used as argument for the NewSender method and
// helps configuring a new Sender instance.
type SenderConfigurator func(*SenderConfiguration)

// Sender is used to send webmentions.
type Sender interface {
	// Send a webmention to the specified endpoint indicating that source
	// mentioned target.
	Send(ctx context.Context, endpoint string, mention Mention) error
}

type simpleSender struct {
	client *http.Client
}

// NewSender creates a configured sender implementation.
func NewSender(configurators ...SenderConfigurator) Sender {
	cfg := &SenderConfiguration{
		HTTPClient: &http.Client{},
	}
	for _, c := range configurators {
		c(cfg)
	}
	return &simpleSender{
		client: cfg.HTTPClient,
	}
}

func (s *simpleSender) Send(ctx context.Context, endpoint string, mention Mention) error {
	logger := zerolog.Ctx(ctx)
	v := url.Values{}
	v.Set("source", mention.Source)
	v.Set("target", mention.Target)
	data := v.Encode()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	logger.Debug().Msgf("Sending mention: %v", r)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.client.Do(r)
	if err != nil {
		return fmt.Errorf("webmention request failed: %w", err)
	}
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code returned: %v", resp.StatusCode)
	}
	return nil
}
