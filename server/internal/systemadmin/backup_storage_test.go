package systemadmin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBackupArtifactsArePrivatePersistentAndBounded(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CASINO_BACKUP_DIR", root)
	ctx := context.Background()
	key, err := putBackupArtifact(ctx, "backups/2026/11/example.tar.gz", []byte("verified-backup"))
	require.NoError(t, err)
	require.Equal(t, "private:backups/2026/11/example.tar.gz", key)
	name := filepath.Join(root, "2026/11/example.tar.gz")
	info, err := os.Stat(name)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	contents, err := readBackupArtifact(ctx, key, 100)
	require.NoError(t, err)
	require.Equal(t, []byte("verified-backup"), contents)
	_, err = readBackupArtifact(ctx, key, 2)
	require.Error(t, err)
	_, err = putBackupArtifact(ctx, "backups/../../static/leak", []byte("invalid"))
	require.Error(t, err)
	require.NoError(t, deleteBackupArtifact(ctx, key))
	_, err = os.Stat(name)
	require.True(t, os.IsNotExist(err))
}
