package webmention

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

// Verifier is used to check if a given response body produced by
// fetching mention.Source contains a link to mention.Target.
type Verifier interface {
	Verify(ctx context.Context, resp *http.Response, body io.Reader, mention Mention) error
}

type htmlVerifier struct {
}

func (v *htmlVerifier) Verify(ctx context.Context, resp *http.Response, body io.Reader, mention Mention) error {
	tokenizer := html.NewTokenizer(body)
loop:
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				break loop
			}
			return err
		case html.StartTagToken:
			tagName, _ := tokenizer.TagName()
			switch string(tagName) {
			case "a":
				href := getAttr(tokenizer, "href")
				if href == mention.Target {
					return nil
				}
			}

		}
	}
	return fmt.Errorf("target not found in content")
}

// NewVerifier creates a new verifier instance.
func NewVerifier() Verifier {
	return &htmlVerifier{}
}

func getAttr(tokenizer *html.Tokenizer, attr string) string {
	var result string
	for {
		key, value, more := tokenizer.TagAttr()
		if string(key) == attr {
			result = string(value)
		}
		if !more {
			break
		}
	}
	return result
}
