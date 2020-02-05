package webmention_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/webmention"
)

func TestVerify(t *testing.T) {
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
}
