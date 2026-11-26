package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"casino/internal/storage"
)

const maxThumbnailBytes int64 = 10 << 20

var safeFilePartPattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
var thumbnailHTTPClient = &http.Client{Timeout: 20 * time.Second}

func downloadGameThumbnail(ctx context.Context, providerCode string, providerGameID string, source string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(source))
	if err != nil {
		return "", fmt.Errorf("parse thumbnail URL: %w", err)
	}
	if parsed.Scheme != "https" || parsed.Host == "" {
		return "", errors.New("thumbnail URL must use HTTPS")
	}

	hash := sha256.Sum256([]byte(source))
	providerDir := safeFilePart(providerCode)
	baseName := safeFilePart(providerGameID) + "-" + hex.EncodeToString(hash[:6])

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create thumbnail request: %w", err)
	}
	request.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/gif")
	response, err := thumbnailHTTPClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("download thumbnail: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("download thumbnail: HTTP %d", response.StatusCode)
	}
	if response.ContentLength > maxThumbnailBytes {
		return "", errors.New("thumbnail exceeds 10 MiB")
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, maxThumbnailBytes+1))
	if err != nil {
		return "", fmt.Errorf("read thumbnail: %w", err)
	}
	if int64(len(data)) > maxThumbnailBytes {
		return "", errors.New("thumbnail exceeds 10 MiB")
	}
	contentType := http.DetectContentType(data)
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("unsupported thumbnail content type %q", contentType)
	}
	extension := thumbnailExtension(contentType)
	if extension == "" {
		return "", fmt.Errorf("unsupported thumbnail content type %q", contentType)
	}
	filename := baseName + extension
	asset, err := storage.PutBytes(ctx, "games/"+providerDir+"/"+filename, data, contentType)
	if err != nil {
		return "", fmt.Errorf("store thumbnail: %w", err)
	}
	return storage.NormalizeReference(asset.URL), nil
}

func safeFilePart(value string) string {
	value = strings.Trim(safeFilePartPattern.ReplaceAllString(strings.TrimSpace(value), "-"), "-.")
	if value == "" {
		return "unknown"
	}
	return value
}

func thumbnailExtension(contentType string) string {
	contentType, _, _ = mime.ParseMediaType(contentType)
	switch contentType {
	case "image/avif":
		return ".avif"
	case "image/webp":
		return ".webp"
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}
