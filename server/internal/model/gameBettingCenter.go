package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

type GameBetOrders struct {
	sgorm.Model `gorm:"embedded"`

	SiteID                   uint64     `gorm:"column:site_id" json:"siteID"`
	BetNo                    string     `gorm:"column:bet_no" json:"betNo"`
	PlatformBetID            string     `gorm:"column:platform_bet_id" json:"platformBetID"`
	UserID                   uint64     `gorm:"column:user_id" json:"userID"`
	WalletID                 uint64     `gorm:"column:wallet_id" json:"walletID"`
	GameID                   uint64     `gorm:"column:game_id" json:"gameID"`
	ProviderID               uint64     `gorm:"column:provider_id" json:"providerID"`
	ProviderCode             string     `gorm:"column:provider_code" json:"providerCode"`
	ProviderGameID           string     `gorm:"column:provider_game_id" json:"providerGameID"`
	ProviderRoundID          string     `gorm:"column:provider_round_id" json:"providerRoundID"`
	ProviderBetTransactionID string     `gorm:"column:provider_bet_transaction_id" json:"providerBetTransactionID"`
	ProviderWinTransactionID string     `gorm:"column:provider_win_transaction_id" json:"providerWinTransactionID"`
	GameType                 string     `gorm:"column:game_type" json:"gameType"`
	BetAmountMinor           int64      `gorm:"column:bet_amount_minor" json:"betAmountMinor"`
	ValidBetAmountMinor      int64      `gorm:"column:valid_bet_amount_minor" json:"validBetAmountMinor"`
	WinAmountMinor           int64      `gorm:"column:win_amount_minor" json:"winAmountMinor"`
	PlatformProfitMinor      int64      `gorm:"column:platform_profit_minor" json:"platformProfitMinor"`
	Currency                 string     `gorm:"column:currency" json:"currency"`
	BetStatus                string     `gorm:"column:bet_status" json:"status"`
	RollbackStatus           string     `gorm:"column:rollback_status" json:"rollbackStatus"`
	FailureReason            string     `gorm:"column:failure_reason" json:"failureReason"`
	BetAt                    *time.Time `gorm:"column:bet_at" json:"betAt"`
	SettledAt                *time.Time `gorm:"column:settled_at" json:"settledAt"`
}

func (GameBetOrders) TableName() string { return "game_bet_records" }

type GameRounds struct {
	sgorm.Model `gorm:"embedded"`

	SiteID              uint64     `gorm:"column:site_id" json:"siteID"`
	ProviderCode        string     `gorm:"column:provider_code" json:"providerCode"`
	ProviderRoundID     string     `gorm:"column:provider_round_id" json:"providerRoundID"`
	ProviderGameID      string     `gorm:"column:provider_game_id" json:"providerGameID"`
	GameID              uint64     `gorm:"column:game_id" json:"gameID"`
	UserID              uint64     `gorm:"column:user_id" json:"userID"`
	WalletID            uint64     `gorm:"column:wallet_id" json:"walletID"`
	TotalBetMinor       int64      `gorm:"column:total_bet_minor" json:"totalBetMinor"`
	TotalWinMinor       int64      `gorm:"column:total_win_minor" json:"totalWinMinor"`
	PlatformProfitMinor int64      `gorm:"column:platform_profit_minor" json:"platformProfitMinor"`
	Currency            string     `gorm:"column:currency" json:"currency"`
	TransactionCount    uint       `gorm:"column:transaction_count" json:"transactionCount"`
	Status              string     `gorm:"column:status" json:"status"`
	StartedAt           *time.Time `gorm:"column:started_at" json:"startedAt"`
	CompletedAt         *time.Time `gorm:"column:completed_at" json:"completedAt"`
}

func (GameRounds) TableName() string { return "game_rounds" }

type GameRollbacks struct {
	sgorm.Model `gorm:"embedded"`

	SiteID                        uint64     `gorm:"column:site_id" json:"siteID"`
	RollbackID                    string     `gorm:"column:rollback_id" json:"rollbackID"`
	ProviderCode                  string     `gorm:"column:provider_code" json:"providerCode"`
	ProviderTransactionID         string     `gorm:"column:provider_transaction_id" json:"providerTransactionID"`
	ProviderRoundID               string     `gorm:"column:provider_round_id" json:"providerRoundID"`
	OriginalProviderTransactionID string     `gorm:"column:original_provider_transaction_id" json:"originalProviderTransactionID"`
	OriginalWalletTransactionID   uint64     `gorm:"column:original_wallet_transaction_id" json:"originalWalletTransactionID"`
	RollbackWalletTransactionID   uint64     `gorm:"column:rollback_wallet_transaction_id" json:"rollbackWalletTransactionID"`
	BetRecordID                   uint64     `gorm:"column:bet_record_id" json:"betRecordID"`
	UserID                        uint64     `gorm:"column:user_id" json:"userID"`
	GameID                        uint64     `gorm:"column:game_id" json:"gameID"`
	ProviderGameID                string     `gorm:"column:provider_game_id" json:"providerGameID"`
	OriginalOperation             string     `gorm:"column:original_operation" json:"originalOperation"`
	OriginalAmountMinor           int64      `gorm:"column:original_amount_minor" json:"originalAmountMinor"`
	RollbackAmountMinor           int64      `gorm:"column:rollback_amount_minor" json:"rollbackAmountMinor"`
	Currency                      string     `gorm:"column:currency" json:"currency"`
	Status                        string     `gorm:"column:status" json:"status"`
	FailureReason                 string     `gorm:"column:failure_reason" json:"failureReason"`
	CompletedAt                   *time.Time `gorm:"column:completed_at" json:"completedAt"`
}

func (GameRollbacks) TableName() string { return "game_rollbacks" }

type GameProviderCallbackLogs struct {
	sgorm.Model `gorm:"embedded"`

	SiteID                 uint64 `gorm:"column:site_id" json:"siteID"`
	CallbackID             string `gorm:"column:callback_id" json:"callbackID"`
	ProviderCode           string `gorm:"column:provider_code" json:"providerCode"`
	Operation              string `gorm:"column:operation" json:"operation"`
	UserID                 uint64 `gorm:"column:user_id" json:"userID"`
	ProviderPlayerID       string `gorm:"column:provider_player_id" json:"providerPlayerID"`
	ProviderTransactionID  string `gorm:"column:provider_transaction_id" json:"providerTransactionID"`
	ProviderRoundID        string `gorm:"column:provider_round_id" json:"providerRoundID"`
	ProviderGameID         string `gorm:"column:provider_game_id" json:"providerGameID"`
	AmountMinor            int64  `gorm:"column:amount_minor" json:"amountMinor"`
	Currency               string `gorm:"column:currency" json:"currency"`
	HTTPMethod             string `gorm:"column:http_method" json:"httpMethod"`
	RequestPath            string `gorm:"column:request_path" json:"requestPath"`
	RequestHeaders         string `gorm:"column:request_headers;type:json" json:"requestHeaders"`
	QueryParameters        string `gorm:"column:query_parameters;type:json" json:"queryParameters"`
	RequestBody            string `gorm:"column:request_body;type:json" json:"requestBody"`
	SourceIP               string `gorm:"column:source_ip" json:"sourceIP"`
	SignatureResult        string `gorm:"column:signature_result" json:"signatureResult"`
	TimestampValidation    string `gorm:"column:timestamp_validation" json:"timestampValidation"`
	ProcessingResult       string `gorm:"column:processing_result" json:"processingResult"`
	HTTPStatus             int    `gorm:"column:http_status" json:"httpStatus"`
	NormalizedCommand      string `gorm:"column:normalized_command;type:json" json:"normalizedCommand"`
	WalletResult           string `gorm:"column:wallet_result;type:json" json:"walletResult"`
	ProviderResponse       string `gorm:"column:provider_response;type:json" json:"providerResponse"`
	DurationMS             uint64 `gorm:"column:duration_ms" json:"durationMS"`
	CorrelationID          string `gorm:"column:correlation_id" json:"correlationID"`
	FailureReason          string `gorm:"column:failure_reason" json:"failureReason"`
	MarkedForInvestigation bool   `gorm:"column:marked_for_investigation" json:"markedForInvestigation"`
}

func (GameProviderCallbackLogs) TableName() string { return "game_provider_callback_logs" }

type GameAbnormalBetEvents struct {
	sgorm.Model `gorm:"embedded"`

	SiteID             uint64 `gorm:"column:site_id" json:"siteID"`
	RiskEventID        string `gorm:"column:risk_event_id" json:"riskEventID"`
	BetRecordID        uint64 `gorm:"column:bet_record_id" json:"betRecordID"`
	RoundID            uint64 `gorm:"column:round_id" json:"roundID"`
	CallbackLogID      uint64 `gorm:"column:callback_log_id" json:"callbackLogID"`
	UserID             uint64 `gorm:"column:user_id" json:"userID"`
	ProviderCode       string `gorm:"column:provider_code" json:"providerCode"`
	GameID             uint64 `gorm:"column:game_id" json:"gameID"`
	ProviderRoundID    string `gorm:"column:provider_round_id" json:"providerRoundID"`
	RiskType           string `gorm:"column:risk_type" json:"riskType"`
	RiskLevel          string `gorm:"column:risk_level" json:"riskLevel"`
	ReviewStatus       string `gorm:"column:review_status" json:"reviewStatus"`
	BetAmountMinor     int64  `gorm:"column:bet_amount_minor" json:"betAmountMinor"`
	WinAmountMinor     int64  `gorm:"column:win_amount_minor" json:"winAmountMinor"`
	TriggerCount       uint   `gorm:"column:trigger_count" json:"triggerCount"`
	AssignedReviewerID uint64 `gorm:"column:assigned_reviewer_id" json:"assignedReviewerID"`
	AssignedReviewer   string `gorm:"column:assigned_reviewer" json:"assignedReviewer"`
	LatestReason       string `gorm:"column:latest_reason" json:"latestReason"`
}

func (GameAbnormalBetEvents) TableName() string { return "game_abnormal_bet_events" }

type GameAbnormalBetReviewHistories struct {
	sgorm.Model `gorm:"embedded"`

	SiteID           uint64 `gorm:"column:site_id" json:"siteID"`
	AbnormalBetID    uint64 `gorm:"column:abnormal_bet_id" json:"abnormalBetID"`
	Action           string `gorm:"column:action" json:"action"`
	BeforeStatus     string `gorm:"column:before_status" json:"beforeStatus"`
	AfterStatus      string `gorm:"column:after_status" json:"afterStatus"`
	OperatorID       uint64 `gorm:"column:operator_id" json:"operatorID"`
	OperatorUsername string `gorm:"column:operator_username" json:"operatorUsername"`
	Reason           string `gorm:"column:reason" json:"reason"`
}

func (GameAbnormalBetReviewHistories) TableName() string { return "game_abnormal_bet_review_histories" }

type GameAbnormalBetNotes struct {
	sgorm.Model `gorm:"embedded"`

	SiteID           uint64 `gorm:"column:site_id" json:"siteID"`
	AbnormalBetID    uint64 `gorm:"column:abnormal_bet_id" json:"abnormalBetID"`
	OperatorID       uint64 `gorm:"column:operator_id" json:"operatorID"`
	OperatorUsername string `gorm:"column:operator_username" json:"operatorUsername"`
	Content          string `gorm:"column:content" json:"content"`
}

func (GameAbnormalBetNotes) TableName() string { return "game_abnormal_bet_notes" }
