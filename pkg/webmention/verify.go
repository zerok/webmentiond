package webmention

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/zerok/webmentiond/pkg/shorteners"
	"golang.org/x/net/html"
	"willnorris.com/go/microformats"
)

type VerifyOptions struct {
	MaxRedirects int
}

// Verify uses a basic HTTP client and a default Verifier.
func Verify(ctx context.Context, mention *Mention, configurators ...func(c *VerifyOptions)) error {
	cfg := &VerifyOptions{
		MaxRedirects: 10,
	}
	for _, c := range configurators {
		c(cfg)
	}
	client := &http.Client{}
	client.CheckRedirect = func(r *http.Request, via []*http.Request) error {
		if cfg.MaxRedirects > -1 && len(via) > cfg.MaxRedirects {
			return errors.New("too many redirects")
		}
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mention.Source, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	v := NewVerifier()
	defer resp.Body.Close()
	return v.Verify(ctx, resp, resp.Body, mention)
}

// Verifier is used to check if a given response body produced by
// fetching mention.Source contains a link to mention.Target.
type Verifier interface {
	Verify(ctx context.Context, resp *http.Response, body io.Reader, mention *Mention) error
}

type htmlVerifier struct {
}

func resolveURL(u string, resp *http.Response) (string, error) {
	if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
		return u, nil
	}
	if resp == nil || resp.Request == nil {
		return u, nil
	}
	pu, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("failed to parse relative URL")
	}
	ru := resp.Request.URL.ResolveReference(pu)
	if ru == nil {
		return "", fmt.Errorf("failed to resolve URL")
	}
	return ru.String(), nil
}

func (v *htmlVerifier) Verify(ctx context.Context, resp *http.Response, body io.Reader, mention *Mention) error {
	var tokenBuffer bytes.Buffer
	var mfBuffer bytes.Buffer
	sourceURL, err := url.Parse(mention.Source)
	if err != nil {
		return err
	}
	io.Copy(io.MultiWriter(&tokenBuffer, &mfBuffer), body)
	tokenizer := html.NewTokenizer(&tokenBuffer)
	mf := microformats.Parse(&mfBuffer, sourceURL)
	inTitle := false
	inAudio := false
	inVideo := false
	title := ""
	u, err := url.Parse(mention.Source)
	if err == nil {
		title = u.Hostname()
	}
	var contentOK bool
loop:
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.TextToken:
			if inTitle {
				title = strings.TrimSpace(string(tokenizer.Text()))
			}
		case html.EndTagToken:
			tagName, _ := tokenizer.TagName()
			switch string(tagName) {
			case "title":
				inTitle = false
			case "audio":
				inAudio = false
			case "video":
				inVideo = false
			}
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				break loop
			}
			return err
		case html.SelfClosingTagToken:
			fallthrough
		case html.StartTagToken:
			tagName, _ := tokenizer.TagName()
			switch string(tagName) {
			case "title":
				inTitle = true
			case "audio":
				inAudio = true
			case "video":
				inVideo = true
			case "source":
				if inVideo || inAudio {
					src, err := resolveURL(getAttr(tokenizer, "src"), resp)
					if err != nil {
						continue
					}
					if src == mention.Target {
						mention.Title = title
						contentOK = true
						continue
					}
					link, err := shorteners.Resolve(ctx, src)
					if err != nil {
						continue
					}
					if link == mention.Target {
						mention.Title = title
						contentOK = true
						continue
					}
				}
			case "img":
				src, err := resolveURL(getAttr(tokenizer, "src"), resp)
				if err != nil {
					continue
				}
				if src == mention.Target {
					mention.Title = title
					contentOK = true
					continue
				}
				link, err := shorteners.Resolve(ctx, src)
				if err != nil {
					continue
				}
				if link == mention.Target {
					mention.Title = title
					contentOK = true
					continue
				}
			case "a":
				href, err := resolveURL(getAttr(tokenizer, "href"), resp)
				if err != nil {
					continue
				}
				if href == mention.Target {
					mention.Title = title
					contentOK = true
					continue
				}
				link, err := shorteners.Resolve(ctx, href)
				if err != nil {
					continue
				}
				if link == mention.Target {
					mention.Title = title
					contentOK = true
					continue
				}
			}

		}
	}
	if !contentOK {
		return fmt.Errorf("target not found in content")
	}
	mfFillMentionFromData(mention, mf)
	if mention.RSVP != "" {
		mention.Type = "rsvp"
	}
	return nil
}

func mfFillMentionFromData(mention *Mention, mf *microformats.Data) {
	for _, i := range mf.Items {
		mfFillMention(mention, i)
	}
}

func mfFillMention(mention *Mention, mf *microformats.Microformat) bool {
	if mfHasType(mf, "h-entry") {
		if name, ok := mf.Properties["name"]; ok && len(name) > 0 {
			mention.Title = name[0].(string)
		}
		if commented, ok := mf.Properties["in-reply-to"]; ok && len(commented) > 0 {
			if commentedItem, ok := commented[0].(string); ok && commentedItem == mention.Target {
				mention.Type = "comment"
			}
		}
		if commented, ok := mf.Properties["like-of"]; ok && len(commented) > 0 {
			if commentedItem, ok := commented[0].(string); ok && commentedItem == mention.Target {
				mention.Type = "like"
			}
		}
		if rsvp, ok := mf.Properties["rsvp"]; ok && len(rsvp) > 0 {
			mention.RSVP = rsvp[0].(string)
		}
		if contents, ok := mf.Properties["content"]; ok && len(contents) > 0 {
			if content, ok := contents[0].(map[string]string); ok {
				if contentValue, ok := content["value"]; ok {
					mention.Content = contentValue
				}
			}
		}
		if authors, ok := mf.Properties["author"]; ok && len(authors) > 0 {
			if author, ok := authors[0].(*microformats.Microformat); ok {
				if names, ok := author.Properties["name"]; ok && len(names) > 0 {
					mention.AuthorName = names[0].(string)
				}
			}
		}
		return true
	} else if len(mf.Children) > 0 {
		for _, m := range mf.Children {
			if mfFillMention(mention, m) {
				return true
			}
		}
	}
	return false
}

func mfHasType(mf *microformats.Microformat, typ string) bool {
	for _, t := range mf.Type {
		if typ == t {
			return true
		}
	}
	return false
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
