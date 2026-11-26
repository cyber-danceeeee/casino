package systemadmin

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEnsureMenuCatalogSkipsWhenCurrent(t *testing.T) {
	db := newMenuCatalogTestDB(t)
	fixed := time.Date(2026, 7, 21, 8, 0, 0, 0, time.UTC)
	if err := db.Exec("INSERT INTO game_admin_roles (id, code, deleted_at) VALUES (1, 'super-admin', NULL)").Error; err != nil {
		t.Fatal(err)
	}
	for index, code := range adminMenuCodes() {
		id := index + 1
		if err := db.Exec("INSERT INTO game_admin_menus (id, code, route_path, icon, visible, enabled, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, 1, 1, ?, ?, NULL)", id, code, "/"+code, "Menu", fixed, fixed).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec("INSERT INTO game_admin_role_menus (id, role_id, menu_id, deleted_at) VALUES (?, ?, ?, NULL)", id, 1, id).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := New(db).EnsureMenuCatalog(); err != nil {
		t.Fatal(err)
	}
	var changed int64
	if err := db.Table("game_admin_menus").Where("updated_at <> created_at").Count(&changed).Error; err != nil {
		t.Fatal(err)
	}
	if changed != 0 {
		t.Fatalf("current catalog should not be rewritten, changed rows=%d", changed)
	}
}

func TestEnsureDefaultMenuIconsOnlySeedsUntouchedBlankIcons(t *testing.T) {
	db := newMenuCatalogTestDB(t)
	created := time.Date(2026, 7, 21, 8, 0, 0, 0, time.UTC)
	edited := created.Add(time.Hour)
	if err := db.Exec("INSERT INTO game_admin_menus (id, code, icon, visible, enabled, created_at, updated_at, deleted_at) VALUES (1, 'system-menus', '', 1, 1, ?, ?, NULL), (2, 'funds', '', 1, 1, ?, ?, NULL)", created, created, created, edited).Error; err != nil {
		t.Fatal(err)
	}

	if err := ensureDefaultMenuIcons(db, time.Date(2026, 7, 21, 9, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Code string
		Icon string
	}
	if err := db.Table("game_admin_menus").Select("code,icon").Order("id").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Icon != "Menu" || rows[1].Icon != "" {
		t.Fatalf("unexpected icon seeding result: %#v", rows)
	}
}

func newMenuCatalogTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE game_admin_menus (id INTEGER PRIMARY KEY, code TEXT, route_path TEXT, icon TEXT, visible INTEGER, enabled INTEGER, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,

		`CREATE TABLE game_admin_roles (id INTEGER PRIMARY KEY, code TEXT, deleted_at DATETIME)`,
		`CREATE TABLE game_admin_role_menus (id INTEGER PRIMARY KEY, role_id INTEGER, menu_id INTEGER, deleted_at DATETIME)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}
