package webmention

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
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

var linkHeaderRe = regexp.MustCompile("<([^>]+)>;\\s*rel=\"?webmention\"?")

func (ed *simpleEndpointDiscoverer) DiscoverEndpoint(ctx context.Context, u string) (string, error) {
	logger := zerolog.Ctx(ctx)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	resp, err := ed.client.Do(req)
	if err != nil {
		return "", err
	}
	var endpointCandidate string
	var candidateFound bool
	logger.Debug().Msg("Checking for endpoint in header")
	linkHeaders := resp.Header.Values("Link")
	for _, linkHeader := range linkHeaders {
		if matches := linkHeaderRe.FindStringSubmatch(linkHeader); len(matches) > 1 {
			endpointCandidate = matches[1]
			candidateFound = true
			break
		}
	}
	if !candidateFound {
		logger.Debug().Msg("Checking for endpoint in content")
		tokenizer := html.NewTokenizer(resp.Body)
	tokenloop:
		for {
			tt := tokenizer.Next()
			switch tt {
			case html.ErrorToken:
				err := tokenizer.Err()
				if err == io.EOF {
					break tokenloop
				}
				return "", err
			case html.SelfClosingTagToken:
				fallthrough
			case html.EndTagToken:
				fallthrough
			case html.StartTagToken:
				tn, _ := tokenizer.TagName()
				switch string(tn) {
				case "a":
					fallthrough
				case "link":
					var rel string
					var href string
					var hrefPresent bool
					for {
						rawattrkey, attrval, hasMore := tokenizer.TagAttr()
						attrkey := string(rawattrkey)
						switch attrkey {
						case "rel":
							rel = string(attrval)
						case "href":
							hrefPresent = true
							href = string(attrval)
						}
						if !hasMore {
							break
						}
					}
					if hrefPresent && isCorrectRel(rel) {
						endpointCandidate = href
						candidateFound = true
						break tokenloop
					}
				}
			}
		}
		defer resp.Body.Close()
	}
	if endpointCandidate == "" && candidateFound {
		endpointCandidate = u
	}
	if strings.HasPrefix(endpointCandidate, "/") {
		baseURL := *req.URL
		path, err := url.Parse(endpointCandidate)
		if err != nil {
			return "", err
		}
		baseURL.Path = path.Path
		baseURL.RawQuery = path.RawQuery
		return baseURL.String(), nil
	}
	return endpointCandidate, nil
}

func isCorrectRel(rel string) bool {
	rels := strings.Split(rel, " ")
	for _, r := range rels {
		if strings.TrimSpace(r) == "webmention" {
			return true
		}
	}
	return false
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
