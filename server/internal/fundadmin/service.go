package fundadmin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/model"
	"casino/internal/tenancy"
)

const (
	walletActive = 1
	walletFrozen = 2
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

func normalizeQuery(q ListQuery) ListQuery {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	if q.Values == nil {
		q.Values = map[string]string{}
	}
	for key, value := range q.Values {
		q.Values[key] = strings.TrimSpace(value)
	}
	return q
}

func queryValue(q ListQuery, key string) string { return strings.TrimSpace(q.Values[key]) }

func pageOffset(q ListQuery) int { return (q.Page - 1) * q.PageSize }

func requireReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return ErrReasonRequired
	}
	return nil
}

func walletStatus(value int) string {
	if value == walletFrozen {
		return "FROZEN"
	}
	return "ACTIVE"
}

func accountStatus(value int) string {
	switch value {
	case 2:
		return "DISABLED"
	case 3:
		return "FROZEN"
	case 4:
		return "LOCKED"
	case 5:
		return "CLOSED"
	default:
		return "ACTIVE"
	}
}

func depositStatus(value int) string {
	switch value {
	case 2:
		return "SUCCESS"
	case 3:
		return "FAILED"
	case 4:
		return "CANCELLED"
	case 5:
		return "EXPIRED"
	case 6:
		return "MANUAL_REVIEW"
	default:
		return "PENDING"
	}
}

func withdrawalStatus(value int) string {
	switch value {
	case 2:
		return "SUCCESS"
	case 3:
		return "FAILED"
	case 4:
		return "REJECTED"
	case 5:
		return "CANCELLED"
	case 6:
		return "RISK_REVIEW"
	case 7:
		return "FINANCE_REVIEW"
	case 8:
		return "APPROVED"
	case 9:
		return "PROCESSING"
	default:
		return "PENDING"
	}
}

func transactionStatus(value int) string {
	switch value {
	case 2:
		return "FAILED"
	case 3:
		return "REVERSED"
	default:
		return "COMPLETED"
	}
}

func money(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}

func moneyPtr(value decimal.Decimal) *decimal.Decimal { return &value }

func uniqueNo(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, time.Now().UTC().Format("20060102150405.000000000"))
}

func nullableString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func jsonState(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func (s *Service) audit(tx *gorm.DB, actor Actor, permission, action, targetType string, targetID, userID uint64, before, after any, reason string, amount *decimal.Decimal, currency string) error {
	row := model.GameAdminOperationLogs{
		TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID,
		OperatorID: actor.ID, OperatorUsername: actor.Username, Permission: permission, Action: action,
		TargetType: targetType, TargetID: targetID, TargetUserID: userID,
		BeforeState: jsonState(before), AfterState: jsonState(after), Amount: amount,
		Currency: currency, Reason: strings.TrimSpace(reason), IPAddress: actor.IP,
		UserAgent: actor.UserAgent, CorrelationID: actor.CorrelationID,
	}
	return tx.Create(&row).Error
}

func lockWallet(tx *gorm.DB, walletID uint64, scope tenancy.Scope) (model.GameUserWallets, error) {
	var wallet model.GameUserWallets
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameUserWallets{}).Where("id = ? AND deleted_at IS NULL", walletID)
	if err := tenancy.ApplySite(query, "", scope).Take(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return wallet, ErrWalletNotFound
		}
		return wallet, err
	}
	return wallet, nil
}

func lockUserWallet(tx *gorm.DB, userID uint64, currency string, scope tenancy.Scope) (model.GameUserWallets, error) {
	var wallet model.GameUserWallets
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameUserWallets{}).Where("user_id = ? AND deleted_at IS NULL", userID)
	query = tenancy.ApplySite(query, "", scope)
	if strings.TrimSpace(currency) != "" {
		query = query.Where("currency = ?", strings.TrimSpace(currency))
	}
	if err := query.Order("id ASC").Take(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return wallet, ErrWalletNotFound
		}
		return wallet, err
	}
	return wallet, nil
}

func createLedger(tx *gorm.DB, wallet model.GameUserWallets, txnNo, category, direction, sourceType string, sourceID uint64, orderNo string, amount decimal.Decimal, beforeCash, afterCash, beforeFrozen, afterFrozen decimal.Decimal, actor Actor, reason string, idempotencyKey *string) (uint64, error) {
	now := time.Now()
	tenantID, siteID := wallet.TenantID, wallet.SiteID
	if tenantID == 0 {
		tenantID = actor.Scope.TenantID
	}
	if siteID == 0 {
		siteID = actor.Scope.SiteID
	}
	row := model.GameWalletTransactions{
		TenantID: tenantID, SiteID: siteID,
		TxnNo: txnNo, UserID: wallet.UserID, WalletID: wallet.ID, WalletType: "CASH",
		Direction: direction, Category: category, SourceType: sourceType, SourceID: sourceID,
		IdempotencyKey: idempotencyKey, OrderNo: orderNo, Amount: moneyPtr(amount), Currency: wallet.Currency,
		Status: 1, BalanceBefore: moneyPtr(beforeCash), BalanceAfter: moneyPtr(afterCash),
		FrozenBalanceBefore: moneyPtr(beforeFrozen), FrozenBalanceAfter: moneyPtr(afterFrozen),
		Remark: strings.TrimSpace(reason), OperatorID: actor.ID, OperatorUsername: actor.Username,
		CompletedAt: &now, RawMetadata: "{}",
	}
	if err := tx.Create(&row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}
