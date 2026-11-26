// bootstrap-admin sets the seeded platform administrator's password exactly once.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"casino/internal/config"
	"casino/internal/database"
)

func main() {
	filename := flag.String("c", "configs/casino.yml", "configuration file")
	flag.Parse()
	if err := config.Init(*filename); err != nil {
		log.Fatal(err)
	}
	database.InitDB()
	defer database.CloseDB()
	if err := initialize(database.GetDB(), os.Getenv("CASINO_BOOTSTRAP_ADMIN_PASSWORD")); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Platform administrator initialized (existing passwords are preserved).")
}

func initialize(db *gorm.DB, password string) error {
	var account struct {
		ID           uint64
		PasswordHash string
	}
	err := db.Table("game_admin_accounts").Select("id,password_hash").
		Where("username=? AND tenant_id=0 AND deleted_at IS NULL", "system-admin").Take(&account).Error
	if err != nil {
		return fmt.Errorf("find seeded platform administrator: %w", err)
	}
	if account.PasswordHash != "" {
		return nil
	}
	if len(password) < 12 || len(password) > 72 {
		return fmt.Errorf("CASINO_BOOTSTRAP_ADMIN_PASSWORD must contain 12–72 UTF-8 bytes")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// The predicate prevents a concurrent bootstrap from overwriting a password.
	return db.Table("game_admin_accounts").Where("id=? AND password_hash=''", account.ID).
		Updates(map[string]any{"password_hash": string(hash), "must_change_password": 1}).Error
}
