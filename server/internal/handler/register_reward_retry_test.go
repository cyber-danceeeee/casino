package handler

import (
	"casino/internal/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"strings"
	"testing"
)

func registerRewardRetryDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	raw, _ := db.DB()
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { raw.Close() })
	for _, value := range []any{&model.GameRegisterRewardClaims{}, &model.GameUserWallets{}, &model.GameWalletTransactions{}} {
		stmt := &gorm.Statement{DB: db}
		require.NoError(t, stmt.Parse(value))
		columns := []string{}
		for _, field := range stmt.Schema.Fields {
			if field.DBName == "" {
				continue
			}
			kind := "TEXT"
			switch field.GORMDataType {
			case schema.Int, schema.Uint, schema.Bool:
				kind = "INTEGER"
			case schema.Time:
				kind = "DATETIME"
			}
			if field.DBName == "id" {
				kind = "INTEGER PRIMARY KEY"
			}
			columns = append(columns, "`"+field.DBName+"` "+kind)
		}
		require.NoError(t, db.Exec("CREATE TABLE "+stmt.Schema.Table+" ("+strings.Join(columns, ",")+")").Error)
	}
	return db
}
func TestRegisterRewardRetryReturnsReceiptWithoutAnotherCredit(t *testing.T) {
	db := registerRewardRetryDB(t)
	before := decimal.NewFromInt(50)
	amount := decimal.NewFromInt(7)
	wallet := model.GameUserWallets{SiteID: 1, UserID: 9, Currency: "IDR", CashBalance: &before, Status: 1}
	require.NoError(t, db.Create(&wallet).Error)
	claim := model.GameRegisterRewardClaims{SiteID: 1, UserID: 9, Currency: "IDR", Amount: &amount, Status: 1, IdempotencyKey: "register-reward-test-9"}
	require.NoError(t, db.Create(&claim).Error)
	user := model.GameUsers{SiteID: 1}
	user.ID = 9
	first, receipt, err := claimRegisterRewardFromDB(db, user)
	require.NoError(t, err)
	require.True(t, first.CashBalance.Equal(decimal.NewFromInt(57)))
	require.Equal(t, "57.00", receipt["cashAfter"])
	// Another legitimate wallet change must survive retrying the old claim.
	require.NoError(t, db.Model(&wallet).Update("cash_balance", "67.00").Error)
	retry, again, err := claimRegisterRewardFromDB(db, user)
	require.NoError(t, err)
	require.True(t, retry.CashBalance.Equal(decimal.NewFromInt(67)))
	require.Equal(t, receipt, again)
	var count int64
	require.NoError(t, db.Model(&model.GameWalletTransactions{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	_, _, err = claimRegisterRewardFromDB(db, model.GameUsers{})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
