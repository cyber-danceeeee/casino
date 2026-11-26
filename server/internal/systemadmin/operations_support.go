package systemadmin

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PlayerSupportRow struct {
	ID               uint64 `json:"playerID"`
	UserNo           string `json:"userNo"`
	Username         string `json:"username"`
	PhoneCountryCode string `json:"phoneCountryCode"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`

	SiteID           uint64           `json:"siteID"`
	SiteName         string           `json:"siteName"`
	Status           int              `json:"status"`
	CountryCode      string           `json:"countryCode"`
	Currency         string           `json:"currency"`
	AvailableBalance *decimal.Decimal `json:"availableBalance"`
	FrozenBalance    *decimal.Decimal `json:"frozenBalance"`
	BonusBalance     *decimal.Decimal `json:"bonusBalance"`
	VIPLevel         uint             `json:"vipLevel"`
	RegisteredAt     time.Time        `json:"registeredAt"`
	RegisteredIP     string           `json:"registeredIP"`
	LastLoginAt      *time.Time       `json:"lastLoginAt"`
	LastLoginIP      string           `json:"lastLoginIP"`
	KYCStatus        string           `json:"kycStatus"`
	RiskStatus       string           `json:"riskStatus"`
}

type TimelineItem struct {
	ID              string           `json:"id"`
	EventType       string           `json:"eventType"`
	OccurredAt      time.Time        `json:"occurredAt"`
	Amount          *decimal.Decimal `json:"amount,omitempty"`
	Currency        string           `json:"currency,omitempty"`
	BalanceBefore   *decimal.Decimal `json:"balanceBefore,omitempty"`
	BalanceAfter    *decimal.Decimal `json:"balanceAfter,omitempty"`
	BusinessOrderNo string           `json:"businessOrderNo,omitempty"`
	ExternalOrderNo string           `json:"externalOrderNo,omitempty"`
	ProviderCode    string           `json:"providerCode,omitempty"`
	Status          string           `json:"status,omitempty"`
	RequestID       string           `json:"requestID,omitempty"`
	Operator        string           `json:"operator,omitempty"`
	Description     string           `json:"description"`
	DetailType      string           `json:"detailType"`
	DetailID        uint64           `json:"detailID"`
}

type SupportCaseInput struct {
	SiteID                uint64   `json:"siteID"`
	PlayerID              uint64   `json:"playerID"`
	PlayerNo              string   `json:"playerNo"`
	Category              string   `json:"category"`
	Priority              string   `json:"priority"`
	Status                string   `json:"status"`
	Title                 string   `json:"title"`
	Description           string   `json:"description"`
	AssignedAdminID       uint64   `json:"assignedAdminID"`
	Resolution            string   `json:"resolution"`
	RelatedOrderIDs       []string `json:"relatedOrderIDs"`
	RelatedTransactionIDs []string `json:"relatedTransactionIDs"`
	Reason                string   `json:"reason"`
	IdempotencyKey        string   `json:"idempotencyKey"`
}

type SupportCaseNoteInput struct {
	Note   string `json:"note"`
	Reason string `json:"reason"`
}

func (s *Service) SearchPlayers(q Query, actor Actor) (Result, error) {
	term := strings.TrimSpace(value(q, "query"))
	searchType := strings.ToUpper(defaultString(value(q, "searchType"), "AUTO"))
	if len([]rune(term)) < 2 || strings.ContainsAny(term, "%_") {
		return Result{}, NewError("SEARCH_TERM_REQUIRED", "enter at least two literal characters")
	}
	if searchType == "AUTO" && len([]rune(term)) < 4 {
		return Result{}, NewError("SEARCH_TERM_TOO_SHORT", "automatic search requires at least four characters")
	}
	allowed := map[string]bool{"AUTO": true, "PLAYER_ID": true, "PLAYER_NUMBER": true, "USERNAME": true, "PHONE": true, "EMAIL": true, "EXTERNAL_PLAYER_ID": true, "REGISTRATION_IP": true, "DEVICE_ID": true, "BANK_ACCOUNT": true, "DEPOSIT_ORDER": true, "WITHDRAWAL_ORDER": true, "GAME_ROUND_ID": true, "PROVIDER_TRANSACTION_ID": true}
	if !allowed[searchType] {
		return Result{}, NewError("INVALID_SEARCH_TYPE", "unsupported player search type")
	}

	rows := []PlayerSupportRow{}
	db := s.db.Table("game_users u").Select("u.id, u.user_no, u.username, u.phone_country_code, u.phone, u.email, u.site_id, s.site_name, u.status, u.country_code, u.vip_level, COALESCE(w.currency, '') AS currency, COALESCE(w.cash_balance, 0) AS available_balance, COALESCE(w.frozen_balance, 0) AS frozen_balance, COALESCE(w.bonus_balance, 0) AS bonus_balance, u.created_at AS registered_at, u.registered_ip, u.last_login_at, u.last_login_ip, COALESCE((SELECT wo.kyc_status FROM game_withdraw_orders AS wo WHERE wo.user_id = u.id AND wo.deleted_at IS NULL ORDER BY wo.id DESC LIMIT 1), '') AS kyc_status, COALESCE((SELECT rp.risk_status FROM game_user_risk_profiles AS rp WHERE rp.user_id = u.id AND rp.site_id = u.site_id AND rp.deleted_at IS NULL ORDER BY rp.id DESC LIMIT 1), CASE WHEN u.risk_review_required = 1 THEN 'REVIEW' ELSE 'NORMAL' END) AS risk_status").
		Joins("LEFT JOIN game_site_config AS s ON s.id = u.site_id AND s.deleted_at IS NULL").
		Joins("LEFT JOIN game_user_wallets AS w ON w.user_id = u.id AND w.site_id = u.site_id AND w.deleted_at IS NULL").
		Where("u.deleted_at IS NULL")
	var err error
	db, err = operationalScopeAlias(db, "u", actor, value(q, "siteID"))
	if err != nil {
		return Result{}, err
	}
	db = applyPlayerSearch(db, searchType, term)
	result, err := paginate(db.Order("u.id DESC"), q, &rows)
	if err != nil {
		return Result{}, err
	}
	for index := range rows {
		maskPlayerSupport(&rows[index], actor.Sensitive)
	}
	result.List = rows
	return result, nil
}

func applyPlayerSearch(db *gorm.DB, searchType string, term string) *gorm.DB {
	contains := "%" + term + "%"
	switch searchType {
	case "PLAYER_NUMBER":
		return db.Where("u.user_no=?", term)
	case "PLAYER_ID":
		id, _ := strconv.ParseUint(term, 10, 64)
		return db.Where("u.id=? OR u.user_no=?", id, term)
	case "USERNAME":
		return db.Where("u.username LIKE ?", contains)
	case "PHONE":
		return db.Where("u.phone LIKE ?", contains)
	case "EMAIL":
		return db.Where("u.email LIKE ?", contains)
	case "REGISTRATION_IP":
		return db.Where("u.registered_ip=?", term)
	case "EXTERNAL_PLAYER_ID":
		return db.Where("EXISTS(SELECT 1 FROM game_provider_players AS pp WHERE pp.user_id = u.id AND pp.site_id = u.site_id AND pp.provider_player_id = ? AND pp.deleted_at IS NULL)", term)
	case "DEVICE_ID":
		return db.Where("EXISTS(SELECT 1 FROM game_user_devices AS d WHERE d.user_id = u.id AND d.site_id = u.site_id AND d.device_id = ? AND d.deleted_at IS NULL)", term)
	case "BANK_ACCOUNT":
		return db.Where("EXISTS(SELECT 1 FROM game_withdraw_accounts AS a WHERE a.user_id = u.id AND a.site_id = u.site_id AND (a.account_no = ? OR a.phone = ?) AND a.deleted_at IS NULL)", term, term)
	case "DEPOSIT_ORDER":
		return db.Where("EXISTS(SELECT 1 FROM game_deposit_orders AS o WHERE o.user_id = u.id AND o.site_id = u.site_id AND (o.order_no = ? OR o.channel_order_no = ?) AND o.deleted_at IS NULL)", term, term)
	case "WITHDRAWAL_ORDER":
		return db.Where("EXISTS(SELECT 1 FROM game_withdraw_orders AS o WHERE o.user_id = u.id AND o.site_id = u.site_id AND (o.order_no = ? OR o.channel_order_no = ?) AND o.deleted_at IS NULL)", term, term)
	case "GAME_ROUND_ID":
		return db.Where("EXISTS(SELECT 1 FROM game_bet_records AS b WHERE b.user_id = u.id AND b.site_id = u.site_id AND b.provider_round_id = ? AND b.deleted_at IS NULL)", term)
	case "PROVIDER_TRANSACTION_ID":
		return db.Where("EXISTS(SELECT 1 FROM game_provider_transactions AS p WHERE p.user_id = u.id AND p.site_id = u.site_id AND p.provider_transaction_id = ? AND p.deleted_at IS NULL)", term)
	default:
		id, _ := strconv.ParseUint(term, 10, 64)
		return db.Where("u.id = ? OR u.user_no = ? OR u.username LIKE ? OR u.phone LIKE ? OR u.email LIKE ? OR u.registered_ip = ? OR EXISTS(SELECT 1 FROM game_provider_players AS pp WHERE pp.user_id = u.id AND pp.site_id = u.site_id AND pp.provider_player_id = ? AND pp.deleted_at IS NULL) OR EXISTS(SELECT 1 FROM game_user_devices AS d WHERE d.user_id = u.id AND d.site_id = u.site_id AND d.device_id = ? AND d.deleted_at IS NULL) OR EXISTS(SELECT 1 FROM game_withdraw_accounts AS a WHERE a.user_id = u.id AND a.site_id = u.site_id AND (a.account_no = ? OR a.phone = ?) AND a.deleted_at IS NULL) OR EXISTS(SELECT 1 FROM game_deposit_orders AS o WHERE o.user_id = u.id AND o.site_id = u.site_id AND (o.order_no = ? OR o.channel_order_no = ?) AND o.deleted_at IS NULL) OR EXISTS(SELECT 1 FROM game_withdraw_orders AS o WHERE o.user_id = u.id AND o.site_id = u.site_id AND (o.order_no = ? OR o.channel_order_no = ?) AND o.deleted_at IS NULL) OR EXISTS(SELECT 1 FROM game_provider_transactions AS p WHERE p.user_id = u.id AND p.site_id = u.site_id AND (p.provider_transaction_id = ? OR p.provider_round_id = ?) AND p.deleted_at IS NULL)", id, term, contains, contains, contains, term, term, term, term, term, term, term, term, term, term, term)
	}
}

func (s *Service) Player360(playerID uint64, actor Actor) (map[string]any, error) {
	player, err := s.supportPlayer(playerID, actor)
	if err != nil {
		return nil, err
	}
	wallets := []map[string]any{}
	deposits := []map[string]any{}
	withdrawals := []map[string]any{}
	games := []map[string]any{}
	promotions := []map[string]any{}
	logins := []map[string]any{}
	devices := []map[string]any{}
	risk := []map[string]any{}
	audits := []map[string]any{}
	tags := []map[string]any{}
	scope := []any{player.SiteID, player.ID}
	_ = s.db.Table("game_user_wallets").Select("id,currency,cash_balance cashBalance,bonus_balance bonusBalance,withdrawable_balance withdrawableBalance,frozen_balance frozenBalance,status,last_transaction_at lastTransactionAt").Where("site_id = ? AND user_id = ? AND deleted_at IS NULL", scope...).Find(&wallets).Error
	_ = s.db.Table("game_deposit_orders").Select("id,order_no orderNo,channel_order_no channelOrderNo,amount,paid_amount paidAmount,currency,method_code methodCode,channel_code channelCode,status,created_at createdAt,paid_at paidAt,completed_at completedAt").Where("site_id = ? AND user_id = ? AND deleted_at IS NULL", scope...).Order("id DESC").Limit(100).Find(&deposits).Error
	_ = s.db.Table("game_withdraw_orders").Select("id,order_no orderNo,channel_order_no channelOrderNo,amount,actual_amount actualAmount,currency,payment_method paymentMethod,status,audit_status auditStatus,risk_status riskStatus,kyc_status kycStatus,created_at createdAt,completed_at completedAt").Where("site_id = ? AND user_id = ? AND deleted_at IS NULL", scope...).Order("id DESC").Limit(100).Find(&withdrawals).Error
	_ = s.db.Table("game_bet_records").Select("id,bet_no betNo,provider_code providerCode,provider_round_id providerRoundID,bet_amount_minor betAmountMinor,win_amount_minor winAmountMinor,currency,bet_status status,bet_at betAt,settled_at settledAt").Where("site_id = ? AND user_id = ? AND deleted_at IS NULL", scope...).Order("id DESC").Limit(100).Find(&games).Error
	_ = s.db.Table("game_promotion_claims").Select("id,promotion_id promotionID,trigger_type triggerType,trigger_order_no triggerOrderNo,currency,reward_amount rewardAmount,status,trace_id traceID,created_at createdAt").Where("site_id = ? AND user_id = ? AND deleted_at IS NULL", scope...).Order("id DESC").Limit(100).Find(&promotions).Error
	_ = s.db.Table("game_user_login_records").Select("id,success,ip_address ipAddress,country,device_id deviceID,device_type deviceType,platform,os,browser,occurred_at occurredAt").Where("site_id = ? AND user_id = ? AND deleted_at IS NULL", scope...).Order("occurred_at DESC").Limit(100).Find(&logins).Error
	_ = s.db.Table("game_user_devices").Select("id,device_id deviceID,platform,device_model deviceModel,os,browser,first_ip firstIP,last_ip lastIP,first_seen_at firstSeenAt,last_seen_at lastSeenAt,status,trust_status trustStatus,risk_status riskStatus").Where("site_id = ? AND user_id = ? AND deleted_at IS NULL", scope...).Order("last_seen_at DESC").Limit(100).Find(&devices).Error
	_ = s.db.Table("game_user_risk_events").Select("id,event_type eventType,risk_level riskLevel,source,description,operator_username operator,created_at createdAt").Where("site_id = ? AND user_id = ? AND deleted_at IS NULL", scope...).Order("id DESC").Limit(100).Find(&risk).Error
	_ = s.db.Table("game_admin_operation_logs").Select("id,action,result,operator_username operator,reason,correlation_id requestID,created_at createdAt").Where("site_id = ? AND target_user_id = ? AND deleted_at IS NULL", scope...).Order("id DESC").Limit(100).Find(&audits).Error
	_ = s.db.Table("game_user_tag_relations r").Select("t.id,t.name,t.tag_type tagType").Joins("JOIN game_user_tags AS t ON t.id = r.tag_id AND t.site_id = r.site_id AND t.deleted_at IS NULL").Where("r.site_id = ? AND r.user_id = ? AND r.deleted_at IS NULL", scope...).Find(&tags).Error

	summary := map[string]any{}
	_ = s.db.Raw("SELECT (SELECT COALESCE(SUM(paid_amount), 0) FROM game_deposit_orders WHERE site_id = ? AND user_id = ? AND status = 2 AND deleted_at IS NULL) AS depositAmount, (SELECT COUNT(*) FROM game_deposit_orders WHERE site_id = ? AND user_id = ? AND status = 2 AND deleted_at IS NULL) AS depositCount, (SELECT COALESCE(SUM(actual_amount), 0) FROM game_withdraw_orders WHERE site_id = ? AND user_id = ? AND status = 2 AND deleted_at IS NULL) AS withdrawalAmount, (SELECT COUNT(*) FROM game_withdraw_orders WHERE site_id = ? AND user_id = ? AND status = 2 AND deleted_at IS NULL) AS withdrawalCount, (SELECT COALESCE(SUM(bet_amount_minor), 0) FROM game_bet_records WHERE site_id = ? AND user_id = ? AND deleted_at IS NULL) AS betAmountMinor, (SELECT COALESCE(SUM(win_amount_minor), 0) FROM game_bet_records WHERE site_id = ? AND user_id = ? AND deleted_at IS NULL) AS winAmountMinor", append(append(append(append(append(scope, scope...), scope...), scope...), scope...), scope...)...).Scan(&summary).Error
	maskPlayerSupport(&player, actor.Sensitive)
	return map[string]any{"overview": player, "summary": summary, "wallets": wallets, "deposits": deposits, "withdrawals": withdrawals, "games": games, "promotions": promotions, "loginHistory": logins, "devices": devices, "risk": risk, "audit": audits, "tags": tags}, nil
}

func (s *Service) Player360ByNumber(playerNo string, actor Actor) (map[string]any, error) {
	player, err := s.supportPlayerByNumber(playerNo, actor)
	if err != nil {
		return nil, err
	}
	return s.Player360(player.ID, actor)
}

func (s *Service) PlayerTimeline(playerID uint64, q Query, actor Actor) (Result, error) {
	player, err := s.supportPlayer(playerID, actor)
	if err != nil {
		return Result{}, err
	}
	from, to, err := supportTimeRange(value(q, "from"), value(q, "to"))
	if err != nil {
		return Result{}, err
	}
	items := make([]TimelineItem, 0, 256)
	s.appendLoginTimeline(&items, player, from, to)
	s.appendPaymentTimeline(&items, player, from, to)
	s.appendWalletTimeline(&items, player, from, to)
	s.appendGameTimeline(&items, player, from, to)
	s.appendPromotionTimeline(&items, player, from, to)
	s.appendRiskAuditTimeline(&items, player, from, to)
	items = filterTimeline(items, q)
	sort.SliceStable(items, func(i, j int) bool { return items[i].OccurredAt.After(items[j].OccurredAt) })
	p, size := page(q)
	total := len(items)
	start := (p - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return Result{List: items[start:end], Total: int64(total), Page: p, PageSize: size}, nil
}

func (s *Service) PlayerTimelineByNumber(playerNo string, q Query, actor Actor) (Result, error) {
	player, err := s.supportPlayerByNumber(playerNo, actor)
	if err != nil {
		return Result{}, err
	}
	return s.PlayerTimeline(player.ID, q, actor)
}

func (s *Service) SearchOrders(q Query, actor Actor) (Result, error) {
	term := strings.TrimSpace(value(q, "query"))
	playerID := parseID(value(q, "playerID"))
	if playerNo := strings.TrimSpace(value(q, "playerNo")); playerNo != "" {
		player, lookupErr := s.supportPlayerByNumber(playerNo, actor)
		if lookupErr != nil {
			return Result{}, lookupErr
		}
		playerID = player.ID
	}
	if term == "" && playerID == 0 {
		return Result{}, NewError("SEARCH_TERM_REQUIRED", "order number or player is required")
	}
	if strings.ContainsAny(term, "%_") {
		return Result{}, NewError("INVALID_SEARCH_TERM", "wildcards are not allowed")
	}
	from, to, err := supportTimeRange(value(q, "from"), value(q, "to"))
	if err != nil {
		return Result{}, err
	}
	items := []map[string]any{}
	s.appendOrderRows(&items, "DEPOSIT", "game_deposit_orders", "order_no", "channel_order_no", q, actor, term, playerID, from, to)
	s.appendOrderRows(&items, "WITHDRAWAL", "game_withdraw_orders", "order_no", "channel_order_no", q, actor, term, playerID, from, to)
	s.appendOrderRows(&items, "WALLET", "game_wallet_transactions", "txn_no", "external_txn_no", q, actor, term, playerID, from, to)
	s.appendOrderRows(&items, "PROMOTION", "game_promotion_claims", "trigger_order_no", "trigger_event_id", q, actor, term, playerID, from, to)
	s.appendGameOrderRows(&items, q, actor, term, playerID, from, to)
	s.appendCallbackOrderRows(&items, q, actor, term, playerID, from, to)
	sort.SliceStable(items, func(i, j int) bool { return fmt.Sprint(items[i]["createdAt"]) > fmt.Sprint(items[j]["createdAt"]) })
	p, size := page(q)
	total := len(items)
	start := (p - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return Result{List: items[start:end], Total: int64(total), Page: p, PageSize: size}, nil
}

func (s *Service) SupportOrderDetail(kind string, id uint64, actor Actor) (map[string]any, error) {
	kind = strings.ToUpper(strings.TrimSpace(kind))
	if kind == "DEPOSIT" || kind == "WITHDRAWAL" {
		result, err := s.PaymentDiagnostics(kind, id, actor)
		if err != nil {
			return nil, err
		}
		result["statusTimeline"] = supportOrderStatusTimeline(result["order"])
		return result, nil
	}
	tables := map[string]string{
		"WALLET":            "game_wallet_transactions",
		"PROMOTION":         "game_promotion_claims",
		"GAME":              "game_bet_records",
		"PROVIDER_CALLBACK": "game_provider_callback_logs",
	}
	table := tables[kind]
	if table == "" {
		return nil, NewError("INVALID_ORDER_TYPE", "unsupported support order type")
	}
	var order map[string]any
	db := s.db.Table(table).Where("id=? AND deleted_at IS NULL", id)
	db, _ = operationalScope(db, actor, "")
	if err := db.Take(&order).Error; err != nil {
		return nil, rowNotFound("support_order")
	}
	siteID, _ := anyUint64(order["site_id"])
	playerID, _ := anyUint64(order["user_id"])
	identifiers := supportOrderIdentifiers(order)
	related := []map[string]any{}
	if playerID > 0 {
		relatedDB := s.db.Table("game_provider_transactions").Select("id,provider_code providerCode,provider_transaction_id providerTransactionID,provider_round_id providerRoundID,operation,amount_minor amountMinor,currency,status,created_at createdAt").Where("site_id = ? AND user_id = ? AND deleted_at IS NULL", siteID, playerID)
		if len(identifiers) > 0 {
			relatedDB = relatedDB.Where("provider_transaction_id IN ? OR provider_round_id IN ?", identifiers, identifiers)
		}
		_ = relatedDB.Order("id DESC").Limit(100).Find(&related).Error
	}
	events := []map[string]any{}
	_ = s.db.Table("event_outbox").Select("id,event_id eventID,event_type eventType,status,trace_id traceID,correlation_id correlationID,occurred_at occurredAt,last_error lastError").Where("site_id = ? AND (aggregate_id = ? OR correlation_id IN (?)) AND deleted_at IS NULL", siteID, fmt.Sprint(id), identifiers).Order("id DESC").Limit(100).Find(&events).Error
	audits := []map[string]any{}
	_ = s.db.Table("game_admin_operation_logs").Select("id,action,result,operator_username operator,reason,correlation_id requestID,created_at createdAt").Where("site_id = ? AND target_id = ? AND deleted_at IS NULL", siteID, id).Order("id DESC").Limit(100).Find(&audits).Error
	return map[string]any{
		"orderType": kind, "order": redact(order), "statusTimeline": supportOrderStatusTimeline(order),
		"relatedRecords": related, "events": events, "audit": audits,
	}, nil
}

func supportOrderIdentifiers(order map[string]any) []string {
	keys := []string{"order_no", "channel_order_no", "txn_no", "external_txn_no", "trigger_order_no", "trigger_event_id", "bet_no", "platform_bet_id", "provider_bet_transaction_id", "provider_win_transaction_id", "provider_transaction_id", "provider_round_id", "callback_id", "correlation_id"}
	values := make([]string, 0, len(keys))
	seen := map[string]bool{}
	for _, key := range keys {
		value := strings.TrimSpace(fmt.Sprint(order[key]))
		if value != "" && value != "<nil>" && !seen[value] {
			seen[value] = true
			values = append(values, value)
		}
	}
	if len(values) == 0 {
		values = append(values, "__no_related_identifier__")
	}
	return values
}

func supportOrderStatusTimeline(value any) []map[string]any {
	order, ok := value.(map[string]any)
	if !ok {
		return []map[string]any{}
	}
	stages := []struct{ key, status string }{
		{"created_at", "CREATED"}, {"requested_at", "REQUESTED"}, {"paid_at", "PAID"},
		{"approved_at", "APPROVED"}, {"settled_at", "SETTLED"}, {"completed_at", "COMPLETED"},
	}
	items := make([]map[string]any, 0, len(stages)+1)
	for _, stage := range stages {
		if at := order[stage.key]; at != nil && strings.TrimSpace(fmt.Sprint(at)) != "" && fmt.Sprint(at) != "<nil>" {
			items = append(items, map[string]any{"status": stage.status, "occurredAt": at})
		}
	}
	status := order["status"]
	if status == nil {
		status = order["bet_status"]
	}
	if status == nil {
		status = order["processing_result"]
	}
	if at := order["updated_at"]; status != nil && at != nil {
		items = append(items, map[string]any{"status": status, "occurredAt": at})
	}
	return items
}

func (s *Service) ListSupportCases(q Query, actor Actor) (Result, error) {
	rows := []map[string]any{}
	db := s.db.Table("game_support_cases c").Select("c.*").Where("c.deleted_at IS NULL")
	usersExist := s.db.Migrator().HasTable("game_users")
	adminsExist := s.db.Migrator().HasTable("game_admin_accounts")
	if usersExist {
		db = db.Select("c.*,u.user_no playerNo,u.username playerUsername").Joins("LEFT JOIN game_users AS u ON u.id = c.player_id AND u.deleted_at IS NULL")
	}
	if usersExist && adminsExist {
		db = db.Select("c.*,u.user_no playerNo,u.username playerUsername,a.username assignedAdminUsername").Joins("LEFT JOIN game_admin_accounts a ON a.id=c.assigned_admin_id AND a.deleted_at IS NULL")
	} else if adminsExist {
		db = db.Select("c.*,a.username assignedAdminUsername").Joins("LEFT JOIN game_admin_accounts a ON a.id=c.assigned_admin_id AND a.deleted_at IS NULL")
	}
	var err error
	db, err = operationalScopeAlias(db, "c", actor, value(q, "siteID"))
	if err != nil {
		return Result{}, err
	}
	db = exact(db, "c.status", strings.ToUpper(value(q, "status")))
	db = exact(db, "c.category", strings.ToUpper(value(q, "category")))
	db = exact(db, "c.priority", strings.ToUpper(value(q, "priority")))
	db = exact(db, "c.player_id", value(q, "playerID"))
	if playerNo := strings.TrimSpace(value(q, "playerNo")); playerNo != "" {
		if !s.db.Migrator().HasTable("game_users") {
			return Result{List: rows, Page: 1, PageSize: 20}, nil
		}
		db = db.Where("u.user_no=?", playerNo)
	}
	db = like(db, "c.case_no", value(q, "query"))
	return paginate(db.Order("c.updated_at DESC,c.id DESC"), q, &rows)
}

func (s *Service) SupportCaseDetail(id uint64, actor Actor) (map[string]any, error) {
	var row map[string]any
	db := s.db.Table("game_support_cases c").Select("c.*").Where("c.id=? AND c.deleted_at IS NULL", id)
	usersExist := s.db.Migrator().HasTable("game_users")
	if usersExist {
		db = db.Select("c.*,u.user_no playerNo,u.username playerUsername").Joins("LEFT JOIN game_users AS u ON u.id = c.player_id AND u.deleted_at IS NULL")
	}
	if usersExist && s.db.Migrator().HasTable("game_admin_accounts") {
		db = db.Select("c.*,u.user_no playerNo,u.username playerUsername,a.username assignedAdminUsername").Joins("LEFT JOIN game_admin_accounts a ON a.id=c.assigned_admin_id AND a.deleted_at IS NULL")
	}
	db, _ = operationalScopeAlias(db, "c", actor, "")
	if err := db.Take(&row).Error; err != nil {
		return nil, rowNotFound("support_case")
	}
	notes := []map[string]any{}
	_ = s.db.Table("game_support_case_notes").Select("id,author_id authorID,author_username authorUsername,note,created_at createdAt").Where("case_id = ? AND site_id = ? AND deleted_at IS NULL", id, row["site_id"]).Order("id ASC").Find(&notes).Error
	return map[string]any{"case": row, "notes": notes}, nil
}

func (s *Service) CreateSupportCase(input SupportCaseInput, actor Actor) (uint64, error) {
	siteID, err := operationalTargetScope(s.db, actor, input.SiteID)
	if err != nil {
		return 0, err
	}
	if input.PlayerID == 0 && strings.TrimSpace(input.PlayerNo) != "" {
		player, lookupErr := s.supportPlayerByNumber(input.PlayerNo, actor)
		if lookupErr != nil {
			return 0, lookupErr
		}
		input.PlayerID = player.ID
	}
	if input.PlayerID > 0 {
		if _, err := s.supportPlayerInScope(input.PlayerID, siteID); err != nil {
			return 0, err
		}
	}
	if err := validateSupportCase(&input, true); err != nil {
		return 0, err
	}
	key := strings.TrimSpace(input.IdempotencyKey)
	if key == "" {
		return 0, NewError("IDEMPOTENCY_KEY_REQUIRED", "idempotency key is required")
	}
	var id uint64
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_support_cases").Select("id").Where("created_by=? AND idempotency_key=?", actor.ID, key).Take(&id).Error; err == nil {
			return nil
		}
		caseNo := fmt.Sprintf("SC-%s-%06d", now().Format("20060102"), now().UnixNano()%1000000)
		row := map[string]any{"case_no": caseNo, "site_id": siteID, "player_id": input.PlayerID, "category": input.Category, "priority": input.Priority, "status": input.Status, "title": strings.TrimSpace(input.Title), "description": strings.TrimSpace(input.Description), "assigned_admin_id": input.AssignedAdminID, "created_by": actor.ID, "resolution": strings.TrimSpace(input.Resolution), "related_order_ids": strings.Join(cleanIdentifiers(input.RelatedOrderIDs), ","), "related_transaction_ids": strings.Join(cleanIdentifiers(input.RelatedTransactionIDs), ","), "idempotency_key": key, "created_at": now(), "updated_at": now()}
		if err := tx.Table("game_support_cases").Create(row).Error; err != nil {
			return duplicateError(err)
		}
		if err := tx.Table("game_support_cases").Select("id").Where("created_by=? AND idempotency_key=?", actor.ID, key).Take(&id).Error; err != nil {
			return err
		}
		return audit(tx, targetAuditActor(actor, siteID), "support:case:write", "support.case.create", "support_case", id, nil, row, input.Reason)
	})
	return id, err
}

func (s *Service) UpdateSupportCase(id uint64, input SupportCaseInput, actor Actor) error {
	if err := requireReason(input.Reason); err != nil {
		return err
	}
	if err := validateSupportCase(&input, false); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var before map[string]any
		db := tx.Table("game_support_cases").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id)
		db, _ = operationalScope(db, actor, "")
		if err := db.Take(&before).Error; err != nil {
			return rowNotFound("support_case")
		}
		siteID, _ := anyUint64(before["site_id"])
		updates := map[string]any{"category": input.Category, "priority": input.Priority, "status": input.Status, "title": strings.TrimSpace(input.Title), "description": strings.TrimSpace(input.Description), "assigned_admin_id": input.AssignedAdminID, "resolution": strings.TrimSpace(input.Resolution), "related_order_ids": strings.Join(cleanIdentifiers(input.RelatedOrderIDs), ","), "related_transaction_ids": strings.Join(cleanIdentifiers(input.RelatedTransactionIDs), ","), "updated_at": now()}
		if input.Status == "CLOSED" {
			updates["closed_at"] = now()
		} else {
			updates["closed_at"] = nil
		}
		if err := tx.Table("game_support_cases").Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, targetAuditActor(actor, siteID), "support:case:write", "support.case.update", "support_case", id, before, updates, input.Reason)
	})
}

func (s *Service) AddSupportCaseNote(id uint64, input SupportCaseNoteInput, actor Actor) (uint64, error) {
	if strings.TrimSpace(input.Note) == "" {
		return 0, NewError("NOTE_REQUIRED", "note is required")
	}
	if err := requireReason(input.Reason); err != nil {
		return 0, err
	}
	detail, err := s.SupportCaseDetail(id, actor)
	if err != nil {
		return 0, err
	}
	row := detail["case"].(map[string]any)
	siteID, _ := anyUint64(row["site_id"])
	var noteID uint64
	err = s.db.Transaction(func(tx *gorm.DB) error {
		note := map[string]any{"site_id": siteID, "case_id": id, "author_id": actor.ID, "author_username": actor.Username, "note": strings.TrimSpace(input.Note), "created_at": now(), "updated_at": now()}
		if err := tx.Table("game_support_case_notes").Create(note).Error; err != nil {
			return err
		}
		if err := tx.Table("game_support_case_notes").Select("id").Where("case_id=? AND author_id=?", id, actor.ID).Order("id DESC").Take(&noteID).Error; err != nil {
			return err
		}
		return audit(tx, targetAuditActor(actor, siteID), "support:case:write", "support.case.note", "support_case", id, nil, map[string]any{"noteID": noteID}, input.Reason)
	})
	return noteID, err
}

func (s *Service) supportPlayer(playerID uint64, actor Actor) (PlayerSupportRow, error) {
	var row PlayerSupportRow
	db := s.db.Table("game_users u").Select("u.id, u.user_no, u.username, u.phone_country_code, u.phone, u.email, u.site_id, s.site_name, u.status, u.country_code, u.vip_level, COALESCE(w.currency, '') AS currency, COALESCE(w.cash_balance, 0) AS available_balance, COALESCE(w.frozen_balance, 0) AS frozen_balance, COALESCE(w.bonus_balance, 0) AS bonus_balance, u.created_at AS registered_at, u.registered_ip, u.last_login_at, u.last_login_ip, COALESCE((SELECT wo.kyc_status FROM game_withdraw_orders AS wo WHERE wo.user_id = u.id AND wo.deleted_at IS NULL ORDER BY wo.id DESC LIMIT 1), '') AS kyc_status, COALESCE((SELECT rp.risk_status FROM game_user_risk_profiles AS rp WHERE rp.user_id = u.id AND rp.site_id = u.site_id AND rp.deleted_at IS NULL ORDER BY rp.id DESC LIMIT 1), 'NORMAL') AS risk_status").
		Joins("LEFT JOIN game_site_config s ON s.id=u.site_id AND s.deleted_at IS NULL").Joins("LEFT JOIN game_user_wallets AS w ON w.user_id = u.id AND w.site_id = u.site_id AND w.deleted_at IS NULL").Where("u.id=? AND u.deleted_at IS NULL", playerID)
	db, _ = operationalScopeAlias(db, "u", actor, "")
	if err := db.Take(&row).Error; err != nil {
		return row, rowNotFound("player")
	}
	return row, nil
}

func (s *Service) supportPlayerByNumber(playerNo string, actor Actor) (PlayerSupportRow, error) {
	var id uint64
	db := s.db.Table("game_users u").Select("u.id").Where("u.user_no=? AND u.deleted_at IS NULL", strings.TrimSpace(playerNo))
	db, _ = operationalScopeAlias(db, "u", actor, "")
	if err := db.Take(&id).Error; err != nil {
		return PlayerSupportRow{}, rowNotFound("player")
	}
	return s.supportPlayer(id, actor)
}

func (s *Service) supportPlayerInScope(playerID uint64, siteID uint64) (uint64, error) {
	var id uint64
	db := s.db.Table("game_users").Select("id").Where("id = ? AND deleted_at IS NULL", playerID)
	if siteID > 0 {
		db = db.Where("site_id=?", siteID)
	}
	if err := db.Take(&id).Error; err != nil {
		return 0, rowNotFound("player")
	}
	return id, nil
}

func maskPlayerSupport(row *PlayerSupportRow, sensitive bool) {
	if sensitive {
		return
	}
	row.Phone = maskTail(row.Phone, 4)
	row.Email = maskEmail(row.Email)
	row.RegisteredIP = maskIP(row.RegisteredIP)
	row.LastLoginIP = maskIP(row.LastLoginIP)
}

func maskTail(value string, visible int) string {
	value = strings.TrimSpace(value)
	if len(value) <= visible {
		return strings.Repeat("*", len(value))
	}
	return strings.Repeat("*", len(value)-visible) + value[len(value)-visible:]
}

func maskEmail(value string) string {
	parts := strings.Split(strings.TrimSpace(value), "@")
	if len(parts) != 2 {
		return maskTail(value, 2)
	}
	name := parts[0]
	if len(name) > 1 {
		name = name[:1] + strings.Repeat("*", len(name)-1)
	}
	return name + "@" + parts[1]
}

func maskIP(value string) string {
	parts := strings.Split(value, ".")
	if len(parts) == 4 {
		return parts[0] + "." + parts[1] + ".*.*"
	}
	return maskTail(value, 4)
}

func supportTimeRange(fromValue string, toValue string) (time.Time, time.Time, error) {
	to := now()
	from := to.Add(-30 * 24 * time.Hour)
	var err error
	if strings.TrimSpace(toValue) != "" {
		to, err = parseSupportTime(toValue, true)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if strings.TrimSpace(fromValue) != "" {
		from, err = parseSupportTime(fromValue, false)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if !to.After(from) {
		return time.Time{}, time.Time{}, NewError("INVALID_TIME_RANGE", "end time must be after start time")
	}
	if to.Sub(from) > 30*24*time.Hour+time.Second {
		return time.Time{}, time.Time{}, NewError("TIME_RANGE_TOO_LARGE", "maximum synchronous query range is 30 days")
	}
	return from, to, nil
}

func parseSupportTime(value string, endOfDay bool) (time.Time, error) {
	value = strings.TrimSpace(value)
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return time.Time{}, NewError("INVALID_TIME", "time must be RFC3339 or YYYY-MM-DD")
	}
	if endOfDay {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return parsed, nil
}

func filterTimeline(items []TimelineItem, q Query) []TimelineItem {
	types := map[string]bool{}
	for _, item := range strings.Split(strings.ToUpper(value(q, "eventTypes")), ",") {
		if strings.TrimSpace(item) != "" {
			types[strings.TrimSpace(item)] = true
		}
	}
	status := strings.ToUpper(value(q, "status"))
	provider := strings.ToUpper(value(q, "providerCode"))
	minAmount, minOK := decimal.NewFromString(value(q, "minAmount"))
	maxAmount, maxOK := decimal.NewFromString(value(q, "maxAmount"))
	result := make([]TimelineItem, 0, len(items))
	for _, item := range items {
		if len(types) > 0 && !types[item.EventType] {
			continue
		}
		if status != "" && strings.ToUpper(item.Status) != status {
			continue
		}
		if provider != "" && strings.ToUpper(item.ProviderCode) != provider {
			continue
		}
		if minOK == nil && item.Amount != nil && item.Amount.LessThan(minAmount) {
			continue
		}
		if maxOK == nil && item.Amount != nil && item.Amount.GreaterThan(maxAmount) {
			continue
		}
		result = append(result, item)
	}
	return result
}

func cleanIdentifiers(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && len(value) <= 128 && !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}
	return result
}

func validateSupportCase(input *SupportCaseInput, create bool) error {
	input.Category = strings.ToUpper(defaultString(input.Category, "OTHER"))
	input.Priority = strings.ToUpper(defaultString(input.Priority, "NORMAL"))
	input.Status = strings.ToUpper(defaultString(input.Status, "OPEN"))
	if !map[string]bool{"ACCOUNT": true, "DEPOSIT": true, "WITHDRAWAL": true, "WALLET": true, "GAME": true, "PROMOTION": true, "TECHNICAL": true, "OTHER": true}[input.Category] {
		return NewError("INVALID_CASE_CATEGORY", "invalid case category")
	}
	if !map[string]bool{"LOW": true, "NORMAL": true, "HIGH": true, "URGENT": true}[input.Priority] {
		return NewError("INVALID_CASE_PRIORITY", "invalid case priority")
	}
	if !map[string]bool{"OPEN": true, "INVESTIGATING": true, "WAITING_CUSTOMER": true, "RESOLVED": true, "CLOSED": true}[input.Status] {
		return NewError("INVALID_CASE_STATUS", "invalid case status")
	}
	if strings.TrimSpace(input.Title) == "" {
		return NewError("CASE_TITLE_REQUIRED", "case title is required")
	}
	if create && strings.TrimSpace(input.Reason) == "" {
		input.Reason = "support case created"
	}
	return nil
}

func (s *Service) appendLoginTimeline(items *[]TimelineItem, player PlayerSupportRow, from time.Time, to time.Time) {
	if !player.RegisteredAt.Before(from) && !player.RegisteredAt.After(to) {
		*items = append(*items, TimelineItem{ID: fmt.Sprintf("register:%d", player.ID), EventType: "REGISTER", OccurredAt: player.RegisteredAt, Status: "SUCCESS", Description: "Player account registered", DetailType: "PLAYER", DetailID: player.ID})
	}
	rows := []struct {
		ID         uint64
		Success    int
		IPAddress  string
		DeviceID   string
		OccurredAt time.Time
	}{}
	_ = s.db.Table("game_user_login_records").Select("id,success,ip_address,device_id,occurred_at").
		Where("site_id = ? AND user_id = ? AND occurred_at BETWEEN ? AND ? AND deleted_at IS NULL", player.SiteID, player.ID, from, to).
		Order("occurred_at DESC").Limit(500).Find(&rows).Error
	for _, row := range rows {
		eventType, status, description := "LOGIN_FAILED", "FAILED", "Login failed"
		if row.Success == 1 {
			eventType, status, description = "LOGIN_SUCCESS", "SUCCESS", "Login succeeded"
		}
		*items = append(*items, TimelineItem{ID: fmt.Sprintf("login:%d", row.ID), EventType: eventType, OccurredAt: row.OccurredAt, Status: status, Description: description, DetailType: "LOGIN", DetailID: row.ID})
	}
}

func (s *Service) appendPaymentTimeline(items *[]TimelineItem, player PlayerSupportRow, from time.Time, to time.Time) {
	deposits := []struct {
		ID             uint64
		OrderNo        string
		ChannelOrderNo string
		Amount         *decimal.Decimal
		Currency       string
		Status         int
		CreatedAt      time.Time
		PaidAt         *time.Time
		CompletedAt    *time.Time
	}{}
	_ = s.db.Table("game_deposit_orders").Select("id,order_no,channel_order_no,amount,currency,status,created_at,paid_at,completed_at").
		Where("site_id = ? AND user_id = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", player.SiteID, player.ID, from, to).
		Order("id DESC").Limit(500).Find(&deposits).Error
	for _, row := range deposits {
		*items = append(*items, TimelineItem{ID: fmt.Sprintf("deposit-created:%d", row.ID), EventType: "DEPOSIT_CREATED", OccurredAt: row.CreatedAt, Amount: row.Amount, Currency: row.Currency, BusinessOrderNo: row.OrderNo, ExternalOrderNo: row.ChannelOrderNo, Status: depositStatus(row.Status), Description: "Deposit order created", DetailType: "DEPOSIT", DetailID: row.ID})
		at := row.CompletedAt
		if row.PaidAt != nil {
			at = row.PaidAt
		}
		if at != nil && !at.Before(from) && !at.After(to) {
			eventType, description := "DEPOSIT_FAILED", "Deposit failed"
			if row.Status == 2 {
				eventType, description = "DEPOSIT_PAID", "Deposit paid"
			}
			*items = append(*items, TimelineItem{ID: fmt.Sprintf("deposit-result:%d", row.ID), EventType: eventType, OccurredAt: *at, Amount: row.Amount, Currency: row.Currency, BusinessOrderNo: row.OrderNo, ExternalOrderNo: row.ChannelOrderNo, Status: depositStatus(row.Status), Description: description, DetailType: "DEPOSIT", DetailID: row.ID})
		}
	}

	withdrawals := []struct {
		ID             uint64
		OrderNo        string
		ChannelOrderNo string
		Amount         *decimal.Decimal
		Currency       string
		Status         int
		AuditStatus    int
		CreatedAt      time.Time
		ReviewedAt     *time.Time
		CompletedAt    *time.Time
	}{}
	_ = s.db.Table("game_withdraw_orders").Select("id,order_no,channel_order_no,amount,currency,status,audit_status,created_at,reviewed_at,completed_at").
		Where("site_id = ? AND user_id = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", player.SiteID, player.ID, from, to).
		Order("id DESC").Limit(500).Find(&withdrawals).Error
	for _, row := range withdrawals {
		*items = append(*items, TimelineItem{ID: fmt.Sprintf("withdraw-created:%d", row.ID), EventType: "WITHDRAW_CREATED", OccurredAt: row.CreatedAt, Amount: row.Amount, Currency: row.Currency, BusinessOrderNo: row.OrderNo, ExternalOrderNo: row.ChannelOrderNo, Status: withdrawStatus(row.Status), Description: "Withdrawal order created", DetailType: "WITHDRAWAL", DetailID: row.ID})
		if row.ReviewedAt != nil && !row.ReviewedAt.Before(from) && !row.ReviewedAt.After(to) {
			eventType, description := "WITHDRAW_APPROVED", "Withdrawal approved"
			if row.AuditStatus == 3 {
				eventType, description = "WITHDRAW_REJECTED", "Withdrawal rejected"
			}
			*items = append(*items, TimelineItem{ID: fmt.Sprintf("withdraw-review:%d", row.ID), EventType: eventType, OccurredAt: *row.ReviewedAt, Amount: row.Amount, Currency: row.Currency, BusinessOrderNo: row.OrderNo, ExternalOrderNo: row.ChannelOrderNo, Status: withdrawStatus(row.Status), Description: description, DetailType: "WITHDRAWAL", DetailID: row.ID})
		}
		if row.CompletedAt != nil && !row.CompletedAt.Before(from) && !row.CompletedAt.After(to) {
			eventType, description := "WITHDRAW_REJECTED", "Withdrawal failed"
			if row.Status == 2 {
				eventType, description = "WITHDRAW_PAID", "Withdrawal paid"
			}
			*items = append(*items, TimelineItem{ID: fmt.Sprintf("withdraw-result:%d", row.ID), EventType: eventType, OccurredAt: *row.CompletedAt, Amount: row.Amount, Currency: row.Currency, BusinessOrderNo: row.OrderNo, ExternalOrderNo: row.ChannelOrderNo, Status: withdrawStatus(row.Status), Description: description, DetailType: "WITHDRAWAL", DetailID: row.ID})
		}
	}
}

func (s *Service) appendWalletTimeline(items *[]TimelineItem, player PlayerSupportRow, from time.Time, to time.Time) {
	rows := []struct {
		ID               uint64
		TxnNo            string
		ExternalTxnNo    string
		Direction        string
		Category         string
		ProviderCode     string
		Amount           *decimal.Decimal
		Currency         string
		Status           int
		BalanceBefore    *decimal.Decimal
		BalanceAfter     *decimal.Decimal
		OperatorUsername string
		CreatedAt        time.Time
	}{}
	_ = s.db.Table("game_wallet_transactions").Select("id,txn_no,external_txn_no,direction,category,provider_code,amount,currency,status,balance_before,balance_after,operator_username,created_at").
		Where("site_id = ? AND user_id = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", player.SiteID, player.ID, from, to).
		Order("id DESC").Limit(750).Find(&rows).Error
	for _, row := range rows {
		eventType := "WALLET_DEBIT"
		if strings.EqualFold(row.Direction, "CREDIT") {
			eventType = "WALLET_CREDIT"
		}
		category := strings.ToUpper(row.Category)
		if strings.Contains(category, "UNFREEZE") {
			eventType = "WALLET_UNFREEZE"
		} else if strings.Contains(category, "FREEZE") {
			eventType = "WALLET_FREEZE"
		}
		*items = append(*items, TimelineItem{ID: fmt.Sprintf("wallet:%d", row.ID), EventType: eventType, OccurredAt: row.CreatedAt, Amount: row.Amount, Currency: row.Currency, BalanceBefore: row.BalanceBefore, BalanceAfter: row.BalanceAfter, BusinessOrderNo: row.TxnNo, ExternalOrderNo: row.ExternalTxnNo, ProviderCode: row.ProviderCode, Status: numericStatus(row.Status), Operator: row.OperatorUsername, Description: row.Category, DetailType: "WALLET", DetailID: row.ID})
	}
}

func (s *Service) appendGameTimeline(items *[]TimelineItem, player PlayerSupportRow, from time.Time, to time.Time) {
	launches := []struct {
		ID            uint64
		LaunchNo      string
		LaunchStatus  int
		CorrelationID string
		LaunchTime    *time.Time
		CreatedAt     time.Time
	}{}
	_ = s.db.Table("game_game_launch_logs").Select("id,launch_no,launch_status,correlation_id,launch_time,created_at").Where("site_id = ? AND user_id = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", player.SiteID, player.ID, from, to).Order("id DESC").Limit(500).Find(&launches).Error
	for _, row := range launches {
		at := row.CreatedAt
		if row.LaunchTime != nil {
			at = *row.LaunchTime
		}
		*items = append(*items, TimelineItem{ID: fmt.Sprintf("launch:%d", row.ID), EventType: "GAME_LAUNCH", OccurredAt: at, BusinessOrderNo: row.LaunchNo, Status: numericStatus(row.LaunchStatus), RequestID: row.CorrelationID, Description: "Game launch", DetailType: "GAME_LAUNCH", DetailID: row.ID})
	}
	transactions := []struct {
		ID                                                                                uint64
		ProviderCode, ProviderTransactionID, ProviderRoundID, Operation, Currency, Status string
		AmountMinor, BalanceBeforeMinor, BalanceAfterMinor                                int64
		CreatedAt                                                                         time.Time
	}{}
	_ = s.db.Table("game_provider_transactions").Select("id,provider_code,provider_transaction_id,provider_round_id,operation,amount_minor,currency,status,balance_before_minor,balance_after_minor,created_at").Where("site_id = ? AND user_id = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", player.SiteID, player.ID, from, to).Order("id DESC").Limit(750).Find(&transactions).Error
	for _, row := range transactions {
		eventType := providerTimelineEvent(row.Operation)
		amount := decimal.NewFromInt(row.AmountMinor).Shift(-2)
		before := decimal.NewFromInt(row.BalanceBeforeMinor).Shift(-2)
		after := decimal.NewFromInt(row.BalanceAfterMinor).Shift(-2)
		*items = append(*items, TimelineItem{ID: fmt.Sprintf("provider:%d", row.ID), EventType: eventType, OccurredAt: row.CreatedAt, Amount: &amount, Currency: row.Currency, BalanceBefore: &before, BalanceAfter: &after, BusinessOrderNo: row.ProviderRoundID, ExternalOrderNo: row.ProviderTransactionID, ProviderCode: row.ProviderCode, Status: row.Status, Description: strings.Title(strings.ToLower(row.Operation)), DetailType: "PROVIDER_TRANSACTION", DetailID: row.ID})
	}
}

func (s *Service) appendPromotionTimeline(items *[]TimelineItem, player PlayerSupportRow, from time.Time, to time.Time) {
	rows := []struct {
		ID                                        uint64
		TriggerOrderNo, Currency, Status, TraceID string
		RewardAmount                              *decimal.Decimal
		CreatedAt                                 time.Time
	}{}
	_ = s.db.Table("game_promotion_claims").Select("id,trigger_order_no,currency,reward_amount,status,trace_id,created_at").Where("site_id = ? AND user_id = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", player.SiteID, player.ID, from, to).Order("id DESC").Limit(500).Find(&rows).Error
	for _, row := range rows {
		*items = append(*items, TimelineItem{ID: fmt.Sprintf("promotion:%d", row.ID), EventType: "PROMOTION_CLAIMED", OccurredAt: row.CreatedAt, Amount: row.RewardAmount, Currency: row.Currency, BusinessOrderNo: row.TriggerOrderNo, Status: row.Status, RequestID: row.TraceID, Description: "Promotion claimed", DetailType: "PROMOTION", DetailID: row.ID})
	}
}

func (s *Service) appendRiskAuditTimeline(items *[]TimelineItem, player PlayerSupportRow, from time.Time, to time.Time) {
	risk := []struct {
		ID                                                  uint64
		EventType, RiskLevel, Description, OperatorUsername string
		CreatedAt                                           time.Time
	}{}
	_ = s.db.Table("game_user_risk_events").Select("id,event_type,risk_level,description,operator_username,created_at").Where("site_id = ? AND user_id = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", player.SiteID, player.ID, from, to).Order("id DESC").Limit(500).Find(&risk).Error
	for _, row := range risk {
		*items = append(*items, TimelineItem{ID: fmt.Sprintf("risk:%d", row.ID), EventType: "RISK_ACTION", OccurredAt: row.CreatedAt, Status: row.RiskLevel, Operator: row.OperatorUsername, Description: row.EventType + ": " + row.Description, DetailType: "RISK", DetailID: row.ID})
	}
	audits := []struct {
		ID                                                      uint64
		Action, Result, OperatorUsername, Reason, CorrelationID string
		CreatedAt                                               time.Time
	}{}
	_ = s.db.Table("game_admin_operation_logs").Select("id,action,result,operator_username,reason,correlation_id,created_at").Where("site_id = ? AND target_user_id = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", player.SiteID, player.ID, from, to).Order("id DESC").Limit(500).Find(&audits).Error
	for _, row := range audits {
		*items = append(*items, TimelineItem{ID: fmt.Sprintf("audit:%d", row.ID), EventType: "ADMIN_ACTION", OccurredAt: row.CreatedAt, Status: row.Result, RequestID: row.CorrelationID, Operator: row.OperatorUsername, Description: row.Action + ": " + row.Reason, DetailType: "AUDIT", DetailID: row.ID})
	}
}

func (s *Service) appendOrderRows(items *[]map[string]any, kind string, table string, internalColumn string, externalColumn string, q Query, actor Actor, term string, playerID uint64, from time.Time, to time.Time) {
	selects := map[string]string{
		"DEPOSIT":    "o.id, o.site_id AS siteID, o.user_id AS playerID, o.order_no AS orderNo, o.channel_order_no AS externalOrderNo, o.amount, o.currency, o.status, o.channel_code AS providerCode, o.created_at AS createdAt",
		"WITHDRAWAL": "o.id, o.site_id AS siteID, o.user_id AS playerID, o.order_no AS orderNo, o.channel_order_no AS externalOrderNo, o.amount, o.currency, o.status, o.payment_method AS providerCode, o.created_at AS createdAt",
		"WALLET":     "o.id, o.site_id AS siteID, o.user_id AS playerID, o.txn_no AS orderNo, o.external_txn_no AS externalOrderNo, o.amount, o.currency, o.status, o.provider_code AS providerCode, o.created_at AS createdAt",
		"PROMOTION":  "o.id, o.site_id AS siteID, o.user_id AS playerID, o.trigger_order_no AS orderNo, o.trigger_event_id AS externalOrderNo, o.reward_amount AS amount, o.currency, o.status, '' AS providerCode, o.created_at AS createdAt",
	}
	rows := []map[string]any{}
	selectSQL := selects[kind]
	db := s.db.Table(table + " o")
	if s.db.Migrator().HasTable("game_users") {
		selectSQL += ",u.user_no playerNo"
		db = db.Joins("LEFT JOIN game_users AS u ON u.id = o.user_id AND u.site_id = o.site_id AND u.deleted_at IS NULL")
	}
	db = db.Select(selectSQL).Where("o.deleted_at IS NULL AND o.created_at BETWEEN ? AND ?", from, to)
	db, _ = operationalScopeAlias(db, "o", actor, value(q, "siteID"))
	if term != "" {
		db = db.Where("o."+internalColumn+"=? OR o."+externalColumn+"=?", term, term)
	}
	if playerID > 0 {
		db = db.Where("o.user_id=?", playerID)
	}
	if status := strings.TrimSpace(value(q, "status")); status != "" {
		db = db.Where("o.status=?", supportOrderStatusValue(kind, status))
	}
	_ = db.Order("o.created_at DESC").Limit(500).Find(&rows).Error
	for _, row := range rows {
		row["orderType"] = kind
		if kind == "DEPOSIT" {
			row["status"] = depositStatus(intValue(row["status"]))
		} else if kind == "WITHDRAWAL" {
			row["status"] = withdrawStatus(intValue(row["status"]))
		} else if kind == "WALLET" {
			row["status"] = numericStatus(intValue(row["status"]))
		}
		*items = append(*items, row)
	}
}

func supportOrderStatusValue(kind string, status string) any {
	status = strings.ToUpper(strings.TrimSpace(status))
	if _, err := strconv.Atoi(status); err == nil {
		return status
	}
	if kind == "PROMOTION" {
		return status
	}
	common := map[string]int{"PENDING": 1, "SUCCESS": 2, "FAILED": 3}
	if value, ok := common[status]; ok {
		return value
	}
	if kind == "DEPOSIT" {
		if status == "EXPIRED" {
			return 4
		}
		if status == "CANCELLED" {
			return 5
		}
	}
	if kind == "WITHDRAWAL" {
		if status == "RETURNED" {
			return 4
		}
		if status == "CANCELLED" {
			return 5
		}
		if status == "REVIEWING" || status == "PROCESSING" {
			return 6
		}
	}
	return status
}

func intValue(value any) int {
	parsed, _ := strconv.Atoi(fmt.Sprint(value))
	return parsed
}

func (s *Service) appendGameOrderRows(items *[]map[string]any, q Query, actor Actor, term string, playerID uint64, from time.Time, to time.Time) {
	rows := []map[string]any{}
	selectSQL := "o.id, o.site_id AS siteID, o.user_id AS playerID, o.bet_no AS orderNo, o.provider_bet_transaction_id AS externalOrderNo, o.bet_amount_minor AS amountMinor, o.currency, o.bet_status AS status, o.provider_code AS providerCode, o.provider_round_id AS roundID, o.created_at AS createdAt"
	db := s.db.Table("game_bet_records o")
	if s.db.Migrator().HasTable("game_users") {
		selectSQL += ",u.user_no playerNo"
		db = db.Joins("LEFT JOIN game_users AS u ON u.id = o.user_id AND u.site_id = o.site_id AND u.deleted_at IS NULL")
	}
	db = db.Select(selectSQL).Where("o.deleted_at IS NULL AND o.created_at BETWEEN ? AND ?", from, to)
	db, _ = operationalScopeAlias(db, "o", actor, value(q, "siteID"))
	if term != "" {
		db = db.Where("o.bet_no=? OR o.platform_bet_id=? OR o.provider_bet_transaction_id=? OR o.provider_win_transaction_id=? OR o.provider_round_id=?", term, term, term, term, term)
	}
	if playerID > 0 {
		db = db.Where("o.user_id=?", playerID)
	}
	if status := strings.TrimSpace(value(q, "status")); status != "" {
		db = db.Where("o.bet_status=?", status)
	}
	_ = db.Order("o.created_at DESC").Limit(500).Find(&rows).Error
	for _, row := range rows {
		row["orderType"] = "GAME"
		*items = append(*items, row)
	}
}

func (s *Service) appendCallbackOrderRows(items *[]map[string]any, q Query, actor Actor, term string, playerID uint64, from time.Time, to time.Time) {
	rows := []map[string]any{}
	selectSQL := "o.id, o.site_id AS siteID, o.user_id AS playerID, o.callback_id AS orderNo, o.provider_transaction_id AS externalOrderNo, o.amount_minor AS amountMinor, o.currency, o.processing_result AS status, o.provider_code AS providerCode, o.provider_round_id AS roundID, o.duration_ms AS durationMS, o.created_at AS createdAt"
	db := s.db.Table("game_provider_callback_logs o")
	if s.db.Migrator().HasTable("game_users") {
		selectSQL += ",u.user_no playerNo"
		db = db.Joins("LEFT JOIN game_users AS u ON u.id = o.user_id AND u.site_id = o.site_id AND u.deleted_at IS NULL")
	}
	db = db.Select(selectSQL).Where("o.deleted_at IS NULL AND o.created_at BETWEEN ? AND ?", from, to)
	db, _ = operationalScopeAlias(db, "o", actor, value(q, "siteID"))
	if term != "" {
		db = db.Where("o.callback_id=? OR o.provider_transaction_id=? OR o.provider_round_id=?", term, term, term)
	}
	if playerID > 0 {
		db = db.Where("o.user_id=?", playerID)
	}
	if status := strings.TrimSpace(value(q, "status")); status != "" {
		db = db.Where("o.processing_result=?", status)
	}
	_ = db.Order("o.created_at DESC").Limit(500).Find(&rows).Error
	for _, row := range rows {
		row["orderType"] = "PROVIDER_CALLBACK"
		*items = append(*items, row)
	}
}

func providerTimelineEvent(operation string) string {
	switch strings.ToUpper(operation) {
	case "BET":
		return "GAME_BET"
	case "WIN":
		return "GAME_WIN"
	case "REFUND":
		return "GAME_REFUND"
	case "ROLLBACK", "CANCEL":
		return "GAME_ROLLBACK"
	default:
		return "GAME_" + strings.ToUpper(operation)
	}
}

func depositStatus(status int) string {
	switch status {
	case 1:
		return "PENDING"
	case 2:
		return "SUCCESS"
	case 3:
		return "FAILED"
	case 4:
		return "EXPIRED"
	case 5:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

func withdrawStatus(status int) string {
	switch status {
	case 1:
		return "PENDING"
	case 2:
		return "SUCCESS"
	case 3:
		return "FAILED"
	case 4:
		return "RETURNED"
	case 5:
		return "CANCELLED"
	case 6:
		return "REVIEWING"
	default:
		return "UNKNOWN"
	}
}

func numericStatus(status int) string {
	if status == 1 {
		return "PENDING"
	}
	if status == 2 {
		return "SUCCESS"
	}
	if status == 3 {
		return "FAILED"
	}
	return strconv.Itoa(status)
}
