package main

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInitializeOnlySetsAnEmptyPassword(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE game_admin_accounts (id INTEGER PRIMARY KEY, username TEXT, deleted_at DATETIME, password_hash TEXT, must_change_password INTEGER)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO game_admin_accounts VALUES (1, 'system-admin', NULL, '', 0)").Error; err != nil {
		t.Fatal(err)
	}
	if err := initialize(db, "short"); err == nil {
		t.Fatal("accepted a short password")
	}
	if err := initialize(db, "a-long-initial-password"); err != nil {
		t.Fatal(err)
	}
	var row struct {
		PasswordHash       string
		MustChangePassword int
	}
	db.Table("game_admin_accounts").Take(&row)
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte("a-long-initial-password")); err != nil {
		t.Fatal(err)
	}
	if row.MustChangePassword != 1 {
		t.Fatal("password change was not required")
	}
	if err := initialize(db, "a-different-long-password"); err != nil {
		t.Fatal(err)
	}
	var actual string
	db.Table("game_admin_accounts").Select("password_hash").Scan(&actual)
	if actual != row.PasswordHash {
		t.Fatal("existing password was overwritten")
	}
}
