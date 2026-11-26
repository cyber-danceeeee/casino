package betadmin

import (
	"time"

	"casino/internal/content"
)

type ListQuery struct {
	Page     int
	PageSize int
	Values   map[string]string
	Scope    content.Scope
}

type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type Actor struct {
	ID            uint64
	Username      string
	Scope         content.Scope
	IP            string
	UserAgent     string
	CorrelationID string
}

type ReasonCommand struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

type AssignCommand struct {
	ReviewerID uint64 `json:"reviewerID" binding:"required"`
	Reviewer   string `json:"reviewer" binding:"required,max=64"`
	Reason     string `json:"reason" binding:"required,max=500"`
}

type ReviewCommand struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason" binding:"required,max=500"`
}

type NoteCommand struct {
	Content string `json:"content" binding:"required,max=1000"`
}

type GenerateReconciliationCommand struct {
	SiteID       uint64 `json:"siteID"`
	ProviderCode string `json:"providerCode" binding:"required,max=64"`
	BusinessDate string `json:"businessDate" binding:"required"`
	Currency     string `json:"currency" binding:"required,max=8"`
}

type ReconciliationStatusCommand struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason" binding:"required,max=500"`
}

type BetRow struct {
	ID                       uint64     `json:"id"`
	PlatformBetID            string     `json:"platformBetID"`
	UserID                   uint64     `json:"userID"`
	UserNo                   string     `json:"userNo"`
	Username                 string     `json:"username"`
	GameID                   uint64     `json:"gameID"`
	GameName                 string     `json:"gameName"`
	GameIcon                 string     `json:"gameIcon"`
	ProviderCode             string     `json:"providerCode"`
	ProviderGameID           string     `json:"providerGameID"`
	ProviderRoundID          string     `json:"providerRoundID"`
	ProviderBetTransactionID string     `json:"providerBetTransactionID"`
	ProviderWinTransactionID string     `json:"providerWinTransactionID"`
	GameType                 string     `json:"gameType"`
	BetAmountMinor           int64      `json:"betAmountMinor"`
	ValidBetAmountMinor      int64      `json:"validBetAmountMinor"`
	WinAmountMinor           int64      `json:"winAmountMinor"`
	PlatformProfitMinor      int64      `json:"platformProfitMinor"`
	BetAmount                string     `json:"betAmount"`
	ValidBetAmount           string     `json:"validBetAmount"`
	WinAmount                string     `json:"winAmount"`
	PlatformProfit           string     `json:"platformProfit"`
	Currency                 string     `json:"currency"`
	Status                   string     `json:"status"`
	RollbackStatus           string     `json:"rollbackStatus"`
	FailureReason            string     `json:"failureReason"`
	BetAt                    *time.Time `json:"betAt"`
	SettledAt                *time.Time `json:"settledAt"`
	CreatedAt                time.Time  `json:"createdAt"`
	UpdatedAt                time.Time  `json:"updatedAt"`
}

type RoundRow struct {
	ID                  uint64     `json:"id"`
	ProviderRoundID     string     `json:"providerRoundID"`
	UserID              uint64     `json:"userID"`
	UserNo              string     `json:"userNo"`
	Username            string     `json:"username"`
	GameID              uint64     `json:"gameID"`
	GameName            string     `json:"gameName"`
	ProviderCode        string     `json:"providerCode"`
	ProviderGameID      string     `json:"providerGameID"`
	TotalBetMinor       int64      `json:"totalBetMinor"`
	TotalWinMinor       int64      `json:"totalWinMinor"`
	PlatformProfitMinor int64      `json:"platformProfitMinor"`
	TotalBet            string     `json:"totalBet"`
	TotalWin            string     `json:"totalWin"`
	PlatformProfit      string     `json:"platformProfit"`
	Currency            string     `json:"currency"`
	TransactionCount    uint       `json:"transactionCount"`
	Status              string     `json:"status"`
	StartedAt           *time.Time `json:"startedAt"`
	CompletedAt         *time.Time `json:"completedAt"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type RollbackRow struct {
	ID                            uint64     `json:"id"`
	RollbackID                    string     `json:"rollbackID"`
	OriginalProviderTransactionID string     `json:"originalTransactionID"`
	ProviderTransactionID         string     `json:"providerTransactionID"`
	ProviderRoundID               string     `json:"providerRoundID"`
	UserID                        uint64     `json:"userID"`
	UserNo                        string     `json:"userNo"`
	Username                      string     `json:"username"`
	ProviderCode                  string     `json:"providerCode"`
	GameID                        uint64     `json:"gameID"`
	GameName                      string     `json:"gameName"`
	OriginalOperation             string     `json:"originalOperation"`
	OriginalAmountMinor           int64      `json:"originalAmountMinor"`
	RollbackAmountMinor           int64      `json:"rollbackAmountMinor"`
	OriginalAmount                string     `json:"originalAmount"`
	RollbackAmount                string     `json:"rollbackAmount"`
	Currency                      string     `json:"currency"`
	Status                        string     `json:"status"`
	FailureReason                 string     `json:"failureReason"`
	CreatedAt                     time.Time  `json:"createdAt"`
	CompletedAt                   *time.Time `json:"completedAt"`
}

type AbnormalRow struct {
	ID               uint64    `json:"id"`
	RiskEventID      string    `json:"riskEventID"`
	BetRecordID      uint64    `json:"betRecordID"`
	UserID           uint64    `json:"userID"`
	UserNo           string    `json:"userNo"`
	Username         string    `json:"username"`
	ProviderCode     string    `json:"providerCode"`
	GameID           uint64    `json:"gameID"`
	GameName         string    `json:"gameName"`
	ProviderRoundID  string    `json:"providerRoundID"`
	RiskType         string    `json:"riskType"`
	RiskLevel        string    `json:"riskLevel"`
	ReviewStatus     string    `json:"reviewStatus"`
	BetAmountMinor   int64     `json:"betAmountMinor"`
	WinAmountMinor   int64     `json:"winAmountMinor"`
	BetAmount        string    `json:"betAmount"`
	WinAmount        string    `json:"winAmount"`
	TriggerCount     uint      `json:"triggerCount"`
	AssignedReviewer string    `json:"assignedReviewer"`
	LatestReason     string    `json:"latestReason"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type CallbackRow struct {
	ID                     uint64    `json:"id"`
	CallbackID             string    `json:"callbackID"`
	ProviderCode           string    `json:"providerCode"`
	Operation              string    `json:"operation"`
	UserID                 uint64    `json:"userID"`
	UserNo                 string    `json:"userNo"`
	Username               string    `json:"username"`
	UserPhone              string    `json:"userPhone"`
	GameID                 uint64    `json:"gameID"`
	GameName               string    `json:"gameName"`
	ProviderPlayerID       string    `json:"providerPlayerID"`
	ProviderTransactionID  string    `json:"providerTransactionID"`
	ProviderRoundID        string    `json:"providerRoundID"`
	ProviderGameID         string    `json:"providerGameID"`
	AmountMinor            int64     `json:"amountMinor"`
	Amount                 string    `json:"amount"`
	Currency               string    `json:"currency"`
	HTTPMethod             string    `json:"httpMethod"`
	SignatureResult        string    `json:"signatureResult"`
	ProcessingResult       string    `json:"processingResult"`
	HTTPStatus             int       `json:"httpStatus"`
	SourceIP               string    `json:"sourceIP"`
	FailureReason          string    `json:"failureReason"`
	MarkedForInvestigation bool      `json:"markedForInvestigation"`
	CreatedAt              time.Time `json:"createdAt"`
}

type ReconciliationRunRow struct {
	ID    uint64 `json:"id"`
	RunNo string `json:"runNo"`

	SiteID                   uint64     `json:"siteID"`
	ProviderID               uint64     `json:"providerID"`
	ProviderCode             string     `json:"providerCode"`
	BusinessDate             time.Time  `json:"businessDate"`
	Currency                 string     `json:"currency"`
	SourceMode               string     `json:"sourceMode"`
	Status                   string     `json:"status"`
	ProviderTransactionCount int64      `json:"providerTransactionCount"`
	CallbackCount            int64      `json:"callbackCount"`
	WalletTransactionCount   int64      `json:"walletTransactionCount"`
	BetRecordCount           int64      `json:"betRecordCount"`
	DiscrepancyCount         int64      `json:"discrepancyCount"`
	OpenDiscrepancyCount     int64      `json:"openDiscrepancyCount"`
	StartedAt                time.Time  `json:"startedAt"`
	CompletedAt              *time.Time `json:"completedAt"`
	FailureReason            string     `json:"failureReason"`
	CreatedByUsername        string     `json:"createdByUsername"`
	CreatedAt                time.Time  `json:"createdAt"`
}

type ReconciliationItemRow struct {
	ID     uint64 `json:"id"`
	RunID  uint64 `json:"runID"`
	ItemNo string `json:"itemNo"`

	SiteID                   uint64     `json:"siteID"`
	ProviderID               uint64     `json:"providerID"`
	ProviderCode             string     `json:"providerCode"`
	BusinessDate             time.Time  `json:"businessDate"`
	Currency                 string     `json:"currency"`
	DiscrepancyType          string     `json:"discrepancyType"`
	Severity                 string     `json:"severity"`
	Status                   string     `json:"status"`
	UserID                   uint64     `json:"userID"`
	UserNo                   string     `json:"userNo"`
	Username                 string     `json:"username"`
	ProviderTransactionRowID uint64     `json:"providerTransactionRowID"`
	ProviderTransactionID    string     `json:"providerTransactionID"`
	ProviderRoundID          string     `json:"providerRoundID"`
	ProviderGameID           string     `json:"providerGameID"`
	CallbackLogID            uint64     `json:"callbackLogID"`
	WalletTransactionID      uint64     `json:"walletTransactionID"`
	BetRecordID              uint64     `json:"betRecordID"`
	RollbackRecordID         uint64     `json:"rollbackRecordID"`
	ExpectedAmountMinor      int64      `json:"expectedAmountMinor"`
	ActualAmountMinor        int64      `json:"actualAmountMinor"`
	DifferenceAmountMinor    int64      `json:"differenceAmountMinor"`
	ExpectedAmount           string     `json:"expectedAmount"`
	ActualAmount             string     `json:"actualAmount"`
	DifferenceAmount         string     `json:"differenceAmount"`
	Detail                   string     `json:"detail"`
	ResolvedByUsername       string     `json:"resolvedByUsername"`
	ResolvedAt               *time.Time `json:"resolvedAt"`
	ResolutionReason         string     `json:"resolutionReason"`
	CreatedAt                time.Time  `json:"createdAt"`
	UpdatedAt                time.Time  `json:"updatedAt"`
}

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

func newError(code string, message string) error { return &Error{Code: code, Message: message} }
