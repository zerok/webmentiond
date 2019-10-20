package webmention

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// EndpointDiscoveryConfiguration allows to pass configuration
// parameters to a new Discoverer.
type EndpointDiscoveryConfiguration struct {
	HTTPClient *http.Client
}

// EndpointDiscoveryConfigurator is passed to NewEndDiscoverer to
// configure it.
type EndpointDiscoveryConfigurator func(*EndpointDiscoveryConfiguration)

// EndpointDiscoverer is used to discover the Webmention endpoint for
// the given URL.
type EndpointDiscoverer interface {
	DiscoverEndpoint(ctx context.Context, url string) (string, error)
}

type simpleEndpointDiscoverer struct {
	client *http.Client
}

var linkHeaderRe = regexp.MustCompile("<([^>]+)>;\\s*rel=\"webmention\"")

func (ed *simpleEndpointDiscoverer) DiscoverEndpoint(ctx context.Context, u string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	resp, err := ed.client.Do(req)
	if err != nil {
		return "", err
	}
	var endpointCandidate string
	linkHeader := resp.Header.Get("Link")
	if matches := linkHeaderRe.FindStringSubmatch(linkHeader); len(matches) > 1 {
		endpointCandidate = matches[1]
	}
	tokenizer := html.NewTokenizer(resp.Body)
tokenloop:
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break tokenloop
			}
			return "", err
		}
		if tt == html.StartTagToken {
			tn, _ := tokenizer.TagName()
			switch string(tn) {
			case "a":
				fallthrough
			case "link":
				var rel string
				var href string
				for {
					rawattrkey, attrval, hasMore := tokenizer.TagAttr()
					attrkey := string(rawattrkey)
					switch attrkey {
					case "rel":
						rel = string(attrval)
					case "href":
						href = string(attrval)
					}
					if !hasMore {
						break
					}
				}
				if rel == "webmention" {
					endpointCandidate = href
					break tokenloop
				}
			}
		}
	}
	defer resp.Body.Close()
	if strings.HasPrefix(endpointCandidate, "/") {
		baseURL := *req.URL
		baseURL.Path = endpointCandidate
		return baseURL.String(), nil
	}
	return linkHeader, nil
}

// NewEndpointDiscoverer creates a new EndpointDiscoverer configured
// with the given configurators.
func NewEndpointDiscoverer(configurators ...EndpointDiscoveryConfigurator) EndpointDiscoverer {
	cfg := &EndpointDiscoveryConfiguration{
		HTTPClient: &http.Client{},
	}
	for _, c := range configurators {
		c(cfg)
	}
	return &simpleEndpointDiscoverer{
		client: cfg.HTTPClient,
	}
}
