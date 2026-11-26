package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/game/domain"
	"casino/internal/game/repository"
	"casino/internal/model"
	"casino/internal/tenancy"
)

type WalletCallbackService struct {
	repo *repository.Repository
}

func NewWalletCallbackService(repo *repository.Repository) *WalletCallbackService {
	return &WalletCallbackService{repo: repo}
}

func (s *WalletCallbackService) Execute(ctx context.Context, command domain.WalletCommand) (*domain.WalletResult, error) {
	command.ProviderCode = strings.ToLower(strings.TrimSpace(command.ProviderCode))
	command.Currency = strings.ToUpper(strings.TrimSpace(command.Currency))
	if command.ProviderCode == "" || command.ProviderPlayerID == "" {
		return nil, domain.NewWalletError(domain.WalletErrorInvalidRequest, "missing provider or player", nil)
	}

	player, err := s.repo.PlayerByProviderID(ctx, command.ProviderCode, command.ProviderPlayerID)
	if err != nil {
		return nil, domain.NewWalletError(domain.WalletErrorPlayerNotFound, "provider player not found", err)
	}
	scope := tenancy.Scope{TenantID: player.TenantID, SiteID: player.SiteID}
	command.TenantID = scope.TenantID
	command.SiteID = scope.SiteID
	if command.Currency == "" {
		command.Currency = player.Currency
	}
	if !strings.EqualFold(command.Currency, player.Currency) {
		return nil, domain.NewWalletError(
			domain.WalletErrorCurrencyMismatch,
			fmt.Sprintf("provider player currency mismatch: callback=%s mapping=%s", command.Currency, player.Currency),
			nil,
		)
	}
	if command.Operation == domain.WalletOperationBalance {
		balance, err := s.balance(ctx, scope, player.UserID, command.Currency)
		if err != nil {
			return nil, err
		}
		return &domain.WalletResult{
			Success:               true,
			UserID:                strconv.FormatUint(player.UserID, 10),
			ProviderTransactionID: command.ProviderTransactionID,
			BalanceMinor:          balance,
			Currency:              command.Currency,
		}, nil
	}
	if err := validateWalletCommand(command); err != nil {
		return nil, err
	}
	if command.RequestHash == "" {
		command.RequestHash = commandHash(command)
	}

	var result *domain.WalletResult
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.GameProviderTransactions
		err := tenancy.ApplySite(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameProviderTransactions{}), "", scope).
			Where("provider_code = ? AND provider_transaction_id = ? AND operation = ?",
				command.ProviderCode, command.ProviderTransactionID, command.Operation).
			First(&existing).Error
		if err == nil {
			if existing.RequestHash == command.RequestHash {
				result = &domain.WalletResult{
					Success:               existing.Status == model.GameProviderTransactionSuccess || existing.Status == model.GameProviderTransactionRolledBack,
					UserID:                strconv.FormatUint(existing.UserID, 10),
					InternalTransactionID: strconv.FormatUint(existing.WalletTransactionID, 10),
					ProviderTransactionID: existing.ProviderTransactionID,
					BalanceMinor:          existing.BalanceAfterMinor,
					Currency:              existing.Currency,
					Duplicate:             true,
				}
				if existing.Status == model.GameProviderTransactionRejected {
					return &domain.WalletError{
						Code:         domain.WalletErrorInsufficientBalance,
						Message:      "insufficient balance",
						BalanceMinor: existing.BalanceAfterMinor,
						Currency:     existing.Currency,
					}
				}
				return nil
			}
			return domain.NewWalletError(domain.WalletErrorTransactionConflict, "conflicting duplicate provider transaction", domain.ErrTransactionConflict)
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var wallet model.GameUserWallets
		if err := tenancy.ApplySite(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameUserWallets{}), "", scope).Where("user_id = ?", player.UserID).First(&wallet).Error; err != nil {
			return domain.NewWalletError(domain.WalletErrorPlayerNotFound, "wallet not found", err)
		}
		if ok, err := reconcileWalletCurrency(ctx, tx, scope, &wallet, command.Currency); err != nil {
			return err
		} else if !ok {
			return domain.NewWalletError(domain.WalletErrorCurrencyMismatch, "wallet currency mismatch", nil)
		}
		if wallet.Status != 1 {
			return domain.NewWalletError(domain.WalletErrorInvalidRequest, "wallet is not active", nil)
		}

		before := decimalToMinor(wallet.CashBalance)
		amount := command.AmountMinor
		direction := ""
		category := ""
		after := before
		status := model.GameProviderTransactionSuccess

		switch command.Operation {
		case domain.WalletOperationDebit:
			direction = "debit"
			category = "bet"
			if before < amount {
				status = model.GameProviderTransactionRejected
				after = before
				if err := s.createProviderTransaction(tx, command, player.UserID, 0, status, before, after, "{}"); err != nil {
					return err
				}
				return &domain.WalletError{
					Code:         domain.WalletErrorInsufficientBalance,
					Message:      "insufficient balance",
					BalanceMinor: before,
					Currency:     command.Currency,
				}
			}
			after = before - amount
		case domain.WalletOperationCredit:
			direction = "credit"
			category = "payout"
			after = before + amount
		case domain.WalletOperationRefund:
			direction = "credit"
			category = "refund"
			after = before + amount
		case domain.WalletOperationRollback:
			var original model.GameProviderTransactions
			if err := tenancy.ApplySite(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameProviderTransactions{}), "", scope).
				Where("provider_code = ? AND provider_transaction_id = ? AND status = ?",
					command.ProviderCode, command.RelatedTransactionID, model.GameProviderTransactionSuccess).
				First(&original).Error; err != nil {
				return domain.NewWalletError(domain.WalletErrorInvalidRequest, "original transaction not found", err)
			}
			amount = original.AmountMinor
			category = "rollback"
			if original.Operation == string(domain.WalletOperationDebit) {
				direction = "credit"
				after = before + amount
			} else {
				direction = "debit"
				if before < amount {
					return domain.NewWalletError(domain.WalletErrorInsufficientBalance, "insufficient balance for rollback", nil)
				}
				after = before - amount
			}
			if err := tx.Model(&original).Update("status", model.GameProviderTransactionRolledBack).Error; err != nil {
				return err
			}
		default:
			return domain.NewWalletError(domain.WalletErrorInvalidRequest, "unsupported wallet operation", nil)
		}

		afterDecimal := minorToDecimal(after)
		now := time.Now()
		if err := tx.Model(&wallet).Updates(map[string]any{"cash_balance": afterDecimal, "last_transaction_at": now}).Error; err != nil {
			return err
		}

		idempotencyKey := "provider:" + command.ProviderCode + ":" + command.ProviderTransactionID + ":" + string(command.Operation)
		ledger := &model.GameWalletTransactions{
			TenantID:            scope.TenantID,
			SiteID:              scope.SiteID,
			TxnNo:               walletTxnNo(),
			UserID:              player.UserID,
			WalletID:            wallet.ID,
			WalletType:          "cash",
			Direction:           direction,
			Category:            category,
			SourceType:          "game_provider_transaction",
			ProviderCode:        command.ProviderCode,
			IdempotencyKey:      &idempotencyKey,
			OrderNo:             command.ProviderTransactionID,
			Amount:              decimalPtr(minorToDecimal(amount)),
			Currency:            command.Currency,
			Status:              1,
			BalanceBefore:       decimalPtr(minorToDecimal(before)),
			BalanceAfter:        decimalPtr(afterDecimal),
			FrozenBalanceBefore: decimalPtr(decimal.Zero),
			FrozenBalanceAfter:  decimalPtr(decimal.Zero),
			Remark:              "provider game wallet callback",
			CompletedAt:         &now,
			RawMetadata:         "{}",
		}
		if err := tx.Create(ledger).Error; err != nil {
			return err
		}

		rawResponse, _ := json.Marshal(map[string]any{"success": true, "balance_minor": after})
		if err := s.createProviderTransaction(tx, command, player.UserID, ledger.ID, status, before, after, string(rawResponse)); err != nil {
			return err
		}
		if err := syncBettingProjection(tx, command, player.UserID, wallet.ID, ledger.ID, status, now); err != nil {
			return err
		}
		result = &domain.WalletResult{
			Success:               true,
			UserID:                strconv.FormatUint(player.UserID, 10),
			InternalTransactionID: strconv.FormatUint(ledger.ID, 10),
			ProviderTransactionID: command.ProviderTransactionID,
			BalanceMinor:          after,
			Currency:              command.Currency,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *WalletCallbackService) balance(ctx context.Context, scope tenancy.Scope, userID uint64, currency string) (int64, error) {
	var wallet model.GameUserWallets
	if err := tenancy.ApplySite(s.repo.DB().WithContext(ctx).Model(&model.GameUserWallets{}), "", scope).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return 0, domain.NewWalletError(domain.WalletErrorPlayerNotFound, "wallet not found", err)
	}
	if ok, err := reconcileWalletCurrency(ctx, s.repo.DB(), scope, &wallet, currency); err != nil {
		return 0, err
	} else if !ok {
		return 0, domain.NewWalletError(domain.WalletErrorCurrencyMismatch, "wallet currency mismatch", nil)
	}
	return decimalToMinor(wallet.CashBalance), nil
}

func (s *WalletCallbackService) createProviderTransaction(tx *gorm.DB, command domain.WalletCommand, userID uint64, ledgerID uint64, status string, before int64, after int64, rawResponse string) error {
	rawRequest := string(command.RawPayload)
	if strings.TrimSpace(rawRequest) == "" {
		rawRequest = "{}"
	}
	row := &model.GameProviderTransactions{
		TenantID:              command.TenantID,
		SiteID:                command.SiteID,
		ProviderCode:          command.ProviderCode,
		ProviderTransactionID: command.ProviderTransactionID,
		ProviderRoundID:       command.ProviderRoundID,
		ProviderGameID:        command.ProviderGameID,
		UserID:                userID,
		WalletTransactionID:   ledgerID,
		Operation:             string(command.Operation),
		AmountMinor:           command.AmountMinor,
		Currency:              command.Currency,
		Status:                status,
		RelatedTransactionID:  command.RelatedTransactionID,
		RequestHash:           command.RequestHash,
		RawRequest:            rawRequest,
		RawResponse:           rawResponse,
		BalanceBeforeMinor:    before,
		BalanceAfterMinor:     after,
	}
	if row.Operation == string(domain.WalletOperationRollback) && command.AmountMinor == 0 {
		row.AmountMinor = before - after
		if row.AmountMinor < 0 {
			row.AmountMinor = -row.AmountMinor
		}
	}
	return tx.Create(row).Error
}

func validateWalletCommand(command domain.WalletCommand) error {
	if command.ProviderTransactionID == "" || len(command.ProviderTransactionID) > 128 {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid provider transaction id", nil)
	}
	if command.Currency == "" || len(command.Currency) != 3 {
		return domain.NewWalletError(domain.WalletErrorCurrencyMismatch, "invalid currency", nil)
	}
	if command.AmountMinor < 0 {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid amount", nil)
	}
	if command.Operation == domain.WalletOperationRollback && command.RelatedTransactionID == "" {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "missing rollback related transaction id", nil)
	}
	return nil
}

func commandHash(command domain.WalletCommand) string {
	semantic := map[string]any{
		"provider_code":           command.ProviderCode,
		"operation":               command.Operation,
		"provider_action":         command.ProviderAction,
		"provider_player_id":      command.ProviderPlayerID,
		"provider_transaction_id": command.ProviderTransactionID,
		"provider_round_id":       command.ProviderRoundID,
		"provider_game_id":        command.ProviderGameID,
		"amount_minor":            command.AmountMinor,
		"currency":                command.Currency,
		"related_transaction_id":  command.RelatedTransactionID,
	}
	body, _ := json.Marshal(semantic)
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func decimalToMinor(value *decimal.Decimal) int64 {
	if value == nil {
		return 0
	}
	return value.Mul(decimal.NewFromInt(100)).IntPart()
}

func minorToDecimal(value int64) decimal.Decimal {
	return decimal.NewFromInt(value).Div(decimal.NewFromInt(100))
}

func decimalPtr(value decimal.Decimal) *decimal.Decimal {
	v := value
	return &v
}

func walletTxnNo() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err == nil {
		return "GWT" + hex.EncodeToString(buf)
	}
	return fmt.Sprintf("GWT%d", time.Now().UnixNano())
}
