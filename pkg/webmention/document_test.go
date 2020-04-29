package webmention_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/webmention"
)

func TestDocumentFromURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/document.html")
	}))
	doc, err := webmention.DocumentFromURL(context.Background(), srv.URL)
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, doc.Links(), []string{"https://test1.com", "https://test2.com"})
}

func TestDocumentFromReader(t *testing.T) {
	fp, err := os.Open("testdata/document.html")
	require.NoError(t, err)
	defer fp.Close()
	doc, err := webmention.DocumentFromReader(context.Background(), fp, "https://test1.com")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, doc.Links(), []string{"https://test1.com", "https://test2.com"})
	require.Equal(t, doc.ExternalLinks(), []string{"https://test2.com"})
}
