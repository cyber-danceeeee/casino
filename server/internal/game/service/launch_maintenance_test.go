package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"casino/internal/content"
	"casino/internal/game/domain"
	"casino/internal/game/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLaunchMaintenanceHonorsWindowAndGlobalOverride(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE game_provider_maintenance (id INTEGER PRIMARY KEY, site_id INTEGER, provider_code TEXT, enabled INTEGER, start_at DATETIME, end_at DATETIME, message_zh TEXT, message_en TEXT, affected_countries TEXT, affected_currencies TEXT, block_new_launches INTEGER, deleted_at DATETIME)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE game_game_maintenance (id INTEGER PRIMARY KEY, site_id INTEGER, game_id INTEGER, enabled INTEGER, start_at DATETIME, end_at DATETIME, message_zh TEXT, message_en TEXT, affected_countries TEXT, affected_currencies TEXT, block_new_launches INTEGER, deleted_at DATETIME)").Error; err != nil {
		t.Fatal(err)
	}
	nowValue := time.Now()
	if err := db.Exec("INSERT INTO game_provider_maintenance (id, site_id, provider_code, enabled, start_at, end_at, message_en, affected_countries, affected_currencies, block_new_launches) VALUES (1, 0, 'PG', 1, ?, ?, 'expired', '', '', 1)", nowValue.Add(-2*time.Hour), nowValue.Add(-time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	service := NewLaunchService(nil, repository.New(db))
	code, _, blocked := service.launchMaintenance(context.Background(), content.Scope{SiteID: 70}, "PG", 10, "MX", "MXN")
	if blocked || code != "" {
		t.Fatalf("expired maintenance still blocked launch: %s", code)
	}
	if err := db.Exec("INSERT INTO game_provider_maintenance (id, site_id, provider_code, enabled, start_at, end_at, message_en, affected_countries, affected_currencies, block_new_launches) VALUES (2, 0, 'PG', 1, ?, ?, 'global maintenance', 'MX', 'MXN', 1)", nowValue.Add(-time.Minute), nowValue.Add(time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	code, message, blocked := service.launchMaintenance(context.Background(), content.Scope{SiteID: 70}, "PG", 10, "MX", "MXN")
	if !blocked || code != domain.WalletErrorProviderMaintenance || message != "global maintenance" {
		t.Fatalf("global maintenance did not override availability: blocked=%v code=%s message=%q", blocked, code, message)
	}
}
