package webmention_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/webmention"
)

func TestVerify(t *testing.T) {
	t.Run("handle redirects", func(t *testing.T) {
		ctx := context.Background()
		router := chi.NewRouter()
		router.Get("/initial", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/actual", http.StatusTemporaryRedirect)
		})
		router.Get("/actual", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "<html><body><a href=\"https://zerokspot.com/notes/2020/09/09/podcasts-darknet-diaries/\">text</a></body></html>")
		})
		server := httptest.NewServer(router)
		defer server.Close()
		mention := &webmention.Mention{
			Source: fmt.Sprintf("%s/initial", server.URL),
			Target: "https://zerokspot.com/notes/2020/09/09/podcasts-darknet-diaries/",
		}

		// It should fail if no redirects are allowed
		err := webmention.Verify(ctx, mention, func(o *webmention.VerifyOptions) {
			o.MaxRedirects = 0
		})
		require.Error(t, err)

		// It should work if -1 redirects (infinite) redirects are allowed
		err = webmention.Verify(ctx, mention, func(o *webmention.VerifyOptions) {
			o.MaxRedirects = -1
		})
		require.NoError(t, err)
	})
	t.Run("link exists", func(t *testing.T) {
		ctx := context.Background()
		v := webmention.NewVerifier()
		mention := webmention.Mention{
			Source: "...",
			Target: "https://target.com",
		}
		err := v.Verify(ctx, nil, bytes.NewBufferString("<html><body><a href=\"https://something-else.com\">link</a><a href=\"https://target.com\">link</a></body></html>"), &mention)
		require.NoError(t, err)
	})
	// If the link to be verified is only available as relative
	// URL, then we can extrapolate the full URL from the rest of
	// the request/response:
	t.Run("link exists (relative)", func(t *testing.T) {
		ctx := context.Background()
		v := webmention.NewVerifier()
		mention := webmention.Mention{
			Source: "https://source.com/",
			Target: "https://source.com/target",
		}
		resp := http.Response{
			Request: httptest.NewRequest(http.MethodGet, "https://source.com", nil),
		}
		err := v.Verify(ctx, &resp, bytes.NewBufferString("<html><body><a href=\"https://something-else.com\">link</a><a href=\"/target\">link</a></body></html>"), &mention)
		require.NoError(t, err)
	})
	t.Run("link only in text", func(t *testing.T) {
		ctx := context.Background()
		v := webmention.NewVerifier()
		mention := webmention.Mention{
			Source: "...",
			Target: "https://target.com",
		}
		err := v.Verify(ctx, nil, bytes.NewBufferString("<html><body>https://something-else.com https://target.com</body></html>"), &mention)
		require.Error(t, err)
	})
	t.Run("t.co link", func(t *testing.T) {
		ctx := context.Background()
		v := webmention.NewVerifier()
		mention := webmention.Mention{
			Source: "...",
			Target: "https://resource-types.concourse-ci.org/",
		}
		err := v.Verify(ctx, nil, bytes.NewBufferString("<html><body><a href=\"https://t.co/mEnq1oJX3Q?amp=1\">link</a></body></html>"), &mention)
		require.NoError(t, err)
	})
	t.Run("link doesn't exists", func(t *testing.T) {
		ctx := context.Background()
		v := webmention.NewVerifier()
		mention := webmention.Mention{
			Source: "...",
			Target: "https://target.com",
		}
		err := v.Verify(ctx, nil, bytes.NewBufferString("<html><body><a href=\"https://something-else.com\">link</a></body></html>"), &mention)
		require.Error(t, err)
	})

	t.Run("title-extraction", func(t *testing.T) {
		t.Run("title-present", func(t *testing.T) {
			ctx := context.Background()
			v := webmention.NewVerifier()
			mention := webmention.Mention{
				Source: "...",
				Target: "https://target.com",
			}
			err := v.Verify(ctx, nil, bytes.NewBufferString("<html><head><title>Sample title</title></head><body><a href=\"https://something-else.com\">link</a><a href=\"https://target.com\">link</a></body></html>"), &mention)
			require.NoError(t, err)
			require.Equal(t, "Sample title", mention.Title)
		})
		// If no title is present, the domain name should be used as title:
		t.Run("title-missing", func(t *testing.T) {
			ctx := context.Background()
			v := webmention.NewVerifier()
			mention := webmention.Mention{
				Source: "https://source.com",
				Target: "https://target.com",
			}
			err := v.Verify(ctx, nil, bytes.NewBufferString("<html><head></head><body><a href=\"https://something-else.com\">link</a><a href=\"https://target.com\">link</a></body></html>"), &mention)
			require.NoError(t, err)
			require.Equal(t, "source.com", mention.Title)
		})
	})

	t.Run("h-entry-extraction", func(t *testing.T) {
		ctx := context.Background()
		v := webmention.NewVerifier()
		mention := webmention.Mention{
			Source: "...",
			Target: "https://target.com",
		}
		err := v.Verify(ctx, nil, bytes.NewBufferString("<html><head><title>Sample title</title></head><body><div class=\"h-entry\"><h1 class=\"p-name\">Actual title</h1><a href=\"/\" class=\"u-author h-card\">Author</a><div class=\"e-content\"><p>content</p> <p>next</p></div><a href=\"https://something-else.com\">link</a><a class=\"u-in-reply-to\" href=\"https://target.com\">link</a></div></body></html>"), &mention)
		require.NoError(t, err)
		require.Equal(t, "Actual title", mention.Title)
		require.Equal(t, "content next", mention.Content)
		require.Equal(t, "Author", mention.AuthorName)
		require.Equal(t, "comment", mention.Type)
	})

	t.Run("rsvp-extraction", func(t *testing.T) {
		ctx := context.Background()
		v := webmention.NewVerifier()
		t.Run("non-data", func(t *testing.T) {
			mention := webmention.Mention{
				Source: "...",
				Target: "https://target.com",
			}
			err := v.Verify(ctx, nil, bytes.NewBufferString("<html><head><title>Sample title</title></head><body><div class=\"h-entry\"><h1 class=\"p-name\">Actual title</h1><a href=\"/\" class=\"u-author h-card\">Author</a><div class=\"e-content\"><p>content</p> <p>next</p></div><span class=\"p-rsvp\">yes</span><a class=\"u-in-reply-to\" href=\"https://something-else.com\">link</a><a href=\"https://target.com\">link</a></div></body></html>"), &mention)
			require.NoError(t, err)
			require.Equal(t, "rsvp", mention.Type)
			require.Equal(t, "yes", mention.RSVP)
		})
		t.Run("data", func(t *testing.T) {
			mention := webmention.Mention{
				Source: "...",
				Target: "https://target.com",
			}
			err := v.Verify(ctx, nil, bytes.NewBufferString("<html><head><title>Sample title</title></head><body><div class=\"h-entry\"><h1 class=\"p-name\">Actual title</h1><a href=\"/\" class=\"u-author h-card\">Author</a><div class=\"e-content\"><p>content</p> <p>next</p></div><data class=\"p-rsvp\" value=\"yes\">I'll be there!</data><a class=\"u-in-reply-to\" href=\"https://something-else.com\">link</a><a href=\"https://target.com\">link</a></div></body></html>"), &mention)
			require.NoError(t, err)
			require.Equal(t, "rsvp", mention.Type)
			require.Equal(t, "yes", mention.RSVP)
		})
	})
	t.Run("like", func(t *testing.T) {
		ctx := context.Background()
		v := webmention.NewVerifier()
		mention := webmention.Mention{
			Source: "...",
			Target: "https://target.com",
		}
		err := v.Verify(ctx, nil, bytes.NewBufferString("<html><body class=\"h-entry\"><a href=\"https://something-else.com\">link</a><a href=\"https://target.com\" class=\"u-like-of\">link</a></body></html>"), &mention)
		require.NoError(t, err)
		require.Equal(t, "like", mention.Type)
	})
	t.Run("like of something else", func(t *testing.T) {
		ctx := context.Background()
		v := webmention.NewVerifier()
		mention := webmention.Mention{
			Source: "...",
			Target: "https://target.com",
		}
		err := v.Verify(ctx, nil, bytes.NewBufferString("<html><body class=\"h-entry\"><a href=\"https://something-else.com\" class=\"u-like-of\">link</a><a href=\"https://target.com\">link</a></body></html>"), &mention)
		require.NoError(t, err)
		require.Equal(t, "", mention.Type)
	})
}
