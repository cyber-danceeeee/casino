package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"casino/internal/config"
)

type Asset struct {
	URL       string `json:"url"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mimeType"`
	Key       string `json:"key"`
	ExpiresAt string `json:"expiresAt,omitempty"`
}

func RandomName(extension string) (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	extension = strings.ToLower(strings.TrimSpace(extension))
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	return hex.EncodeToString(random) + extension, nil
}

func UploadFileHeader(ctx context.Context, key string, file *multipart.FileHeader) (Asset, error) {
	if file == nil {
		return Asset{}, errors.New("file is required")
	}
	source, err := file.Open()
	if err != nil {
		return Asset{}, err
	}
	defer source.Close()
	contentType := normalizeContentType(file.Header.Get("Content-Type"), filepath.Ext(file.Filename))
	asset, err := Put(ctx, key, source, file.Size, contentType)
	if err != nil {
		return Asset{}, err
	}
	asset.Name = path.Base(cleanObjectKey(key))
	asset.Size = file.Size
	asset.MimeType = contentType
	return asset, nil
}

func PutBytes(ctx context.Context, key string, data []byte, contentType string) (Asset, error) {
	contentType = normalizeContentType(contentType, filepath.Ext(key))
	return Put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
}

func Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) (Asset, error) {
	key = cleanObjectKey(key)
	if key == "" {
		return Asset{}, errors.New("object key is required")
	}
	if body == nil {
		return Asset{}, errors.New("object body is required")
	}
	contentType = normalizeContentType(contentType, filepath.Ext(key))
	settings := config.StorageConfig()
	driver := strings.ToLower(strings.TrimSpace(settings.Driver))
	if driver == "" || driver == "local" {
		return putLocal(key, body, size, contentType, settings.PublicBaseURL)
	}
	if driver == "oss" {
		return putOSS(ctx, key, body, size, contentType, settings.OSS)
	}
	return Asset{}, fmt.Errorf("unsupported storage driver %q", settings.Driver)
}

func IsAssetURL(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	if strings.HasPrefix(value, "data:") {
		return false
	}
	if strings.HasPrefix(value, "/static/") {
		return true
	}
	parsed, err := url.Parse(value)
	if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
		return true
	}
	return cleanObjectKey(value) != ""
}

func NormalizeReference(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "data:") {
		return ""
	}
	settings := config.StorageConfig()
	if strings.HasPrefix(value, "/static/") {
		if strings.EqualFold(strings.TrimSpace(settings.Driver), "oss") {
			return ossObjectKeyFromStaticReference(value, settings.OSS)
		}
		return value
	}
	if strings.EqualFold(strings.TrimSpace(settings.Driver), "oss") {
		if key, ok := ossObjectKeyFromReference(value, settings.OSS); ok {
			return key
		}
	}
	parsed, err := url.Parse(value)
	if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
		return value
	}
	return cleanObjectKey(value)
}

func PublicURL(ctx context.Context, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	settings := config.StorageConfig()
	driver := strings.ToLower(strings.TrimSpace(settings.Driver))
	if driver == "" || driver == "local" {
		return localPublicURL(value, settings.PublicBaseURL), nil
	}
	if driver == "oss" {
		if key, ok := ossObjectKeyFromReference(value, settings.OSS); ok {
			client, err := ossClient(settings.OSS)
			if err != nil {
				return "", err
			}
			signedURL, _, err := ossSignedGetURL(ctx, client, strings.TrimSpace(settings.OSS.Bucket), key, settings.OSS)
			return signedURL, err
		}
	}
	parsed, err := url.Parse(value)
	if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
		return value, nil
	}
	return value, nil
}

// StablePublicURL returns a URL that can be persisted or reused by browser
// clients without expiring. Private OSS objects are served through the
// backend's /static proxy, which creates a fresh signed request per fetch.
func StablePublicURL(ctx context.Context, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	settings := config.StorageConfig()
	if strings.EqualFold(strings.TrimSpace(settings.Driver), "oss") {
		if key, ok := ossObjectKeyFromReference(value, settings.OSS); ok {
			prefix := strings.Trim(strings.TrimSpace(settings.OSS.ObjectPrefix), "/")
			if prefix != "" {
				if key == prefix {
					key = ""
				} else {
					key = strings.TrimPrefix(key, prefix+"/")
				}
			}
			if key != "" && key != "." {
				return "/static/" + key, nil
			}
		}
	}
	return PublicURL(ctx, value)
}

// Check verifies that the configured backend is reachable without exposing its credentials.
func Check(ctx context.Context) error {
	settings := config.StorageConfig()
	driver := strings.ToLower(strings.TrimSpace(settings.Driver))
	if driver == "" || driver == "local" {
		info, err := os.Stat(config.StaticDir())
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New("local storage root is not a directory")
		}
		return nil
	}
	if driver == "oss" {
		client, err := ossClient(settings.OSS)
		if err != nil {
			return err
		}
		_, err = client.GetBucketInfo(ctx, &oss.GetBucketInfoRequest{Bucket: oss.Ptr(strings.TrimSpace(settings.OSS.Bucket))})
		return err
	}
	return fmt.Errorf("unsupported storage driver %q", settings.Driver)
}

func ReadBytes(ctx context.Context, value string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = 64 << 20
	}
	settings := config.StorageConfig()
	driver := strings.ToLower(strings.TrimSpace(settings.Driver))
	if driver == "" || driver == "local" {
		key := strings.TrimPrefix(NormalizeReference(value), "/static/")
		key = cleanObjectKey(key)
		if key == "" {
			return nil, errors.New("object key is required")
		}
		target := filepath.Join(config.StaticDir(), filepath.FromSlash(key))
		info, err := os.Stat(target)
		if err != nil {
			return nil, err
		}
		if info.Size() > maxBytes {
			return nil, errors.New("object exceeds read limit")
		}
		return os.ReadFile(target)
	}
	if driver == "oss" {
		key, ok := ossObjectKeyFromReference(value, settings.OSS)
		if !ok {
			key = cleanObjectKey(value)
		}
		client, err := ossClient(settings.OSS)
		if err != nil {
			return nil, err
		}
		result, err := client.GetObject(ctx, &oss.GetObjectRequest{Bucket: oss.Ptr(strings.TrimSpace(settings.OSS.Bucket)), Key: oss.Ptr(key)})
		if err != nil {
			return nil, err
		}
		defer result.Body.Close()
		data, err := io.ReadAll(io.LimitReader(result.Body, maxBytes+1))
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > maxBytes {
			return nil, errors.New("object exceeds read limit")
		}
		return data, nil
	}
	return nil, fmt.Errorf("unsupported storage driver %q", settings.Driver)
}

func Exists(ctx context.Context, value string) (bool, error) {
	settings := config.StorageConfig()
	driver := strings.ToLower(strings.TrimSpace(settings.Driver))
	if driver == "" || driver == "local" {
		key := cleanObjectKey(strings.TrimPrefix(NormalizeReference(value), "/static/"))
		if key == "" {
			return false, errors.New("object key is required")
		}
		_, err := os.Stat(filepath.Join(config.StaticDir(), filepath.FromSlash(key)))
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return err == nil, err
	}
	if driver == "oss" {
		key, ok := ossObjectKeyFromReference(value, settings.OSS)
		if !ok {
			key = cleanObjectKey(value)
		}
		client, err := ossClient(settings.OSS)
		if err != nil {
			return false, err
		}
		_, err = client.HeadObject(ctx, &oss.HeadObjectRequest{Bucket: oss.Ptr(strings.TrimSpace(settings.OSS.Bucket)), Key: oss.Ptr(key)})
		return err == nil, err
	}
	return false, fmt.Errorf("unsupported storage driver %q", settings.Driver)
}

func Delete(ctx context.Context, value string) error {
	settings := config.StorageConfig()
	driver := strings.ToLower(strings.TrimSpace(settings.Driver))
	if driver == "" || driver == "local" {
		key := cleanObjectKey(strings.TrimPrefix(NormalizeReference(value), "/static/"))
		if key == "" {
			return errors.New("object key is required")
		}
		return os.Remove(filepath.Join(config.StaticDir(), filepath.FromSlash(key)))
	}
	if driver == "oss" {
		key, ok := ossObjectKeyFromReference(value, settings.OSS)
		if !ok {
			key = cleanObjectKey(value)
		}
		client, err := ossClient(settings.OSS)
		if err != nil {
			return err
		}
		_, err = client.DeleteObject(ctx, &oss.DeleteObjectRequest{Bucket: oss.Ptr(strings.TrimSpace(settings.OSS.Bucket)), Key: oss.Ptr(key)})
		return err
	}
	return fmt.Errorf("unsupported storage driver %q", settings.Driver)
}

func putLocal(key string, body io.Reader, size int64, contentType, publicBaseURL string) (Asset, error) {
	target := filepath.Join(config.StaticDir(), filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return Asset{}, err
	}
	temporary, err := os.CreateTemp(filepath.Dir(target), ".upload-*")
	if err != nil {
		return Asset{}, err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err = io.Copy(temporary, body); err != nil {
		temporary.Close()
		return Asset{}, err
	}
	if err = temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return Asset{}, err
	}
	if err = temporary.Close(); err != nil {
		return Asset{}, err
	}
	if err = os.Rename(temporaryPath, target); err != nil {
		return Asset{}, err
	}
	return Asset{URL: joinURL(publicBaseURL, "/static/"+key), Name: path.Base(key), Size: size, MimeType: contentType, Key: key}, nil
}

func putOSS(ctx context.Context, key string, body io.Reader, size int64, contentType string, settings config.OSSStorage) (Asset, error) {
	bucket := strings.TrimSpace(settings.Bucket)
	client, err := ossClient(settings)
	if err != nil {
		return Asset{}, errors.New("oss storage is not fully configured")
	}
	objectKey := cleanObjectKey(path.Join(strings.Trim(strings.TrimSpace(settings.ObjectPrefix), "/"), key))
	_, err = client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket:        oss.Ptr(bucket),
		Key:           oss.Ptr(objectKey),
		Body:          body,
		ContentLength: oss.Ptr(size),
		ContentType:   oss.Ptr(contentType),
	})
	if err != nil {
		return Asset{}, err
	}
	signedURL, expiresAt, err := ossSignedGetURL(ctx, client, bucket, objectKey, settings)
	if err != nil {
		return Asset{}, err
	}
	return Asset{URL: signedURL, Name: path.Base(key), Size: size, MimeType: contentType, Key: objectKey, ExpiresAt: expiresAt.Format(time.RFC3339)}, nil
}

func ossClient(settings config.OSSStorage) (*oss.Client, error) {
	region := strings.TrimSpace(settings.Region)
	endpoint := strings.TrimSpace(settings.Endpoint)
	if strings.TrimSpace(settings.Bucket) == "" || region == "" || endpoint == "" || strings.TrimSpace(settings.AccessKeyID) == "" || strings.TrimSpace(settings.AccessKeySecret) == "" {
		return nil, errors.New("oss storage is not fully configured")
	}
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(settings.AccessKeyID, settings.AccessKeySecret, settings.SecurityToken)).
		WithRegion(region).
		WithEndpoint(endpoint).
		WithUseCName(settings.UseCName).
		WithUsePathStyle(settings.UsePathStyle).
		WithUseInternalEndpoint(settings.UseInternalEndpoint)
	return oss.NewClient(cfg), nil
}

func ossSignedGetURL(ctx context.Context, client *oss.Client, bucket, objectKey string, settings config.OSSStorage) (string, time.Time, error) {
	expires := ossPresignDuration(settings)
	result, err := client.Presign(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucket),
		Key:    oss.Ptr(objectKey),
	}, oss.PresignExpires(expires))
	if err != nil {
		return "", time.Time{}, err
	}
	return result.URL, result.Expiration, nil
}

func ossPresignDuration(settings config.OSSStorage) time.Duration {
	seconds := settings.SignedURLExpiresSeconds
	if seconds <= 0 {
		seconds = 7200
	}
	const maxV4Seconds = 7 * 24 * 60 * 60
	if seconds > maxV4Seconds {
		seconds = maxV4Seconds
	}
	return time.Duration(seconds) * time.Second
}

func ossObjectKeyFromReference(value string, settings config.OSSStorage) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "data:") {
		return "", false
	}
	if strings.HasPrefix(value, "/static/") {
		key := ossObjectKeyFromStaticReference(value, settings)
		return key, key != ""
	}
	if strings.HasPrefix(value, "oss://") {
		key := cleanObjectKey(strings.TrimPrefix(value, "oss://"))
		return key, key != ""
	}
	parsed, err := url.Parse(value)
	if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
		if !isConfiguredOSSEndpoint(parsed.Host, settings) {
			return "", false
		}
		key := cleanObjectKey(parsed.Path)
		return key, key != ""
	}
	key := cleanObjectKey(value)
	return key, key != ""
}

func ossObjectKeyFromStaticReference(value string, settings config.OSSStorage) string {
	relative := strings.TrimPrefix(strings.TrimSpace(value), "/static/")
	key := cleanObjectKey(relative)
	prefix := strings.Trim(strings.TrimSpace(settings.ObjectPrefix), "/")
	if prefix == "" {
		return key
	}
	return cleanObjectKey(path.Join(prefix, key))
}

func isConfiguredOSSEndpoint(host string, settings config.OSSStorage) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}
	candidates := []string{}
	if endpoint := strings.TrimSpace(settings.Endpoint); endpoint != "" {
		endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
		endpoint = strings.TrimRight(endpoint, "/")
		candidates = append(candidates, endpoint)
		if bucket := strings.TrimSpace(settings.Bucket); bucket != "" && !settings.UseCName {
			candidates = append(candidates, strings.ToLower(bucket+"."+endpoint))
		}
	}
	if base := strings.TrimSpace(settings.PublicBaseURL); base != "" {
		if parsed, err := url.Parse(base); err == nil && parsed.Host != "" {
			candidates = append(candidates, strings.ToLower(parsed.Host))
		}
	}
	for _, candidate := range candidates {
		if host == strings.ToLower(candidate) {
			return true
		}
	}
	return false
}

func localPublicURL(value, publicBaseURL string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
		return value
	}
	if strings.HasPrefix(value, "/static/") {
		return joinURL(publicBaseURL, value)
	}
	return joinURL(publicBaseURL, "/static/"+cleanObjectKey(value))
}

func cleanObjectKey(key string) string {
	key = strings.ReplaceAll(strings.TrimSpace(key), "\\", "/")
	key = path.Clean("/" + key)
	key = strings.TrimPrefix(key, "/")
	if key == "." || key == "" {
		return ""
	}
	parts := strings.Split(key, "/")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "." || part == ".." {
			continue
		}
		cleaned = append(cleaned, part)
	}
	return strings.Join(cleaned, "/")
}

func normalizeContentType(value, extension string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		if mediaType, _, err := mime.ParseMediaType(value); err == nil && mediaType != "" {
			return mediaType
		}
	}
	switch strings.ToLower(strings.TrimSpace(extension)) {
	case ".avif":
		return "image/avif"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	}
	if guessed := mime.TypeByExtension(strings.ToLower(extension)); guessed != "" {
		if mediaType, _, err := mime.ParseMediaType(guessed); err == nil && mediaType != "" {
			return mediaType
		}
	}
	return "application/octet-stream"
}

func joinURL(base, suffix string) string {
	suffix = strings.TrimLeft(strings.TrimSpace(suffix), "/")
	if strings.TrimSpace(base) == "" {
		return "/" + suffix
	}
	return strings.TrimRight(base, "/") + "/" + escapePath(suffix)
}

func escapePath(value string) string {
	parts := strings.Split(value, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}
