package systemadmin

import (
	"context"
	"strings"

	"casino/internal/tenancy"
	"casino/internal/tenantledger"

	"github.com/shopspring/decimal"
)

type TenantLedgerAmountInput struct {
	SiteID         uint64 `json:"siteID"`
	Currency       string `json:"currency"`
	Amount         string `json:"amount"`
	Reason         string `json:"reason"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type TenantCreditLimitInput struct {
	SiteID         uint64 `json:"siteID"`
	Currency       string `json:"currency"`
	CreditLimit    string `json:"creditLimit"`
	Reason         string `json:"reason"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type TenantLedgerReverseInput struct {
	Reason         string `json:"reason"`
	IdempotencyKey string `json:"idempotencyKey"`
}

func (s *Service) TenantLedgerAccounts(tenantID uint64, q Query, actor Actor) (any, error) {
	if err := s.ensureTenantLedgerVisible(tenantID, actor); err != nil {
		return nil, err
	}
	siteID := parseID(value(q, "siteID"))
	if siteID > 0 {
		var count int64
		if err := s.db.Table("game_saas_sites").Where("id=? AND tenant_id=? AND deleted_at IS NULL", siteID, tenantID).Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, rowNotFound("site")
		}
	}
	return tenantledger.New(s.db).ListAccounts(context.Background(), ledgerScope(actor), tenantID, siteID, value(q, "currency"))
}

func (s *Service) TenantLedgerTransaction(tenantID, transactionID uint64, actor Actor) (any, error) {
	if err := s.ensureTenantLedgerVisible(tenantID, actor); err != nil {
		return nil, err
	}
	result, err := tenantledger.New(s.db).GetTransaction(context.Background(), transactionID, ledgerScope(actor))
	if err != nil {
		return nil, ledgerError(err)
	}
	if result.TenantID != tenantID {
		return nil, rowNotFound("ledger transaction")
	}
	return result, nil
}

func (s *Service) CreditTenantDeposit(tenantID uint64, input TenantLedgerAmountInput, actor Actor) (any, error) {
	if err := s.ensurePlatformLedgerOperator(actor); err != nil {
		return nil, err
	}
	if err := s.ensureTenantSiteAvailable(tenantID, input.SiteID); err != nil {
		return nil, err
	}
	amount, err := parseLedgerAmount(input.Amount)
	if err != nil {
		return nil, err
	}
	result, err := tenantledger.New(s.db).CreditTenantDeposit(context.Background(), tenantledger.TenantMoneyCommand{
		Scope:          ledgerScope(actor),
		TenantID:       tenantID,
		SiteID:         input.SiteID,
		Currency:       input.Currency,
		Amount:         amount,
		IdempotencyKey: input.IdempotencyKey,
		Reason:         input.Reason,
		Actor:          ledgerActor(actor),
	})
	return result, ledgerError(err)
}

func (s *Service) SetTenantCreditLimit(tenantID uint64, input TenantCreditLimitInput, actor Actor) (any, error) {
	if err := s.ensurePlatformLedgerOperator(actor); err != nil {
		return nil, err
	}
	if err := s.ensureTenantSiteAvailable(tenantID, input.SiteID); err != nil {
		return nil, err
	}
	limit, err := parseLedgerAmount(input.CreditLimit)
	if err != nil {
		return nil, err
	}
	result, err := tenantledger.New(s.db).SetCreditLimit(context.Background(), tenantledger.CreditLimitCommand{
		Scope:          ledgerScope(actor),
		TenantID:       tenantID,
		SiteID:         input.SiteID,
		Currency:       input.Currency,
		CreditLimit:    limit,
		IdempotencyKey: input.IdempotencyKey,
		Reason:         input.Reason,
		Actor:          ledgerActor(actor),
	})
	return result, ledgerError(err)
}

func (s *Service) FreezeTenantFunds(tenantID uint64, input TenantLedgerAmountInput, actor Actor) (any, error) {
	if err := s.ensurePlatformLedgerOperator(actor); err != nil {
		return nil, err
	}
	if err := s.ensureTenantSiteAvailable(tenantID, input.SiteID); err != nil {
		return nil, err
	}
	amount, err := parseLedgerAmount(input.Amount)
	if err != nil {
		return nil, err
	}
	result, err := tenantledger.New(s.db).Freeze(context.Background(), tenantID, input.SiteID, amount, input.Currency, ledgerActor(actor), input.IdempotencyKey, input.Reason)
	return result, ledgerError(err)
}

func (s *Service) UnfreezeTenantFunds(tenantID uint64, input TenantLedgerAmountInput, actor Actor) (any, error) {
	if err := s.ensurePlatformLedgerOperator(actor); err != nil {
		return nil, err
	}
	if err := s.ensureTenantSiteAvailable(tenantID, input.SiteID); err != nil {
		return nil, err
	}
	amount, err := parseLedgerAmount(input.Amount)
	if err != nil {
		return nil, err
	}
	result, err := tenantledger.New(s.db).Unfreeze(context.Background(), tenantID, input.SiteID, amount, input.Currency, ledgerActor(actor), input.IdempotencyKey, input.Reason)
	return result, ledgerError(err)
}

func (s *Service) ReverseTenantLedgerTransaction(tenantID, transactionID uint64, input TenantLedgerReverseInput, actor Actor) (any, error) {
	if err := s.ensurePlatformLedgerOperator(actor); err != nil {
		return nil, err
	}
	result, err := tenantledger.New(s.db).ReverseTransaction(context.Background(), transactionID, input.Reason, ledgerActor(actor))
	if err != nil {
		return nil, ledgerError(err)
	}
	if result.TenantID != tenantID {
		return nil, rowNotFound("ledger transaction")
	}
	return result, nil
}

func (s *Service) ensureTenantLedgerVisible(tenantID uint64, actor Actor) error {
	if actor.Platform {
		var count int64
		if err := s.db.Table("game_saas_tenants").Where("id=? AND deleted_at IS NULL", tenantID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return rowNotFound("tenant")
		}
		return nil
	}
	if actor.TenantID != tenantID {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	return s.ensureTenantAvailable(tenantID)
}

func (s *Service) ensurePlatformLedgerOperator(actor Actor) error {
	if !actor.Platform {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	return nil
}

func (s *Service) ensureTenantSiteAvailable(tenantID, siteID uint64) error {
	if err := s.ensureTenantAvailable(tenantID); err != nil {
		return err
	}
	if siteID == 0 {
		return nil
	}
	var count int64
	if err := s.db.Table("game_saas_sites").Where("id=? AND tenant_id=? AND enabled=1 AND deleted_at IS NULL", siteID, tenantID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return rowNotFound("site")
	}
	return nil
}

func parseLedgerAmount(value string) (decimal.Decimal, error) {
	amount, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil || amount.IsNegative() {
		return decimal.Zero, NewError("INVALID_AMOUNT", "invalid amount")
	}
	return amount, nil
}

func ledgerScope(actor Actor) tenancy.Scope {
	return tenancy.Scope{TenantID: actor.TenantID, SiteID: actor.SiteID, Platform: actor.Platform}
}

func ledgerActor(actor Actor) tenantledger.Actor {
	return tenantledger.Actor{
		ID: actor.ID, Username: actor.Username, TenantID: actor.TenantID, SiteID: actor.SiteID,
		Platform: actor.Platform, IP: actor.IP, UserAgent: actor.UserAgent, CorrelationID: actor.CorrelationID,
	}
}

func ledgerError(err error) error {
	if err == nil {
		return nil
	}
	if code := tenantledger.ErrorCode(err); code != "INTERNAL_ERROR" {
		return NewError(code, err.Error())
	}
	return err
}
