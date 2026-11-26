package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"casino/configs"
	"casino/internal/config"
	"casino/internal/database"
	"casino/internal/storage"
)

type migrationSummary struct {
	UploadedFiles      int
	DeletedLocalFiles  int
	UpdatedRows        int
	UpdatedColumns     int
	SkippedDBColumns   int
	UploadFailures     int
	DatabaseUpdateErrs int
}

func main() {
	var (
		configFile  string
		dryRun      bool
		deleteLocal bool
	)
	flag.StringVar(&configFile, "c", configs.Path("casino.yml"), "configuration file")
	flag.BoolVar(&dryRun, "dry-run", false, "print planned changes without uploading, updating database, or deleting local files")
	flag.BoolVar(&deleteLocal, "delete-local", false, "delete local static files after successful upload and database rewrite")
	flag.Parse()

	if err := run(configFile, dryRun, deleteLocal); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
}

func run(configFile string, dryRun bool, deleteLocal bool) error {
	if err := config.Init(configFile); err != nil {
		return err
	}
	settings := config.StorageConfig()
	if !strings.EqualFold(strings.TrimSpace(settings.Driver), "oss") {
		return fmt.Errorf("storage.driver must be oss for this migration, got %q", settings.Driver)
	}
	if strings.TrimSpace(settings.OSS.Bucket) == "" || strings.TrimSpace(settings.OSS.Endpoint) == "" || strings.TrimSpace(settings.OSS.Region) == "" {
		return errors.New("oss bucket, endpoint, and region are required")
	}
	database.InitDB()
	defer func() { _ = database.CloseDB() }()

	staticRoot := config.StaticDir()
	files, err := collectStaticFiles(staticRoot)
	if err != nil {
		return err
	}
	fmt.Printf("static root: %s\n", staticRoot)
	fmt.Printf("files found: %d\n", len(files))
	if dryRun {
		previewCount := len(files)
		if previewCount > 20 {
			previewCount = 20
		}
		for _, file := range files[:previewCount] {
			fmt.Printf("[dry-run] upload %s -> %s\n", file.LocalPath, file.ObjectKey)
		}
		if len(files) > previewCount {
			fmt.Printf("[dry-run] ... %d more files\n", len(files)-previewCount)
		}
	}

	summary := migrationSummary{}
	replacements := map[string]string{}
	for _, file := range files {
		localReference := "/static/" + file.RelativePath
		if dryRun {
			replacements[localReference] = file.ObjectKey
			continue
		}
		if err := uploadStaticFile(context.Background(), file); err != nil {
			summary.UploadFailures++
			return fmt.Errorf("upload %s: %w", file.LocalPath, err)
		}
		summary.UploadedFiles++
		replacements[localReference] = file.ObjectKey
		if summary.UploadedFiles%50 == 0 {
			fmt.Printf("uploaded %d/%d\n", summary.UploadedFiles, len(files))
		}
	}

	dbSummary, err := rewriteDatabaseReferences(replacements, dryRun)
	if err != nil {
		return err
	}
	summary.UpdatedRows = dbSummary.UpdatedRows
	summary.UpdatedColumns = dbSummary.UpdatedColumns
	summary.SkippedDBColumns = dbSummary.SkippedDBColumns
	summary.DatabaseUpdateErrs = dbSummary.DatabaseUpdateErrs

	if deleteLocal && !dryRun {
		deleted, err := deleteStaticFiles(files, staticRoot)
		if err != nil {
			return err
		}
		summary.DeletedLocalFiles = deleted
	}

	fmt.Printf("done: uploaded=%d updatedRows=%d updatedColumns=%d skippedColumns=%d deletedLocal=%d uploadFailures=%d dbUpdateErrors=%d dryRun=%v\n",
		summary.UploadedFiles,
		summary.UpdatedRows,
		summary.UpdatedColumns,
		summary.SkippedDBColumns,
		summary.DeletedLocalFiles,
		summary.UploadFailures,
		summary.DatabaseUpdateErrs,
		dryRun,
	)
	return nil
}

type staticFile struct {
	LocalPath    string
	RelativePath string
	ObjectKey    string
	Size         int64
}

func collectStaticFiles(root string) ([]staticFile, error) {
	files := []staticFile{}
	err := filepath.WalkDir(root, func(localPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, localPath)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		files = append(files, staticFile{
			LocalPath:    localPath,
			RelativePath: relative,
			ObjectKey:    staticObjectKey(relative),
			Size:         info.Size(),
		})
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return files, nil
	}
	return files, err
}

func uploadStaticFile(ctx context.Context, file staticFile) error {
	source, err := os.Open(file.LocalPath)
	if err != nil {
		return err
	}
	defer source.Close()
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(file.LocalPath)))
	switch strings.ToLower(filepath.Ext(file.LocalPath)) {
	case ".avif":
		contentType = "image/avif"
	case ".webp":
		contentType = "image/webp"
	}
	if contentType == "" {
		buffer := make([]byte, 512)
		n, _ := source.Read(buffer)
		contentType = http.DetectContentType(buffer[:n])
		if _, err = source.Seek(0, io.SeekStart); err != nil {
			return err
		}
	}
	_, err = storage.Put(ctx, file.RelativePath, source, file.Size, contentType)
	return err
}

func staticObjectKey(relative string) string {
	relative = strings.Trim(strings.ReplaceAll(relative, "\\", "/"), "/")
	settings := config.StorageConfig()
	prefix := strings.Trim(strings.TrimSpace(settings.OSS.ObjectPrefix), "/")
	if prefix == "" {
		return relative
	}
	return path.Join(prefix, relative)
}

func rewriteDatabaseReferences(replacements map[string]string, dryRun bool) (migrationSummary, error) {
	summary := migrationSummary{}
	db := database.GetDB()
	rows, err := db.Raw(`
SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE
FROM information_schema.COLUMNS
WHERE table_schema = DATABASE()
  AND DATA_TYPE IN ('char','varchar','text','tinytext','mediumtext','longtext','json')
ORDER BY TABLE_NAME, ORDINAL_POSITION`).Rows()
	if err != nil {
		return summary, err
	}
	columns := []struct {
		TableName  string
		ColumnName string
		DataType   string
	}{}
	for rows.Next() {
		var column struct {
			TableName  string
			ColumnName string
			DataType   string
		}
		if scanErr := rows.Scan(&column.TableName, &column.ColumnName, &column.DataType); scanErr != nil {
			_ = rows.Close()
			return summary, scanErr
		}
		if strings.TrimSpace(column.TableName) == "" || strings.TrimSpace(column.ColumnName) == "" {
			continue
		}
		if !shouldScanColumn(column.ColumnName, column.DataType) {
			continue
		}
		columns = append(columns, column)
	}
	if err = rows.Err(); err != nil {
		_ = rows.Close()
		return summary, err
	}
	_ = rows.Close()

	for _, column := range columns {
		rows, err := db.Raw(fmt.Sprintf(
			"SELECT id, %s AS asset_value FROM %s WHERE %s LIKE ? OR %s LIKE ?",
			quoteIdentifier(column.ColumnName),
			quoteIdentifier(column.TableName),
			quoteIdentifier(column.ColumnName),
			quoteIdentifier(column.ColumnName),
		), "%/static/%", ossLikePattern()).Rows()
		if err != nil {
			summary.SkippedDBColumns++
			continue
		}
		changedIDs := []uint64{}
		for rows.Next() {
			var id uint64
			var value sql.NullString
			if scanErr := rows.Scan(&id, &value); scanErr != nil {
				_ = rows.Close()
				return summary, scanErr
			}
			if !value.Valid {
				continue
			}
			next, changed := rewriteValue(value.String, replacements)
			if !changed {
				continue
			}
			changedIDs = append(changedIDs, id)
			if dryRun {
				continue
			}
			if err = db.Exec(fmt.Sprintf("UPDATE %s SET %s=? WHERE id=?", quoteIdentifier(column.TableName), quoteIdentifier(column.ColumnName)), next, id).Error; err != nil {
				summary.DatabaseUpdateErrs++
				_ = rows.Close()
				return summary, fmt.Errorf("update %s.%s id=%d: %w", column.TableName, column.ColumnName, id, err)
			}
		}
		if err = rows.Err(); err != nil {
			_ = rows.Close()
			return summary, err
		}
		_ = rows.Close()
		if len(changedIDs) > 0 {
			summary.UpdatedColumns++
			summary.UpdatedRows += len(changedIDs)
			fmt.Printf("%s.%s rows=%d\n", column.TableName, column.ColumnName, len(changedIDs))
		}
	}
	return summary, nil
}

func shouldScanColumn(columnName string, dataType string) bool {
	dataType = strings.ToLower(strings.TrimSpace(dataType))
	if dataType == "json" {
		return true
	}
	name := strings.ToLower(strings.TrimSpace(columnName))
	for _, fragment := range []string{
		"url", "uri", "path", "link", "image", "img", "icon", "logo", "favicon", "banner",
		"cover", "thumbnail", "thumb", "avatar", "poster", "media", "asset", "file",
		"audio", "music", "sound", "video", "animation", "lottie", "background", "bg",
		"config", "theme", "content", "html", "metadata", "payload", "response", "request",
		"body", "data", "rules", "manifest", "preview",
	} {
		if strings.Contains(name, fragment) {
			return true
		}
	}
	return false
}

func rewriteValue(value string, replacements map[string]string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var data any
		if err := json.Unmarshal([]byte(value), &data); err == nil {
			next, changed := rewriteJSON(data, replacements)
			if changed {
				encoded, err := json.Marshal(next)
				if err == nil {
					return string(encoded), true
				}
			}
		}
	}
	return rewriteString(value, replacements)
}

func rewriteJSON(value any, replacements map[string]string) (any, bool) {
	switch typed := value.(type) {
	case string:
		return rewriteString(typed, replacements)
	case []any:
		changed := false
		next := make([]any, len(typed))
		for index, item := range typed {
			rewritten, itemChanged := rewriteJSON(item, replacements)
			next[index] = rewritten
			changed = changed || itemChanged
		}
		return next, changed
	case map[string]any:
		changed := false
		next := make(map[string]any, len(typed))
		for key, item := range typed {
			rewritten, itemChanged := rewriteJSON(item, replacements)
			next[key] = rewritten
			changed = changed || itemChanged
		}
		return next, changed
	default:
		return value, false
	}
}

func rewriteString(value string, replacements map[string]string) (string, bool) {
	next := value
	keys := make([]string, 0, len(replacements))
	for key := range replacements {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, key := range keys {
		next = strings.ReplaceAll(next, key, replacements[key])
	}
	next = rewriteStandaloneStaticPath(next)
	next = rewriteOSSReference(next)
	return next, next != value
}

func rewriteStandaloneStaticPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "/static/") {
		return value
	}
	key := staticObjectKey(strings.TrimPrefix(trimmed, "/static/"))
	return strings.Replace(value, trimmed, key, 1)
}

func rewriteOSSReference(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.Contains(trimmed, " ") {
		return value
	}
	key := storage.NormalizeReference(trimmed)
	if key == "" || key == trimmed || strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") || strings.HasPrefix(key, "/static/") {
		return value
	}
	if parsed, err := url.Parse(trimmed); err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
		return strings.Replace(value, trimmed, key, 1)
	}
	return value
}

func ossLikePattern() string {
	settings := config.StorageConfig()
	if base := strings.TrimSpace(settings.OSS.PublicBaseURL); base != "" {
		if parsed, err := url.Parse(base); err == nil && parsed.Host != "" {
			return "%" + parsed.Host + "%"
		}
	}
	endpoint := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(settings.OSS.Endpoint), "https://"), "http://")
	endpoint = strings.TrimRight(endpoint, "/")
	if endpoint != "" && strings.TrimSpace(settings.OSS.Bucket) != "" {
		return "%" + strings.TrimSpace(settings.OSS.Bucket) + "." + endpoint + "%"
	}
	return "%oss%"
}

func deleteStaticFiles(files []staticFile, root string) (int, error) {
	deleted := 0
	for _, file := range files {
		if err := os.Remove(file.LocalPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return deleted, err
		}
		deleted++
	}
	dirs := []string{}
	_ = filepath.WalkDir(root, func(localPath string, entry os.DirEntry, err error) error {
		if err != nil || !entry.IsDir() || localPath == root {
			return nil
		}
		dirs = append(dirs, localPath)
		return nil
	})
	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	for _, dir := range dirs {
		_ = os.Remove(dir)
	}
	keep := filepath.Join(root, ".gitkeep")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return deleted, err
	}
	if _, err := os.Stat(keep); errors.Is(err, os.ErrNotExist) {
		if err = os.WriteFile(keep, []byte{}, 0o644); err != nil {
			return deleted, err
		}
	}
	return deleted, nil
}

func quoteIdentifier(value string) string {
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}
