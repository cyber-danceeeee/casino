package reportadmin

import (
	"gorm.io/gorm"
)

type Services struct {
	Financial *FinancialReportService
	Game      *GameReportService
	Provider  *ProviderReportService
	User      *UserReportService
	Channel   *ChannelReportService
	Agent     *AgentReportService
	Promotion *PromotionReportService
	Export    *ReportExportService
}

type FinancialReportService struct{ db *gorm.DB }
type GameReportService struct{ db *gorm.DB }
type ProviderReportService struct{ db *gorm.DB }
type UserReportService struct{ db *gorm.DB }
type ChannelReportService struct{ db *gorm.DB }
type AgentReportService struct{ db *gorm.DB }
type PromotionReportService struct{ db *gorm.DB }
type ReportExportService struct {
	db      *gorm.DB
	reports *Services
}

func New(db *gorm.DB) *Services {
	s := &Services{
		Financial: &FinancialReportService{db: db},
		Game:      &GameReportService{db: db},
		Provider:  &ProviderReportService{db: db},
		User:      &UserReportService{db: db},
		Channel:   &ChannelReportService{db: db},
		Agent:     &AgentReportService{db: db},
		Promotion: &PromotionReportService{db: db},
	}
	s.Export = &ReportExportService{db: db, reports: s}
	return s
}

func metricBasis() map[string]string {
	return map[string]string{
		"deposit":       "SUCCESS_PAID_AT",
		"withdrawal":    "SUCCESS_COMPLETED_AT",
		"betting":       "SETTLED_AT_EXCLUDING_ROLLBACK",
		"bonus":         "ISSUED_AT_ACTIVE_OR_COMPLETED",
		"commission":    "BUSINESS_DATE_CONFIRMED_OR_SETTLED",
		"registration":  "USER_CREATED_AT",
		"businessClock": "BACKEND_UTC_PLUS_8_GROUPED_BY_REQUESTED_OFFSET",
	}
}
