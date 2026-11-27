package service

import (
	"context"
	"strings"
	"testing"

	"casino/internal/game/domain"
	"casino/internal/game/repository"
	"casino/internal/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func historyTestDB(t *testing.T) (*gorm.DB, *WalletCallbackService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	for _, value := range []any{&model.GameProviderPlayers{}, &model.GameUserWallets{}, &model.GameProviderTransactions{}, &model.GameWalletTransactions{}, &model.GameGameProviders{}, &model.GameGames{}, &model.GameBetOrders{}, &model.GameRounds{}, &model.GameRollbacks{}} {
		stmt := &gorm.Statement{DB: db}
		require.NoError(t, stmt.Parse(value))
		cols := []string{}
		for _, f := range stmt.Schema.Fields {
			if f.DBName == "" {
				continue
			}
			kind := "TEXT"
			switch f.GORMDataType {
			case schema.Int, schema.Uint, schema.Bool:
				kind = "INTEGER"
			case schema.Float:
				kind = "NUMERIC"
			case schema.Time:
				kind = "DATETIME"
			}
			if f.DBName == "id" {
				kind = "INTEGER PRIMARY KEY"
			}
			cols = append(cols, "`"+f.DBName+"` "+kind)
		}
		require.NoError(t, db.Exec("CREATE TABLE `"+stmt.Schema.Table+"` ("+strings.Join(cols, ",")+")").Error)
	}
	for _, col := range []string{"bet_amount NUMERIC", "valid_bet_amount NUMERIC", "payout_amount NUMERIC", "profit_amount NUMERIC", "status INTEGER"} {
		require.NoError(t, db.Exec("ALTER TABLE game_bet_records ADD COLUMN "+col).Error)
	}
	require.NoError(t, db.AutoMigrate(&model.ProviderWalletHistory{}, &model.ProviderWalletReceipt{}))
	balance := decimal.NewFromInt(1000)
	require.NoError(t, db.Create(&model.GameUserWallets{SiteID: 1, UserID: 1, Currency: "IDR", CashBalance: &balance, Status: 1}).Error)
	require.NoError(t, db.Create(&model.GameProviderPlayers{SiteID: 1, ProviderCode: "allintop", UserID: 1, ProviderPlayerID: "player", Currency: "IDR", Status: 1}).Error)
	return db, NewWalletCallbackService(repository.New(db))
}
func historyCommand(id, history string, op domain.WalletOperation, amount int64, finished bool) domain.WalletCommand {
	return domain.WalletCommand{ProviderCode: "allintop", ProviderPlayerID: "player", ProviderTransactionID: id, ProviderHistoryID: history, ProviderRoundID: "r" + history, ProviderGameID: "vendor/game", Currency: "IDR", Operation: op, AmountMinor: amount, RoundFinished: finished, TrackHistory: true}
}
func historyBalance(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var wallet model.GameUserWallets
	require.NoError(t, db.Take(&wallet).Error)
	return decimalToMinor(wallet.CashBalance)
}
func TestHistoryDuplicateClosedAndCancelAfterSettlement(t *testing.T) {
	db, s := historyTestDB(t)
	ctx := context.Background()
	_, err := s.Execute(ctx, historyCommand("bet", "1", domain.WalletOperationDebit, 1200, false))
	require.NoError(t, err)
	duplicate := historyCommand("bet", "2", domain.WalletOperationCredit, 1200, true)
	_, err = s.Execute(ctx, duplicate)
	var we *domain.WalletError
	require.ErrorAs(t, err, &we)
	require.Equal(t, domain.WalletErrorDuplicateTransaction, we.Code)
	require.EqualValues(t, 98800, historyBalance(t, db))
	_, err = s.Execute(ctx, historyCommand("win", "1", domain.WalletOperationCredit, 2200, true))
	require.NoError(t, err)
	require.EqualValues(t, 101000, historyBalance(t, db))
	_, err = s.Execute(ctx, historyCommand("late", "1", domain.WalletOperationDebit, 100, false))
	require.ErrorAs(t, err, &we)
	require.Equal(t, domain.WalletErrorRoundClosed, we.Code)
	_, err = s.Execute(ctx, historyCommand("cancel", "1", domain.WalletOperationCancel, 999999, true))
	require.NoError(t, err)
	require.EqualValues(t, 100000, historyBalance(t, db))
	var originals []model.GameProviderTransactions
	require.NoError(t, db.Where("provider_transaction_id IN ?", []string{"bet", "win"}).Find(&originals).Error)
	require.Len(t, originals, 2)
	for _, r := range originals {
		require.Equal(t, model.GameProviderTransactionRolledBack, r.Status)
	}
	var bet model.GameBetOrders
	require.NoError(t, db.Take(&bet).Error)
	require.Equal(t, "ROLLED_BACK", bet.BetStatus)
	require.Zero(t, bet.ValidBetAmountMinor)
	require.Zero(t, bet.PlatformProfitMinor)
	var round model.GameRounds
	require.NoError(t, db.Take(&round).Error)
	require.Equal(t, "ROLLED_BACK", round.Status)
	require.Zero(t, round.TotalBetMinor)
	require.Zero(t, round.TotalWinMinor)
	_, err = s.Execute(ctx, historyCommand("cancel-again", "1", domain.WalletOperationCancel, 0, true))
	require.Error(t, err)
	require.EqualValues(t, 100000, historyBalance(t, db))
}
func TestCancellationBeforeBetDoesNotMintBalanceAndBlocksLateBet(t *testing.T) {
	db, s := historyTestDB(t)
	ctx := context.Background()
	_, err := s.Execute(ctx, historyCommand("cancel", "3", domain.WalletOperationCancel, 100000000, true))
	require.NoError(t, err)
	require.EqualValues(t, 100000, historyBalance(t, db))
	_, err = s.Execute(ctx, historyCommand("latebet", "3", domain.WalletOperationDebit, 1000, true))
	var we *domain.WalletError
	require.ErrorAs(t, err, &we)
	require.Equal(t, domain.WalletErrorRoundClosed, we.Code)
	require.EqualValues(t, 100000, historyBalance(t, db))
}
func TestBatchFailureRollsBackWalletLedgersHistoryAndReceipts(t *testing.T) {
	db, s := historyTestDB(t)
	ctx := context.Background()
	commands := []domain.WalletCommand{historyCommand("bet1", "1", domain.WalletOperationDebit, 1000, true), historyCommand("bet2", "2", domain.WalletOperationDebit, 100000, true)}
	_, err := s.ExecuteBatch(ctx, commands)
	require.Error(t, err)
	require.EqualValues(t, 100000, historyBalance(t, db))
	for _, table := range []string{"game_provider_transactions", "game_wallet_transactions", "game_provider_wallet_histories", "game_provider_wallet_receipts", "game_bet_records", "game_rounds"} {
		var count int64
		require.NoError(t, db.Table(table).Count(&count).Error)
		require.Zero(t, count, table)
	}
	commands[1].AmountMinor = 1000
	result, err := s.ExecuteBatch(ctx, commands)
	require.NoError(t, err)
	require.EqualValues(t, 98000, result.BalanceMinor)
}
func TestFinishedLossAndPartialWinProjection(t *testing.T) {
	db, s := historyTestDB(t)
	ctx := context.Background()
	_, err := s.Execute(ctx, historyCommand("bet", "1", domain.WalletOperationDebit, 1000, false))
	require.NoError(t, err)
	_, err = s.Execute(ctx, historyCommand("partial", "1", domain.WalletOperationCredit, 100, false))
	require.NoError(t, err)
	var bet model.GameBetOrders
	require.NoError(t, db.Take(&bet).Error)
	require.Equal(t, "BET_ACCEPTED", bet.BetStatus)
	_, err = s.Execute(ctx, historyCommand("finish", "1", domain.WalletOperationCredit, 0, true))
	require.NoError(t, err)
	require.NoError(t, db.Take(&bet).Error)
	require.Equal(t, "SETTLED", bet.BetStatus)
	require.EqualValues(t, 100, bet.WinAmountMinor)
	_, err = s.Execute(ctx, historyCommand("loss", "2", domain.WalletOperationDebit, 500, true))
	require.NoError(t, err)
	var loss model.GameBetOrders
	require.NoError(t, db.Where("provider_round_id=?", "r2").Take(&loss).Error)
	require.Equal(t, "SETTLED", loss.BetStatus)
	require.Zero(t, loss.WinAmountMinor)
}
func TestBatchDuplicateRollsBackEarlierCommands(t *testing.T) {
	db, s := historyTestDB(t)
	ctx := context.Background()
	_, err := s.Execute(ctx, historyCommand("existing", "8", domain.WalletOperationDebit, 100, true))
	require.NoError(t, err)
	_, err = s.ExecuteBatch(ctx, []domain.WalletCommand{historyCommand("new", "9", domain.WalletOperationDebit, 200, true), historyCommand("existing", "10", domain.WalletOperationCredit, 10, true)})
	require.Error(t, err)
	require.EqualValues(t, 99900, historyBalance(t, db))
	var count int64
	require.NoError(t, db.Model(&model.ProviderWalletReceipt{}).Where("transaction_code=?", "new").Count(&count).Error)
	require.Zero(t, count)
}

func TestFinishedRoundRejectsAnotherHistory(t *testing.T) {
	db, s := historyTestDB(t)
	ctx := context.Background()
	first := historyCommand("first", "1", domain.WalletOperationDebit, 1000, true)
	_, err := s.Execute(ctx, first)
	require.NoError(t, err)
	other := historyCommand("other", "2", domain.WalletOperationCredit, 1000, false)
	other.ProviderRoundID = first.ProviderRoundID
	_, err = s.Execute(ctx, other)
	var we *domain.WalletError
	require.ErrorAs(t, err, &we)
	require.Equal(t, domain.WalletErrorRoundClosed, we.Code)
	require.EqualValues(t, 99000, historyBalance(t, db))
}

func TestCancellationInsufficientFundsRollsBackEveryReversal(t *testing.T) {
	db, s := historyTestDB(t)
	ctx := context.Background()
	for _, c := range []domain.WalletCommand{historyCommand("bet", "1", domain.WalletOperationDebit, 10000, false), historyCommand("win", "1", domain.WalletOperationCredit, 30000, true), historyCommand("spent", "2", domain.WalletOperationDebit, 119000, true)} {
		_, err := s.Execute(ctx, c)
		require.NoError(t, err)
	}
	require.EqualValues(t, 1000, historyBalance(t, db))
	_, err := s.Execute(ctx, historyCommand("cancel", "1", domain.WalletOperationCancel, 0, true))
	var we *domain.WalletError
	require.ErrorAs(t, err, &we)
	require.Equal(t, domain.WalletErrorInsufficientBalance, we.Code)
	require.EqualValues(t, 1000, historyBalance(t, db))
	var count int64
	require.NoError(t, db.Model(&model.GameProviderTransactions{}).Where("status=?", model.GameProviderTransactionRolledBack).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, db.Model(&model.ProviderWalletReceipt{}).Where("transaction_code=?", "cancel").Count(&count).Error)
	require.Zero(t, count)
}
