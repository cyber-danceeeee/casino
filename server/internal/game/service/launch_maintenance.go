package service

import (
	"context"
	"strings"
	"time"

	"casino/internal/content"
	"casino/internal/game/domain"
)

type launchMaintenanceRow struct {
	MessageZh          string `gorm:"column:message_zh"`
	MessageEn          string `gorm:"column:message_en"`
	AffectedCountries  string `gorm:"column:affected_countries"`
	AffectedCurrencies string `gorm:"column:affected_currencies"`
}

func (s *LaunchService) launchMaintenance(ctx context.Context, scope content.Scope, providerCode string, gameID uint64, country string, currency string) (domain.WalletErrorCode, string, bool) {
	nowValue := time.Now()
	var providerRows []launchMaintenanceRow
	_ = s.repo.DB().WithContext(ctx).Table("game_provider_maintenance").
		Where("provider_code=? AND enabled=1 AND block_new_launches=1 AND deleted_at IS NULL", strings.ToUpper(providerCode)).
		Where("(site_id = 0) OR ((site_id = 0 OR site_id = ?))", scope.SiteID).
		Where("start_at IS NULL OR start_at<=?", nowValue).Where("end_at IS NULL OR end_at>?", nowValue).
		Order("site_id ASC, id ASC").Find(&providerRows).Error
	for _, row := range providerRows {
		if launchMaintenanceApplies(row.AffectedCountries, country) && launchMaintenanceApplies(row.AffectedCurrencies, currency) {
			return domain.WalletErrorProviderMaintenance, launchMaintenanceMessage(row, "Provider is under maintenance"), true
		}
	}
	var gameRow launchMaintenanceRow
	err := s.repo.DB().WithContext(ctx).Table("game_game_maintenance").
		Where("site_id = ? AND game_id = ? AND enabled = 1 AND block_new_launches = 1 AND deleted_at IS NULL", scope.SiteID, gameID).
		Where("start_at IS NULL OR start_at<=?", nowValue).Where("end_at IS NULL OR end_at>?", nowValue).
		Take(&gameRow).Error
	if err == nil && launchMaintenanceApplies(gameRow.AffectedCountries, country) && launchMaintenanceApplies(gameRow.AffectedCurrencies, currency) {
		return domain.WalletErrorGameMaintenance, launchMaintenanceMessage(gameRow, "Game is under maintenance"), true
	}
	return "", "", false
}

func launchMaintenanceApplies(csv string, value string) bool {
	if strings.TrimSpace(csv) == "" {
		return true
	}
	for _, item := range strings.Split(csv, ",") {
		if strings.EqualFold(strings.TrimSpace(item), "ALL") || strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(value)) {
			return true
		}
	}
	return false
}

func launchMaintenanceMessage(row launchMaintenanceRow, fallback string) string {
	if strings.TrimSpace(row.MessageEn) != "" {
		return row.MessageEn
	}
	if strings.TrimSpace(row.MessageZh) != "" {
		return row.MessageZh
	}
	return fallback
}
