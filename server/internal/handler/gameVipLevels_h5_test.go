package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestVipProgressIncludesOnlySettledMemberDepositsInWalletCurrency(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, raw.Close()) })
	require.NoError(t, db.Exec(`CREATE TABLE game_deposit_orders (
		id INTEGER PRIMARY KEY, site_id INTEGER, user_id INTEGER, currency TEXT,
		status INTEGER, paid_amount DECIMAL(18,2), deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO game_deposit_orders VALUES
		(1,1,42,'PKR',2,12.25,NULL), (2,1,42,'PKR',2,30.50,NULL),
		(3,1,42,'PKR',1,1000,NULL), (4,1,42,'PKR',3,1000,NULL),
		(5,1,43,'PKR',2,1000,NULL), (6,1,42,'USD',2,1000,NULL),
		(7,0,42,'PKR',2,1000,NULL), (8,1,42,'PKR',2,1000,'2026-01-01'),
		(9,1,0,'PKR',2,1000,NULL)
	`).Error)
	amount, err := h5VipDepositTotal(db, 42, "pkr")
	require.NoError(t, err)
	require.Equal(t, "42.75", amount.StringFixed(2))
	for _, userID := range []uint64{0, 44} {
		amount, err = h5VipDepositTotal(db, userID, "PKR")
		require.NoError(t, err)
		require.True(t, amount.IsZero())
	}
}
