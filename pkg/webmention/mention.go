package webmention

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// ErrUnsupportedMethod is returned by ExtractMention if a non-POST
// request is present.
var ErrUnsupportedMethod = errors.New("unsupported HTTP method")

// ErrUnsupportedContentType is returned by ExtractMention if the POST
// request is not form-urlencoded.
var ErrUnsupportedContentType = errors.New("application/x-www-form-urlencoded content-type required")

// ErrInvalidRequest is returned by ExtractMention if either source or
// target are not provided or not URLs.
var ErrInvalidRequest = errors.New("the request does not contain a source and a target")

// Mention is what is sent to a receiver linking source and target.
type Mention struct {
	Source string
	Target string
	Title  string
}

// ExtractMention parses a given request object and tries to extract
// source and target from it.
func ExtractMention(r *http.Request) (*Mention, error) {
	if r.Method != http.MethodPost {
		return nil, ErrUnsupportedMethod
	}
	ct := r.Header.Get("Content-Type")
	if ct != "application/x-www-form-urlencoded" {
		return nil, ErrUnsupportedContentType
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	form, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, ErrUnsupportedContentType
	}
	source := form.Get("source")
	target := form.Get("target")
	if source == "" || target == "" {
		return nil, ErrInvalidRequest
	}

	if !isAbsoluteURL(source) || !isAbsoluteURL(target) {
		return nil, ErrInvalidRequest
	}
	return &Mention{
		Source: source,
		Target: target,
	}, nil
}

func isAbsoluteURL(s string) bool {
	if !(strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://")) {
		return false
	}
	if _, err := url.Parse(s); err != nil {
		return false
	}
	return true
}
