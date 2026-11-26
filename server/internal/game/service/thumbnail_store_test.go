package service

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadGameThumbnail(t *testing.T) {
	pngData, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=")
	require.NoError(t, err)
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "image/png")
		_, _ = writer.Write(pngData)
	}))
	defer server.Close()

	previousClient := thumbnailHTTPClient
	thumbnailHTTPClient = server.Client()
	t.Cleanup(func() { thumbnailHTTPClient = previousClient })
	t.Chdir(t.TempDir())

	localURL, err := downloadGameThumbnail(context.Background(), "softaggregator", "game/one", server.URL+"/cover")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(localURL, "/static/games/softaggregator/game-one-"))
	require.True(t, strings.HasSuffix(localURL, ".png"))
	data, err := os.ReadFile(filepath.FromSlash("." + localURL))
	require.NoError(t, err)
	require.Equal(t, pngData, data)
}

func TestDownloadGameThumbnailRejectsNonHTTPS(t *testing.T) {
	_, err := downloadGameThumbnail(context.Background(), "provider", "game", "http://example.com/cover.png")
	require.EqualError(t, err, "thumbnail URL must use HTTPS")
}
