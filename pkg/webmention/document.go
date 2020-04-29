package webmention

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type Document struct {
	u     *url.URL
	title string
	links []string
}

func (d *Document) Links() []string {
	return d.links
}

func (d *Document) ExternalLinks() []string {
	es := make([]string, 0, 10)
	for _, l := range d.links {
		lu, err := url.Parse(l)
		if err != nil || lu == nil {
			continue
		}
		if lu.Scheme != "http" && lu.Scheme != "https" {
			continue
		}
		if lu.Host == d.u.Host {
			continue
		}
		es = append(es, l)
	}
	return es
}

func DocumentFromURL(ctx context.Context, u string) (*Document, error) {
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return DocumentFromReader(ctx, resp.Body, u)
}

func DocumentFromReader(ctx context.Context, reader io.Reader, u string) (*Document, error) {
	pu, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	doc := &Document{
		u:     pu,
		links: make([]string, 0, 10),
	}
	tokenizer := html.NewTokenizer(reader)
loop:
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				break loop
			}
			return nil, err
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			if string(name) == "a" {
				link := getAttr(tokenizer, "href")
				if link != "" {
					doc.links = append(doc.links, link)
				}
			}
		}
	}
	return doc, nil
}
