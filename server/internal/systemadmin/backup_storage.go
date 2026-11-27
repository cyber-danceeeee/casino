package systemadmin

import (
	"casino/internal/storage"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const privateBackupPrefix = "private:"

func privateBackupPath(key string) (string, error) {
	name := strings.TrimPrefix(key, privateBackupPrefix)
	if !strings.HasPrefix(name, "backups/") || strings.Contains(name, "\\") || filepath.Clean(name) != name {
		return "", fmt.Errorf("invalid backup artifact path")
	}
	root := os.Getenv("CASINO_BACKUP_DIR")
	if root == "" {
		root = "backups"
	}
	return filepath.Join(root, strings.TrimPrefix(name, "backups/")), nil
}
func putBackupArtifact(ctx context.Context, key string, data []byte) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	target, err := privateBackupPath(key)
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		return "", err
	}
	file, err := os.CreateTemp(filepath.Dir(target), ".backup-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err != nil {
		file.Close()
		return "", err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return "", err
	}
	if err = file.Close(); err != nil {
		return "", err
	}
	if err = os.Rename(file.Name(), target); err != nil {
		return "", err
	}
	return privateBackupPrefix + key, nil
}
func readBackupArtifact(ctx context.Context, key string, max int64) ([]byte, error) {
	if !strings.HasPrefix(key, privateBackupPrefix) {
		return storage.ReadBytes(ctx, key, max)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	name, err := privateBackupPath(key)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, max+1))
	if int64(len(data)) > max {
		return nil, fmt.Errorf("backup exceeds size limit")
	}
	return data, err
}
func deleteBackupArtifact(ctx context.Context, key string) error {
	if !strings.HasPrefix(key, privateBackupPrefix) {
		return storage.Delete(ctx, key)
	}
	name, err := privateBackupPath(key)
	if err != nil {
		return err
	}
	err = os.Remove(name)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
