package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestH5SocialLinksPreferCurrentSettingsAndKeepDisabledDefaultsHidden(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, raw.Close()) })
	require.NoError(t, db.Exec(`CREATE TABLE game_social_links (
		id INTEGER PRIMARY KEY, site_id INTEGER, platform_code TEXT, label TEXT,
		icon_url TEXT, share_url TEXT, share_title TEXT, share_text TEXT,
		share_description TEXT, sort INTEGER, status INTEGER, deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO game_social_links
		(id,site_id,platform_code,label,share_url,sort,status) VALUES
		(1,0,'facebook','Default Facebook','https://default.example',1,1),
		(2,1,'facebook','Current Facebook','https://current.example',3,1),
		(3,0,'telegram','Default Telegram','https://telegram.example',2,1),
		(4,1,'telegram','Disabled Telegram','https://disabled.example',2,2),
		(5,0,'instagram','Default Instagram','https://instagram.example',1,1),
		(6,2,'foreign','Other Installation','https://foreign.example',0,1),
		(7,1,' FACEBOOK ','Duplicate Variant','https://duplicate.example',4,1)
	`).Error)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	links := h5SocialLinks(c, db)
	require.Len(t, links, 2)
	require.Equal(t, "instagram", links[0]["id"])
	require.Equal(t, "facebook", links[1]["id"])
	require.Equal(t, "Current Facebook", links[1]["label"])
	require.Equal(t, "https://current.example", links[1]["url"])
	// Disabling the remaining rows must yield an empty list, not default icons.
	require.NoError(t, db.Exec("UPDATE game_social_links SET status=2").Error)
	require.Empty(t, h5SocialLinks(c, db))
}
