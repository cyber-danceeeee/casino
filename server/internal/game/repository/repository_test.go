package repository

import (
	"encoding/json"
	"reflect"
	"testing"

	"casino/internal/content"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestReadableScopeSeparatesCatalogAndPublishedContent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Exec("CREATE TABLE readable_content (id INTEGER PRIMARY KEY, site_id INTEGER NOT NULL)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO readable_content VALUES (1,0),(2,1),(3,2)").Error; err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, table, alias string
		scope              content.Scope
		want               []uint64
	}{
		{"catalog", "readable_content", "", content.Scope{Platform: true}, []uint64{1}},
		{"aliased catalog", "readable_content AS g", "g", content.Scope{Platform: true}, []uint64{1}},
		{"H5 readable content", "readable_content AS g", "g", content.Scope{SiteID: 1}, []uint64{1, 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var ids []uint64
			if err := applyReadableScope(db.Table(tc.table), tc.alias, tc.scope).Order("id").Pluck("id", &ids).Error; err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(ids, tc.want) {
				t.Fatalf("readable content: got %v, want %v", ids, tc.want)
			}
		})
	}
}

func TestSyncedGameJSONDefaultsAreValid(t *testing.T) {
	for field, value := range map[string]string{
		"name_i18n":          syncedGameNameI18nDefault,
		"category_ids_draft": syncedGameCategoryIDsDraftDefault,
	} {
		if !json.Valid([]byte(value)) {
			t.Fatalf("%s default is invalid JSON: %q", field, value)
		}
	}
}

func TestSyncedJSONOrDefaultRejectsEmptyDocument(t *testing.T) {
	for _, value := range []string{"", "   ", "{", "not-json"} {
		if got := syncedJSONOrDefault(value, "{}"); got != "{}" {
			t.Fatalf("invalid JSON %q normalized to %q", value, got)
		}
	}
	if got := syncedJSONOrDefault(`{"en":"Game"}`, "{}"); got != `{"en":"Game"}` {
		t.Fatalf("valid JSON was changed: %q", got)
	}
}
