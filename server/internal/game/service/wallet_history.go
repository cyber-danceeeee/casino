package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"strconv"
	"time"

	"casino/internal/content"
	"casino/internal/game/domain"
	"casino/internal/game/repository"
	"casino/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ExecuteBatch commits every wallet mutation and projection in one transaction.
// The caller cannot combine providers, currencies, or players in a batch.
func (s *WalletCallbackService) ExecuteBatch(ctx context.Context, commands []domain.WalletCommand) (*domain.WalletResult, error) {
	if len(commands) == 0 || len(commands) > 1000 {
		return nil, domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid wallet batch", nil)
	}
	first := commands[0]
	for _, c := range commands {
		if !c.TrackHistory || c.Operation == domain.WalletOperationBalance || c.ProviderCode != first.ProviderCode || c.ProviderPlayerID != first.ProviderPlayerID || c.Currency != first.Currency {
			return nil, domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid batch scope", nil)
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 4500*time.Millisecond)
	defer cancel()
	var result *domain.WalletResult
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		inner := NewWalletCallbackService(repository.New(tx))
		for _, c := range commands {
			var err error
			result, err = inner.Execute(ctx, c)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (s *WalletCallbackService) executeHistory(ctx context.Context, c domain.WalletCommand, userID uint64) (*domain.WalletResult, error) {
	if c.ProviderHistoryID == "" || len(c.ProviderHistoryID) > 32 || c.ProviderRoundID == "" || len(c.ProviderRoundID) > 128 || c.ProviderGameID == "" || len(c.ProviderGameID) > 128 {
		return nil, domain.NewWalletError(domain.WalletErrorInvalidRequest, "missing wallet history identifiers", nil)
	}
	if c.Operation != domain.WalletOperationDebit && c.Operation != domain.WalletOperationCredit && c.Operation != domain.WalletOperationCancel {
		return nil, domain.NewWalletError(domain.WalletErrorInvalidRequest, "unsupported history operation", nil)
	}
	scope := content.Scope{SiteID: c.SiteID}
	var result *domain.WalletResult
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the wallet before duplicate checks: same-player retries and entire
		// batches serialize through this row, including cross-history transactions.
		var wallet model.GameUserWallets
		if err := content.ApplySite(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&wallet), "", scope).Where("user_id=?", userID).Take(&wallet).Error; err != nil {
			return domain.NewWalletError(domain.WalletErrorPlayerNotFound, "wallet not found", nil)
		}
		if ok, err := reconcileWalletCurrency(ctx, tx, scope, &wallet, c.Currency); err != nil {
			return err
		} else if !ok {
			return domain.NewWalletError(domain.WalletErrorCurrencyMismatch, "wallet currency mismatch", nil)
		}
		if wallet.Status != 1 {
			return domain.NewWalletError(domain.WalletErrorInvalidRequest, "wallet is not active", nil)
		}
		receipt := model.ProviderWalletReceipt{ProviderCode: c.ProviderCode, TransactionCode: c.ProviderTransactionID, UserID: userID, ProviderHistoryID: c.ProviderHistoryID}
		insert := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&receipt)
		if insert.Error != nil {
			return insert.Error
		}
		if insert.RowsAffected == 0 {
			return domain.NewWalletError(domain.WalletErrorDuplicateTransaction, "duplicate provider transaction", nil)
		}
		if c.Operation != domain.WalletOperationCancel {
			var closed model.ProviderWalletHistory
			e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("provider_code=? AND user_id=? AND provider_round_id=? AND status IN ?", c.ProviderCode, userID, c.ProviderRoundID, []string{"FINISHED", "CANCELLED"}).Take(&closed).Error
			if e == nil {
				return domain.NewWalletError(domain.WalletErrorRoundClosed, "wallet round is closed", nil)
			}
			if !errors.Is(e, gorm.ErrRecordNotFound) {
				return e
			}
		}
		var h model.ProviderWalletHistory
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("provider_code=? AND user_id=? AND provider_history_id=?", c.ProviderCode, userID, c.ProviderHistoryID).Take(&h).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h = model.ProviderWalletHistory{SiteID: c.SiteID, ProviderCode: c.ProviderCode, UserID: userID, ProviderHistoryID: c.ProviderHistoryID, ProviderRoundID: c.ProviderRoundID, ProviderGameID: c.ProviderGameID, Status: "OPEN"}
			if err = tx.Create(&h).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if h.ProviderRoundID != c.ProviderRoundID || h.ProviderGameID != c.ProviderGameID || h.SiteID != c.SiteID {
			return domain.NewWalletError(domain.WalletErrorInvalidRequest, "history identifiers do not match", nil)
		}
		if h.Status == "CANCELLED" || (h.Status == "FINISHED" && c.Operation != domain.WalletOperationCancel) {
			return domain.NewWalletError(domain.WalletErrorRoundClosed, "wallet history is closed", nil)
		}
		inner := NewWalletCallbackService(repository.New(tx))
		if c.Operation == domain.WalletOperationCancel {
			result, err = inner.cancelHistory(ctx, tx, c, userID, &wallet)
			h.Status = "CANCELLED"
		} else {
			result, err = inner.execute(ctx, c, true)
			if err == nil {
				if c.Operation == domain.WalletOperationDebit {
					h.BetMinor += c.AmountMinor
				} else {
					h.WinMinor += c.AmountMinor
				}
				if h.BetMinor < 0 || h.WinMinor < 0 {
					return domain.NewWalletError(domain.WalletErrorInvalidRequest, "history total overflow", nil)
				}
			}
			if c.RoundFinished {
				h.Status = "FINISHED"
			}
		}
		if err != nil {
			return err
		}
		h.TransactionCount++
		if err = tx.Save(&h).Error; err != nil {
			return err
		}
		return syncWalletHistoryProjection(tx, c, h, wallet.ID)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (s *WalletCallbackService) cancelHistory(ctx context.Context, tx *gorm.DB, c domain.WalletCommand, userID uint64, wallet *model.GameUserWallets) (*domain.WalletResult, error) {
	var originals []model.GameProviderTransactions
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("site_id=? AND provider_code=? AND user_id=? AND provider_history_id=? AND status=? AND operation IN ?", c.SiteID, c.ProviderCode, userID, c.ProviderHistoryID, model.GameProviderTransactionSuccess, []string{"DEBIT", "CREDIT"}).Order("CASE WHEN operation='DEBIT' THEN 0 ELSE 1 END,id ASC").Find(&originals).Error
	if err != nil {
		return nil, err
	}
	before := decimalToMinor(wallet.CashBalance)
	after := before
	for _, original := range originals {
		// Refund successful bets before reverting successful payouts. Never trust
		// the amount supplied on a cancellation; only our committed ledger counts.
		child := c
		child.TrackHistory = false
		child.Operation = domain.WalletOperationRollback
		child.ProviderAction = "history-cancel"
		child.RelatedTransactionID = original.ProviderTransactionID
		child.AmountMinor = 0
		child.RequestHash = ""
		digest := sha256.Sum256([]byte(c.ProviderTransactionID + "\x00" + strconv.FormatUint(original.ID, 10)))
		child.ProviderTransactionID = "cancel:" + hex.EncodeToString(digest[:])
		r, e := s.execute(ctx, child, true)
		if e != nil {
			return nil, e
		}
		after = r.BalanceMinor
	}
	if err = s.createProviderTransaction(tx, c, userID, 0, model.GameProviderTransactionSuccess, before, after, `{"success":true}`); err != nil {
		return nil, err
	}
	return &domain.WalletResult{Success: true, UserID: strconv.FormatUint(userID, 10), ProviderTransactionID: c.ProviderTransactionID, BalanceMinor: after, Currency: c.Currency}, nil
}
func syncWalletHistoryProjection(tx *gorm.DB, c domain.WalletCommand, h model.ProviderWalletHistory, walletID uint64) error {
	now := time.Now()
	scope := content.Scope{SiteID: c.SiteID}
	gameID, providerID, gameType := resolveGameProjection(tx, scope, c.ProviderCode, c.ProviderGameID)
	key := projectionID("BET", c.ProviderCode, "history:"+strconv.FormatUint(h.UserID, 10)+":"+h.ProviderHistoryID)
	bet := h.BetMinor
	win := h.WinMinor
	valid := bet
	profit := bet - win
	status := "BET_ACCEPTED"
	numericStatus := 1
	rollback := "NONE"
	var settled any
	if h.Status == "FINISHED" {
		status = "SETTLED"
		numericStatus = 2
		settled = now
	}
	if h.Status == "CANCELLED" {
		status = "ROLLED_BACK"
		numericStatus = 2
		rollback = "ROLLED_BACK"
		valid = 0
		win = 0
		profit = 0
		settled = now
	}
	var existing model.GameBetOrders
	e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("provider_code=? AND platform_bet_id=? AND user_id=?", c.ProviderCode, key, h.UserID).Take(&existing).Error
	values := map[string]any{"bet_amount_minor": bet, "valid_bet_amount_minor": valid, "win_amount_minor": win, "platform_profit_minor": profit, "bet_amount": minorToDecimal(bet), "valid_bet_amount": minorToDecimal(valid), "payout_amount": minorToDecimal(win), "profit_amount": minorToDecimal(profit), "bet_status": status, "status": numericStatus, "rollback_status": rollback, "settled_at": settled, "updated_at": now}
	if c.Operation == domain.WalletOperationCredit {
		values["provider_win_transaction_id"] = c.ProviderTransactionID
	}
	if errors.Is(e, gorm.ErrRecordNotFound) {
		for k, v := range map[string]any{"site_id": c.SiteID, "bet_no": key, "platform_bet_id": key, "user_id": h.UserID, "wallet_id": walletID, "game_id": gameID, "provider_id": providerID, "provider_code": c.ProviderCode, "provider_game_id": c.ProviderGameID, "provider_round_id": c.ProviderRoundID, "provider_bet_transaction_id": c.ProviderTransactionID, "provider_win_transaction_id": "", "game_type": gameType, "currency": c.Currency, "failure_reason": "", "bet_at": now, "created_at": now} {
			if _, ok := values[k]; !ok {
				values[k] = v
			}
		}
		if e = tx.Table("game_bet_records").Create(values).Error; e != nil {
			return e
		}
	} else if e != nil {
		return e
	} else if e = tx.Table("game_bet_records").Where("id=?", existing.ID).Updates(values).Error; e != nil {
		return e
	}
	var histories []model.ProviderWalletHistory
	if e = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("provider_code=? AND user_id=? AND provider_round_id=?", c.ProviderCode, h.UserID, c.ProviderRoundID).Find(&histories).Error; e != nil {
		return e
	}
	var totalBet, totalWin int64
	var count uint
	open := false
	active := false
	for _, row := range histories {
		count += row.TransactionCount
		if row.Status == "OPEN" {
			open = true
		}
		if row.Status != "CANCELLED" {
			active = true
			if row.BetMinor > math.MaxInt64-totalBet || row.WinMinor > math.MaxInt64-totalWin {
				return domain.NewWalletError(domain.WalletErrorInvalidRequest, "round total overflow", nil)
			}
			totalBet += row.BetMinor
			totalWin += row.WinMinor
		}
	}
	if totalBet < 0 || totalWin < 0 {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "round total overflow", nil)
	}
	roundStatus := "SETTLED"
	var completed any = now
	if open {
		roundStatus = "BET_ACCEPTED"
		completed = nil
	} else if !active {
		roundStatus = "ROLLED_BACK"
	}
	var round model.GameRounds
	e = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("site_id=? AND provider_code=? AND user_id=? AND provider_round_id=?", c.SiteID, c.ProviderCode, h.UserID, c.ProviderRoundID).Take(&round).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		round = model.GameRounds{SiteID: c.SiteID, ProviderCode: c.ProviderCode, ProviderRoundID: c.ProviderRoundID, ProviderGameID: c.ProviderGameID, GameID: gameID, UserID: h.UserID, WalletID: walletID, Currency: c.Currency, StartedAt: &now}
		if e = tx.Create(&round).Error; e != nil {
			return e
		}
	} else if e != nil {
		return e
	}
	return tx.Model(&round).Updates(map[string]any{"total_bet_minor": totalBet, "total_win_minor": totalWin, "platform_profit_minor": totalBet - totalWin, "transaction_count": count, "status": roundStatus, "completed_at": completed}).Error
}
