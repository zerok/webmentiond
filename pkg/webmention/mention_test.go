package webmention_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/webmention"
)

func TestExtractMention(t *testing.T) {
	tests := []struct {
		Label          string
		Method         string
		Body           io.Reader
		ContentType    string
		ResultHasError bool
		ResultMention  *webmention.Mention
	}{
		{
			Label:          "valid mention",
			Method:         http.MethodPost,
			ContentType:    "application/x-www-form-urlencoded",
			Body:           bytes.NewBufferString("target=https://target.com&source=https://source.com"),
			ResultHasError: false,
			ResultMention: &webmention.Mention{
				Source: "https://source.com",
				Target: "https://target.com",
			},
		},
		{
			Label:          "no content-type",
			Method:         http.MethodPost,
			Body:           bytes.NewBufferString("target=https://target.com&source=https://source.com"),
			ResultHasError: true,
			ResultMention:  nil,
		},
		{
			Label:          "get method",
			Method:         http.MethodGet,
			Body:           nil,
			ResultHasError: true,
			ResultMention:  nil,
		},
		{
			Label:          "source missing",
			Method:         http.MethodPost,
			ContentType:    "application/x-www-form-urlencoded",
			Body:           bytes.NewBufferString("target=https://target.com"),
			ResultHasError: true,
			ResultMention:  nil,
		},
		{
			Label:          "target missing",
			Method:         http.MethodPost,
			ContentType:    "application/x-www-form-urlencoded",
			Body:           bytes.NewBufferString("source=https://source.com"),
			ResultHasError: true,
			ResultMention:  nil,
		},
		{
			Label:          "source not a URL",
			Method:         http.MethodPost,
			ContentType:    "application/x-www-form-urlencoded",
			Body:           bytes.NewBufferString("target=https://target.com&source=test"),
			ResultHasError: true,
			ResultMention:  nil,
		},
		{
			Label:          "target not a URL",
			Method:         http.MethodPost,
			ContentType:    "application/x-www-form-urlencoded",
			Body:           bytes.NewBufferString("target=test&source=https://source.com"),
			ResultHasError: true,
			ResultMention:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Label, func(t *testing.T) {
			r, _ := http.NewRequest(test.Method, "", test.Body)
			r.Header.Set("Content-Type", test.ContentType)
			m, err := webmention.ExtractMention(r)
			if test.ResultHasError {
				require.Error(t, err)
				require.Nil(t, m)
			} else {
				require.NoError(t, err)
				require.NotNil(t, m)
				require.Equal(t, test.ResultMention, m)
			}
		})
	}
}
