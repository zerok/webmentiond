package shorteners

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

type twitterResolver struct{}

func (r *twitterResolver) Resolve(ctx context.Context, link string) (string, error) {
	client := http.Client{}
	// Disable HTTP/2 support for now as there are issues with t.co and Go
	// since 2022-03-26.
	tr := http.Transport{
		TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
	}
	client.Transport = &tr
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
