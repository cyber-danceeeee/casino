package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/game/domain"
	"casino/internal/model"
)

func syncBettingProjection(tx *gorm.DB, command domain.WalletCommand, userID uint64, walletID uint64, walletTransactionID uint64, providerStatus string, now time.Time) error {
	if command.TrackHistory {
		return nil
	} // History-aware projection is updated once the history state is saved.
	if providerStatus != model.GameProviderTransactionSuccess && providerStatus != model.GameProviderTransactionRolledBack {
		return nil
	}

	scope := content.Scope{SiteID: command.SiteID}
	gameID, providerID, gameType := resolveGameProjection(tx, scope, command.ProviderCode, command.ProviderGameID)
	if command.ProviderRoundID != "" {
		if err := upsertRoundProjection(tx, command, userID, walletID, gameID, now); err != nil {
			return err
		}
	}

	switch command.Operation {
	case domain.WalletOperationDebit:
		return createBetProjection(tx, command, userID, walletID, gameID, providerID, gameType, now)
	case domain.WalletOperationCredit:
		return settleBetProjection(tx, command, userID, walletID, gameID, providerID, gameType, false, now)
	case domain.WalletOperationRefund:
		return settleBetProjection(tx, command, userID, walletID, gameID, providerID, gameType, true, now)
	case domain.WalletOperationRollback:
		return createRollbackProjection(tx, command, userID, gameID, walletTransactionID, now)
	default:
		return nil
	}
}

func resolveGameProjection(tx *gorm.DB, scope content.Scope, providerCode string, providerGameID string) (uint64, uint64, string) {
	row := struct {
		GameID     uint64 `gorm:"column:game_id"`
		ProviderID uint64 `gorm:"column:provider_id"`
		GameType   string `gorm:"column:game_type"`
	}{}
	db := content.ApplySite(tx.Table("game_games AS g"), "g", scope)
	db = content.ApplySite(db, "p", scope)
	_ = db.Select("g.id AS game_id, g.provider_id, g.game_type").
		Joins("INNER JOIN game_game_providers AS p ON p.id = g.provider_id AND p.deleted_at IS NULL").
		Where("p.provider_code = ? AND g.provider_game_code = ? AND g.deleted_at IS NULL", providerCode, providerGameID).
		Take(&row).Error
	return row.GameID, row.ProviderID, row.GameType
}

func upsertRoundProjection(tx *gorm.DB, command domain.WalletCommand, userID uint64, walletID uint64, gameID uint64, now time.Time) error {
	scope := content.Scope{SiteID: command.SiteID}
	var round model.GameRounds
	err := content.ApplySite(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameRounds{}), "", scope).
		Where("provider_code = ? AND user_id = ? AND provider_round_id = ?", command.ProviderCode, userID, command.ProviderRoundID).
		First(&round).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		round = model.GameRounds{

			SiteID:       scope.SiteID,
			ProviderCode: command.ProviderCode, ProviderRoundID: command.ProviderRoundID,
			ProviderGameID: command.ProviderGameID, GameID: gameID, UserID: userID, WalletID: walletID,
			Currency: command.Currency, Status: "PENDING", StartedAt: &now,
		}
		if command.Operation == domain.WalletOperationDebit {
			round.TotalBetMinor = command.AmountMinor
			round.PlatformProfitMinor = command.AmountMinor
			round.Status = "BET_ACCEPTED"
		}
		if command.Operation == domain.WalletOperationCredit {
			round.TotalWinMinor = command.AmountMinor
			round.PlatformProfitMinor = -command.AmountMinor
			round.Status = "SETTLED"
			round.CompletedAt = &now
		}
		if command.Operation == domain.WalletOperationRefund {
			round.TotalWinMinor = command.AmountMinor
			round.PlatformProfitMinor = -command.AmountMinor
			round.Status = "REFUNDED"
			round.CompletedAt = &now
		}
		if command.Operation == domain.WalletOperationRollback {
			round.Status = "ROLLED_BACK"
			round.CompletedAt = &now
		}
		round.TransactionCount = 1
		return tx.Create(&round).Error
	}
	if err != nil {
		return err
	}

	updates := map[string]any{
		"transaction_count": gorm.Expr("transaction_count + 1"),
		"provider_game_id":  command.ProviderGameID,
		"game_id":           gameID,
	}
	switch command.Operation {
	case domain.WalletOperationDebit:
		updates["total_bet_minor"] = gorm.Expr("total_bet_minor + ?", command.AmountMinor)
		updates["platform_profit_minor"] = gorm.Expr("platform_profit_minor + ?", command.AmountMinor)
		updates["status"] = "BET_ACCEPTED"
	case domain.WalletOperationCredit:
		updates["total_win_minor"] = gorm.Expr("total_win_minor + ?", command.AmountMinor)
		updates["platform_profit_minor"] = gorm.Expr("platform_profit_minor - ?", command.AmountMinor)
		updates["status"] = "SETTLED"
		updates["completed_at"] = now
	case domain.WalletOperationRefund:
		updates["total_win_minor"] = gorm.Expr("total_win_minor + ?", command.AmountMinor)
		updates["platform_profit_minor"] = gorm.Expr("platform_profit_minor - ?", command.AmountMinor)
		updates["status"] = "REFUNDED"
		updates["completed_at"] = now
	case domain.WalletOperationRollback:
		updates["status"] = "ROLLED_BACK"
		updates["completed_at"] = now
	}
	return tx.Model(&round).Updates(updates).Error
}

func createBetProjection(tx *gorm.DB, command domain.WalletCommand, userID uint64, walletID uint64, gameID uint64, providerID uint64, gameType string, now time.Time) error {
	scope := content.Scope{SiteID: command.SiteID}
	platformID := projectionID("BET", command.ProviderCode, command.ProviderTransactionID)
	amount := minorToDecimal(command.AmountMinor)
	row := &model.GameBetOrders{
		BetNo: platformID, PlatformBetID: platformID, UserID: userID, WalletID: walletID,
		GameID: gameID, ProviderID: providerID, ProviderCode: command.ProviderCode,
		ProviderGameID: command.ProviderGameID, ProviderRoundID: command.ProviderRoundID,
		ProviderBetTransactionID: command.ProviderTransactionID, GameType: gameType,
		BetAmountMinor: command.AmountMinor, ValidBetAmountMinor: command.AmountMinor,
		WinAmountMinor: 0, PlatformProfitMinor: command.AmountMinor, Currency: command.Currency,
		BetStatus: "BET_ACCEPTED", RollbackStatus: "NONE", BetAt: &now,
	}
	return tx.Table(row.TableName()).Create(map[string]any{
		"site_id": scope.SiteID,
		"bet_no":  row.BetNo, "platform_bet_id": row.PlatformBetID, "user_id": row.UserID, "wallet_id": row.WalletID,
		"game_id": row.GameID, "provider_id": row.ProviderID, "provider_code": row.ProviderCode,
		"provider_game_id": row.ProviderGameID, "provider_round_id": row.ProviderRoundID,
		"provider_bet_transaction_id": row.ProviderBetTransactionID, "provider_win_transaction_id": "",
		"game_type": row.GameType, "bet_amount_minor": row.BetAmountMinor, "valid_bet_amount_minor": row.ValidBetAmountMinor,
		"win_amount_minor": 0, "platform_profit_minor": row.PlatformProfitMinor, "currency": row.Currency,
		"bet_amount": amount, "valid_bet_amount": amount, "payout_amount": decimal.Zero, "profit_amount": amount,
		"status": 1, "bet_status": row.BetStatus, "rollback_status": row.RollbackStatus, "failure_reason": "",
		"bet_at": now, "created_at": now, "updated_at": now,
	}).Error
}

func settleBetProjection(tx *gorm.DB, command domain.WalletCommand, userID uint64, walletID uint64, gameID uint64, providerID uint64, gameType string, refunded bool, now time.Time) error {
	scope := content.Scope{SiteID: command.SiteID}
	var bet model.GameBetOrders
	query := content.ApplySite(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameBetOrders{}), "", scope).Where("provider_code = ? AND user_id = ?", command.ProviderCode, userID)
	if command.ProviderRoundID != "" {
		query = query.Where("provider_round_id = ?", command.ProviderRoundID)
	}
	if err := query.Order("id DESC").First(&bet).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		status := "SETTLED"
		if refunded {
			status = "REFUNDED"
		}
		platformID := projectionID("BET", command.ProviderCode, "credit-only:"+command.ProviderTransactionID)
		return tx.Table("game_bet_records").Create(map[string]any{
			"site_id": scope.SiteID,
			"bet_no":  platformID, "platform_bet_id": platformID, "user_id": userID, "wallet_id": walletID,
			"game_id": gameID, "provider_id": providerID, "provider_code": command.ProviderCode,
			"provider_game_id": command.ProviderGameID, "provider_round_id": command.ProviderRoundID,
			"provider_bet_transaction_id": "CREDIT_ONLY:" + command.ProviderTransactionID,
			"provider_win_transaction_id": command.ProviderTransactionID, "game_type": gameType,
			"bet_amount_minor": 0, "valid_bet_amount_minor": 0, "win_amount_minor": command.AmountMinor,
			"platform_profit_minor": -command.AmountMinor, "currency": command.Currency,
			"bet_amount": decimal.Zero, "valid_bet_amount": decimal.Zero, "payout_amount": minorToDecimal(command.AmountMinor),
			"profit_amount": minorToDecimal(-command.AmountMinor), "status": 2, "bet_status": status,
			"rollback_status": "NONE", "failure_reason": "", "bet_at": now, "settled_at": now,
			"created_at": now, "updated_at": now,
		}).Error
	}
	win := bet.WinAmountMinor + command.AmountMinor
	profit := bet.BetAmountMinor - win
	status := "SETTLED"
	if refunded {
		status = "REFUNDED"
	}
	return tx.Table(bet.TableName()).Where("id = ?", bet.ID).Updates(map[string]any{
		"provider_win_transaction_id": command.ProviderTransactionID,
		"win_amount_minor":            win, "platform_profit_minor": profit,
		"payout_amount": minorToDecimal(win), "profit_amount": minorToDecimal(profit),
		"bet_status": status, "status": 2, "settled_at": now,
	}).Error
}

func createRollbackProjection(tx *gorm.DB, command domain.WalletCommand, userID uint64, gameID uint64, walletTransactionID uint64, now time.Time) error {
	scope := content.Scope{SiteID: command.SiteID}
	var original model.GameProviderTransactions
	if err := content.ApplySite(tx.Model(&model.GameProviderTransactions{}), "", scope).Where("provider_code = ? AND provider_transaction_id = ?", command.ProviderCode, command.RelatedTransactionID).First(&original).Error; err != nil {
		return err
	}
	var bet model.GameBetOrders
	_ = content.ApplySite(tx.Model(&model.GameBetOrders{}), "", scope).Where("provider_code = ? AND user_id = ? AND (provider_bet_transaction_id = ? OR provider_win_transaction_id = ?)", command.ProviderCode, userID, command.RelatedTransactionID, command.RelatedTransactionID).First(&bet).Error
	rollbackID := projectionID("RB", command.ProviderCode, command.ProviderTransactionID)
	row := &model.GameRollbacks{

		SiteID:     scope.SiteID,
		RollbackID: rollbackID, ProviderCode: command.ProviderCode, ProviderTransactionID: command.ProviderTransactionID,
		ProviderRoundID: command.ProviderRoundID, OriginalProviderTransactionID: command.RelatedTransactionID,
		OriginalWalletTransactionID: original.WalletTransactionID, RollbackWalletTransactionID: walletTransactionID,
		BetRecordID: bet.ID, UserID: userID, GameID: gameID, ProviderGameID: command.ProviderGameID,
		OriginalOperation: original.Operation, OriginalAmountMinor: original.AmountMinor,
		RollbackAmountMinor: original.AmountMinor, Currency: command.Currency, Status: "COMPLETED", CompletedAt: &now,
	}
	if err := tx.Create(row).Error; err != nil {
		return err
	}
	if bet.ID > 0 {
		if err := tx.Table(bet.TableName()).Where("id = ?", bet.ID).Updates(map[string]any{"bet_status": "ROLLED_BACK", "rollback_status": "ROLLED_BACK"}).Error; err != nil {
			return err
		}
	}
	return nil
}

func projectionID(prefix string, provider string, transactionID string) string {
	sum := sha256.Sum256([]byte(provider + ":" + transactionID))
	return fmt.Sprintf("%s%s", prefix, hex.EncodeToString(sum[:12]))
}
