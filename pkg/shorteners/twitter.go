package shorteners

import (
	"context"
	"fmt"
	"net/http"
)

type twitterResolver struct{}

func (r *twitterResolver) Resolve(ctx context.Context, link string) (string, error) {
	client := http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, link, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode > 400 {
		return "", fmt.Errorf("unexpected status code returned: %d", resp.StatusCode)
	}
	return resp.Header.Get("Location"), nil
}

func init() {
	registerResolver("https://t.co/", &twitterResolver{})
}
