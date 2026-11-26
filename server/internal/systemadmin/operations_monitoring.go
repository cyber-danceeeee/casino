package systemadmin

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentMetric struct {
	Count        int64            `json:"count"`
	Amount       *decimal.Decimal `json:"amount"`
	SuccessCount int64            `json:"successCount"`
	FailedCount  int64            `json:"failedCount"`
	PendingCount int64            `json:"pendingCount"`
	TimeoutCount int64            `json:"timeoutCount"`
	SuccessRate  float64          `json:"successRate"`
}

type PaymentHealthRow struct {
	Kind                      string     `json:"kind"`
	ProviderCode              string     `json:"providerCode"`
	ChannelCode               string     `json:"channelCode"`
	MethodCode                string     `json:"methodCode"`
	TotalRequests             int64      `json:"totalRequests"`
	SuccessCount              int64      `json:"successCount"`
	FailedCount               int64      `json:"failedCount"`
	PendingCount              int64      `json:"pendingCount"`
	TimeoutCount              int64      `json:"timeoutCount"`
	SuccessRate               float64    `json:"successRate"`
	AverageLatencyMS          *float64   `json:"averageLatencyMS"`
	P95LatencyMS              *float64   `json:"p95LatencyMS"`
	LastSuccessfulTransaction *time.Time `json:"lastSuccessfulTransaction"`
	LastFailedTransaction     *time.Time `json:"lastFailedTransaction"`
	OperationalStatus         string     `json:"operationalStatus"`
}

type OperationalNoteInput struct {
	TenantID   uint64 `json:"tenantID"`
	SiteID     uint64 `json:"siteID"`
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetID"`
	Note       string `json:"note"`
	Reason     string `json:"reason"`
}

func (s *Service) PaymentMonitor(q Query, actor Actor) (map[string]any, error) {
	from, to, err := supportTimeRange(value(q, "from"), value(q, "to"))
	if err != nil {
		return nil, err
	}
	deposit, err := s.paymentMetric("game_deposit_orders", true, q, actor, from, to)
	if err != nil {
		return nil, err
	}
	withdrawal, err := s.paymentMetric("game_withdraw_orders", false, q, actor, from, to)
	if err != nil {
		return nil, err
	}
	health, err := s.paymentHealth(q, actor, from, to)
	if err != nil {
		return nil, err
	}
	return map[string]any{"from": from, "to": to, "deposit": deposit, "withdrawal": withdrawal, "health": health}, nil
}

func (s *Service) paymentMetric(table string, deposit bool, q Query, actor Actor, from, to time.Time) (PaymentMetric, error) {
	row := struct {
		Count        int64
		Amount       *decimal.Decimal
		SuccessCount int64
		FailedCount  int64
		PendingCount int64
		TimeoutCount int64
	}{}
	amountColumn := "actual_amount"
	timeoutExpression := "status IN (1,6) AND created_at < ?"
	if deposit {
		amountColumn = "amount"
		timeoutExpression = "status=1 AND expired_at IS NOT NULL AND expired_at < ?"
	}
	db := s.db.Table(table+" o").Select(fmt.Sprintf(`COUNT(*) count,COALESCE(SUM(%s),0) amount,
		SUM(CASE WHEN status=2 THEN 1 ELSE 0 END) success_count,
		SUM(CASE WHEN status IN (3,4,5) THEN 1 ELSE 0 END) failed_count,
		SUM(CASE WHEN status IN (1,6) THEN 1 ELSE 0 END) pending_count,
		SUM(CASE WHEN %s THEN 1 ELSE 0 END) timeout_count`, amountColumn, timeoutExpression), now().Add(-30*time.Minute)).
		Where("o.created_at BETWEEN ? AND ? AND o.deleted_at IS NULL", from, to)
	db, _ = operationalScopeAlias(db, "o", actor, value(q, "tenantID"), value(q, "siteID"))
	db = paymentFilters(db, q, deposit)
	if err := db.Scan(&row).Error; err != nil {
		return PaymentMetric{}, err
	}
	result := PaymentMetric{Count: row.Count, Amount: row.Amount, SuccessCount: row.SuccessCount, FailedCount: row.FailedCount, PendingCount: row.PendingCount, TimeoutCount: row.TimeoutCount}
	if row.Count > 0 {
		result.SuccessRate = float64(row.SuccessCount) * 100 / float64(row.Count)
	}
	return result, nil
}

func paymentFilters(db *gorm.DB, q Query, deposit bool) *gorm.DB {
	if value(q, "currency") != "" {
		db = db.Where("o.currency=?", strings.ToUpper(value(q, "currency")))
	}
	if value(q, "status") != "" {
		db = db.Where("o.status=?", value(q, "status"))
	}
	if deposit {
		if value(q, "channelCode") != "" {
			db = db.Where("o.channel_code=?", strings.ToUpper(value(q, "channelCode")))
		}
		if value(q, "methodCode") != "" {
			db = db.Where("o.method_code=?", strings.ToUpper(value(q, "methodCode")))
		}
		if value(q, "providerCode") != "" {
			db = db.Where("o.channel_code=?", strings.ToUpper(value(q, "providerCode")))
		}
	} else if value(q, "methodCode") != "" {
		db = db.Where("o.payment_method=?", strings.ToUpper(value(q, "methodCode")))
	}
	if value(q, "country") != "" {
		db = db.Where("EXISTS (SELECT 1 FROM game_users pu WHERE pu.id=o.user_id AND pu.tenant_id=o.tenant_id AND pu.country_code=? AND pu.deleted_at IS NULL)", strings.ToUpper(value(q, "country")))
	}
	return db
}

func (s *Service) paymentHealth(q Query, actor Actor, from, to time.Time) ([]PaymentHealthRow, error) {
	deposits := []struct {
		ID                             uint64
		ChannelCode, MethodCode        string
		Status                         int
		CreatedAt                      time.Time
		CompletedAt, PaidAt, ExpiredAt *time.Time
	}{}
	db := s.db.Table("game_deposit_orders o").Select("id,channel_code,method_code,status,created_at,completed_at,paid_at,expired_at").Where("o.created_at BETWEEN ? AND ? AND o.deleted_at IS NULL", from, to)
	db, _ = operationalScopeAlias(db, "o", actor, value(q, "tenantID"), value(q, "siteID"))
	db = paymentFilters(db, q, true)
	if err := db.Order("o.id DESC").Limit(20000).Find(&deposits).Error; err != nil {
		return nil, err
	}
	withdrawals := []struct {
		ID            uint64
		PaymentMethod string
		Status        int
		CreatedAt     time.Time
		CompletedAt   *time.Time
	}{}
	wdb := s.db.Table("game_withdraw_orders o").Select("id,payment_method,status,created_at,completed_at").Where("o.created_at BETWEEN ? AND ? AND o.deleted_at IS NULL", from, to)
	wdb, _ = operationalScopeAlias(wdb, "o", actor, value(q, "tenantID"), value(q, "siteID"))
	wdb = paymentFilters(wdb, q, false)
	if err := wdb.Order("o.id DESC").Limit(20000).Find(&withdrawals).Error; err != nil {
		return nil, err
	}
	type accumulator struct {
		row        PaymentHealthRow
		latencies  []float64
		configured bool
	}
	groups := map[string]*accumulator{}
	for _, item := range deposits {
		key := "DEPOSIT:" + item.ChannelCode + ":" + item.MethodCode
		group := groups[key]
		if group == nil {
			group = &accumulator{row: PaymentHealthRow{Kind: "DEPOSIT", ProviderCode: item.ChannelCode, ChannelCode: item.ChannelCode, MethodCode: item.MethodCode}}
			groups[key] = group
		}
		group.row.TotalRequests++
		end := item.CompletedAt
		if item.PaidAt != nil {
			end = item.PaidAt
		}
		classifyHealthRecord(&group.row, item.Status, item.CreatedAt, end, item.Status == 1 && item.ExpiredAt != nil && item.ExpiredAt.Before(now()))
		if end != nil {
			group.latencies = append(group.latencies, float64(end.Sub(item.CreatedAt).Milliseconds()))
		}
	}
	for _, item := range withdrawals {
		key := "WITHDRAWAL:" + item.PaymentMethod
		group := groups[key]
		if group == nil {
			group = &accumulator{row: PaymentHealthRow{Kind: "WITHDRAWAL", ProviderCode: item.PaymentMethod, MethodCode: item.PaymentMethod}}
			groups[key] = group
		}
		group.row.TotalRequests++
		timeout := (item.Status == 1 || item.Status == 6) && item.CreatedAt.Before(now().Add(-30*time.Minute))
		classifyHealthRecord(&group.row, item.Status, item.CreatedAt, item.CompletedAt, timeout)
		if item.CompletedAt != nil {
			group.latencies = append(group.latencies, float64(item.CompletedAt.Sub(item.CreatedAt).Milliseconds()))
		}
	}
	configured := []struct {
		Code   string
		Status int
	}{}
	cdb := s.db.Table("game_payment_channels").Select("channel_code code,status").Where("deleted_at IS NULL")
	cdb, _ = operationalScope(cdb, actor, value(q, "tenantID"), value(q, "siteID"))
	_ = cdb.Find(&configured).Error
	for _, item := range configured {
		for _, group := range groups {
			if group.row.ChannelCode == item.Code {
				group.configured = item.Status == 1
			}
		}
	}
	result := make([]PaymentHealthRow, 0, len(groups))
	for _, group := range groups {
		finalizeHealth(&group.row, group.latencies, group.configured)
		result = append(result, group.row)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Kind == result[j].Kind {
			return result[i].ProviderCode < result[j].ProviderCode
		}
		return result[i].Kind < result[j].Kind
	})
	return result, nil
}

func classifyHealthRecord(row *PaymentHealthRow, status int, createdAt time.Time, completedAt *time.Time, timeout bool) {
	if status == 2 {
		row.SuccessCount++
		if row.LastSuccessfulTransaction == nil || createdAt.After(*row.LastSuccessfulTransaction) {
			at := createdAt
			row.LastSuccessfulTransaction = &at
		}
	} else if status == 1 || status == 6 {
		row.PendingCount++
	} else {
		row.FailedCount++
		if row.LastFailedTransaction == nil || createdAt.After(*row.LastFailedTransaction) {
			at := createdAt
			row.LastFailedTransaction = &at
		}
	}
	if timeout {
		row.TimeoutCount++
	}
}

func finalizeHealth(row *PaymentHealthRow, latencies []float64, configured bool) {
	if row.TotalRequests > 0 {
		row.SuccessRate = float64(row.SuccessCount) * 100 / float64(row.TotalRequests)
	}
	if len(latencies) > 0 {
		sort.Float64s(latencies)
		total := float64(0)
		for _, v := range latencies {
			total += v
		}
		avg := total / float64(len(latencies))
		p95 := latencies[int(float64(len(latencies)-1)*.95)]
		row.AverageLatencyMS = &avg
		row.P95LatencyMS = &p95
	}
	row.OperationalStatus = "UNKNOWN"
	if !configured && row.ChannelCode != "" {
		row.OperationalStatus = "UNAVAILABLE"
	} else if row.TotalRequests > 0 {
		if row.SuccessRate >= 95 && row.TimeoutCount == 0 {
			row.OperationalStatus = "HEALTHY"
		} else if row.SuccessRate >= 80 {
			row.OperationalStatus = "DEGRADED"
		} else {
			row.OperationalStatus = "UNAVAILABLE"
		}
	}
}

func (s *Service) PaymentDiagnostics(kind string, id uint64, actor Actor) (map[string]any, error) {
	kind = strings.ToUpper(strings.TrimSpace(kind))
	table := "game_deposit_orders"
	walletClause := "wallet_transaction_id"
	if kind == "WITHDRAWAL" {
		table = "game_withdraw_orders"
		walletClause = "final_debit_transaction_id"
	} else if kind != "DEPOSIT" {
		return nil, NewError("INVALID_ORDER_TYPE", "order type must be DEPOSIT or WITHDRAWAL")
	}
	var order map[string]any
	db := s.db.Table(table).Where("id=? AND deleted_at IS NULL", id)
	db, _ = operationalScope(db, actor, "", "")
	if err := db.Take(&order).Error; err != nil {
		return nil, rowNotFound("payment_order")
	}
	tenantID, _ := anyUint64(order["tenant_id"])
	siteID, _ := anyUint64(order["site_id"])
	orderNo := fmt.Sprint(order["order_no"])
	delete(order, "raw_metadata")
	wallet := []map[string]any{}
	_ = s.db.Table("game_wallet_transactions").Select("id,txn_no txnNo,external_txn_no externalTxnNo,direction,category,amount,currency,status,balance_before balanceBefore,balance_after balanceAfter,completed_at completedAt,created_at createdAt").Where("tenant_id=? AND site_id=? AND (order_no=? OR id=?) AND deleted_at IS NULL", tenantID, siteID, orderNo, order[walletClause]).Find(&wallet).Error
	ledger := []map[string]any{}
	_ = s.db.Table("ledger_transactions").Select("id,transaction_no transactionNo,event_type eventType,business_type businessType,business_order_no businessOrderNo,status,debit_amount debitAmount,credit_amount creditAmount,operator_username operator,reason,created_at createdAt").Where("tenant_id=? AND site_id=? AND business_order_no=? AND deleted_at IS NULL", tenantID, siteID, orderNo).Find(&ledger).Error
	callbacks := []map[string]any{}
	if kind == "DEPOSIT" {
		_ = s.db.Table("game_deposit_callback_logs").Select("id,callback_time callbackTime,source_ip sourceIP,signature_valid signatureValid,channel_status channelStatus,processing_result processingResult,retry_count retryCount,created_at createdAt").Where("tenant_id=? AND site_id=? AND deposit_order_id=? AND deleted_at IS NULL", tenantID, siteID, id).Order("id DESC").Find(&callbacks).Error
	} else {
		_ = s.db.Table("game_withdrawal_payout_logs").Select("id,channel_code channelCode,request_result requestResult,channel_order_no channelOrderNo,query_result queryResult,callback_result callbackResult,retry_count retryCount,failure_reason failureReason,requested_at requestedAt,created_at createdAt").Where("tenant_id=? AND site_id=? AND withdrawal_order_id=? AND deleted_at IS NULL", tenantID, siteID, id).Order("id DESC").Find(&callbacks).Error
	}
	events := []map[string]any{}
	_ = s.db.Table("event_outbox").Select("id,event_id eventID,event_type eventType,status,trace_id traceID,correlation_id correlationID,occurred_at occurredAt,last_error lastError").Where("tenant_id=? AND site_id=? AND (aggregate_id=? OR correlation_id=?) AND deleted_at IS NULL", tenantID, siteID, fmt.Sprint(id), orderNo).Order("id DESC").Limit(100).Find(&events).Error
	audits := []map[string]any{}
	_ = s.db.Table("game_admin_operation_logs").Select("id,action,result,operator_username operator,reason,correlation_id requestID,created_at createdAt").Where("tenant_id=? AND site_id=? AND (target_id=? OR correlation_id=?) AND deleted_at IS NULL", tenantID, siteID, id, orderNo).Order("id DESC").Limit(100).Find(&audits).Error
	notes := []map[string]any{}
	_ = s.db.Table("game_operational_notes").Select("id,note,author_username author,created_at createdAt").Where("tenant_id=? AND site_id=? AND target_type=? AND target_id=? AND deleted_at IS NULL", tenantID, siteID, "PAYMENT_"+kind, fmt.Sprint(id)).Order("id ASC").Find(&notes).Error
	return map[string]any{"orderType": kind, "order": redact(order), "walletTransactions": wallet, "ledgerTransactions": ledger, "callbacks": callbacks, "events": events, "audit": audits, "notes": notes, "allowedActions": []string{"MARK_REVIEW", "ADD_NOTE"}}, nil
}

func (s *Service) PaymentDiagnosticsByNumber(kind, orderNumber string, actor Actor) (map[string]any, error) {
	kind = strings.ToUpper(strings.TrimSpace(kind))
	orderNumber = strings.TrimSpace(orderNumber)
	if orderNumber == "" {
		return nil, NewError("ORDER_NUMBER_REQUIRED", "order number is required")
	}
	table := "game_deposit_orders"
	if kind == "WITHDRAWAL" {
		table = "game_withdraw_orders"
	} else if kind != "DEPOSIT" {
		return nil, NewError("INVALID_ORDER_TYPE", "order type must be DEPOSIT or WITHDRAWAL")
	}
	var order struct {
		ID uint64 `gorm:"column:id"`
	}
	db := s.db.Table(table).
		Select("id").
		Where("(order_no=? OR channel_order_no=?) AND deleted_at IS NULL", orderNumber, orderNumber)
	db, _ = operationalScope(db, actor, "", "")
	if err := db.Take(&order).Error; err != nil {
		return nil, rowNotFound("payment_order")
	}
	return s.PaymentDiagnostics(kind, order.ID, actor)
}

func (s *Service) MarkPaymentReview(kind string, id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	kind = strings.ToUpper(strings.TrimSpace(kind))
	table := "game_deposit_orders"
	updates := map[string]any{"reviewer_id": actor.ID, "reviewer_username": actor.Username, "review_reason": strings.TrimSpace(reason), "updated_at": now()}
	if kind == "WITHDRAWAL" {
		table = "game_withdraw_orders"
		updates = map[string]any{"risk_status": "REVIEW", "reviewer_id": actor.ID, "reviewer_username": actor.Username, "review_remark": strings.TrimSpace(reason), "updated_at": now()}
	} else if kind != "DEPOSIT" {
		return NewError("INVALID_ORDER_TYPE", "order type must be DEPOSIT or WITHDRAWAL")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var before map[string]any
		db := tx.Table(table).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id)
		db, _ = operationalScope(db, actor, "", "")
		if err := db.Take(&before).Error; err != nil {
			return rowNotFound("payment_order")
		}
		tenantID, _ := anyUint64(before["tenant_id"])
		siteID, _ := anyUint64(before["site_id"])
		if err := tx.Table(table).Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, targetAuditActor(actor, tenantID, siteID), "monitor:payment:review", "payment.mark_review", strings.ToLower(kind)+"_order", id, before, updates, reason)
	})
}

func (s *Service) ProviderMonitor(q Query, actor Actor) (map[string]any, error) {
	from, to, err := supportTimeRange(value(q, "from"), value(q, "to"))
	if err != nil {
		return nil, err
	}
	providers := []struct{ ProviderCode, Name string }{}
	pdb := s.db.Table("game_game_providers").Select("provider_code,MAX(name) name").Where("deleted_at IS NULL")
	if value(q, "providerCode") != "" {
		pdb = pdb.Where("provider_code=?", strings.ToUpper(value(q, "providerCode")))
	}
	_ = pdb.Group("provider_code").Find(&providers).Error
	rows := make([]map[string]any, 0, len(providers))
	for _, provider := range providers {
		row, err := s.providerHealth(provider.ProviderCode, provider.Name, q, actor, from, to)
		if err != nil {
			return nil, err
		}
		if status := strings.ToUpper(value(q, "status")); status != "" && row["healthStatus"] != status {
			continue
		}
		rows = append(rows, row)
	}
	return map[string]any{"from": from, "to": to, "list": rows}, nil
}

func (s *Service) providerHealth(code, name string, q Query, actor Actor, from, to time.Time) (map[string]any, error) {
	tenantID := value(q, "tenantID")
	siteID := value(q, "siteID")
	games := struct{ EnabledTenantCount, EnabledGameCount int64 }{}
	gdb := s.db.Table("game_games g").Select("COUNT(DISTINCT g.tenant_id) enabled_tenant_count,COUNT(*) enabled_game_count").Joins("JOIN game_game_providers p ON p.id=g.provider_id AND p.deleted_at IS NULL").Where("p.provider_code=? AND g.platform_enabled=1 AND g.status=1 AND g.deleted_at IS NULL", code)
	gdb, _ = operationalScopeAlias(gdb, "g", actor, tenantID, siteID)
	if country := strings.ToUpper(value(q, "country")); country != "" {
		gdb = gdb.Where("(g.supported_countries IS NULL OR JSON_LENGTH(g.supported_countries)=0 OR JSON_CONTAINS(g.supported_countries, JSON_QUOTE(?)))", country)
	}
	if currency := strings.ToUpper(value(q, "currency")); currency != "" {
		gdb = gdb.Where("(g.supported_currencies IS NULL OR JSON_LENGTH(g.supported_currencies)=0 OR JSON_CONTAINS(g.supported_currencies, JSON_QUOTE(?)))", currency)
	}
	_ = gdb.Scan(&games).Error
	launches := []struct {
		Status    int
		CreatedAt time.Time
		SuccessAt *time.Time
	}{}
	ldb := s.db.Table("game_game_launch_logs l").Select("l.launch_status status,l.created_at,l.launch_success_time success_at").Joins("JOIN game_game_providers p ON p.id=l.provider_id AND p.deleted_at IS NULL").Where("p.provider_code=? AND l.created_at BETWEEN ? AND ? AND l.deleted_at IS NULL", code, from, to)
	ldb, _ = operationalScopeAlias(ldb, "l", actor, tenantID, siteID)
	if country := strings.ToUpper(value(q, "country")); country != "" {
		ldb = ldb.Where("l.country=?", country)
	}
	if currency := strings.ToUpper(value(q, "currency")); currency != "" {
		ldb = ldb.Where("l.currency=?", currency)
	}
	_ = ldb.Order("l.id DESC").Limit(10000).Find(&launches).Error
	callbacks := []struct {
		ProcessingResult, SignatureResult string
		DurationMS                        uint64
		CreatedAt                         time.Time
	}{}
	cdb := s.db.Table("game_provider_callback_logs c").Select("processing_result,signature_result,duration_ms,created_at").Where("provider_code=? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", code, from, to)
	cdb, _ = operationalScopeAlias(cdb, "c", actor, tenantID, siteID)
	if currency := strings.ToUpper(value(q, "currency")); currency != "" {
		cdb = cdb.Where("c.currency=?", currency)
	}
	if country := strings.ToUpper(value(q, "country")); country != "" {
		cdb = cdb.Where("EXISTS (SELECT 1 FROM game_users u WHERE u.id=c.user_id AND u.tenant_id=c.tenant_id AND u.site_id=c.site_id AND u.country_code=?)", country)
	}
	_ = cdb.Order("id DESC").Limit(10000).Find(&callbacks).Error
	launchSuccess, launchFailure := int64(0), int64(0)
	latencies := []float64{}
	var lastSuccess, lastFailure *time.Time
	for _, item := range launches {
		if item.Status == 2 {
			launchSuccess++
			lastSuccess = later(lastSuccess, item.CreatedAt)
		} else if item.Status == 3 {
			launchFailure++
			lastFailure = later(lastFailure, item.CreatedAt)
		}
		if item.SuccessAt != nil {
			latencies = append(latencies, float64(item.SuccessAt.Sub(item.CreatedAt).Milliseconds()))
		}
	}
	callbackSuccess, callbackFailure := int64(0), int64(0)
	signatureFailures := int64(0)
	for _, item := range callbacks {
		if strings.EqualFold(item.ProcessingResult, "SUCCESS") {
			callbackSuccess++
			lastSuccess = later(lastSuccess, item.CreatedAt)
		} else {
			callbackFailure++
			lastFailure = later(lastFailure, item.CreatedAt)
		}
		if !strings.EqualFold(item.SignatureResult, "VALID") && item.SignatureResult != "" {
			signatureFailures++
		}
		latencies = append(latencies, float64(item.DurationMS))
	}
	maintenance := int64(0)
	mdb := s.db.Table("game_provider_maintenance").Where("provider_code=? AND enabled=1 AND (start_at IS NULL OR start_at<=?) AND (end_at IS NULL OR end_at>?) AND deleted_at IS NULL", code, now(), now())
	if !actor.Platform {
		mdb = mdb.Where("tenant_id=0 OR (tenant_id=? AND (site_id=0 OR site_id=?))", actor.TenantID, actor.SiteID)
	} else if tenantID != "" {
		mdb = mdb.Where("tenant_id=0 OR tenant_id=?", tenantID)
	}
	_ = mdb.Count(&maintenance).Error
	retryCount := int64(0)
	rdb := s.db.Table("game_system_retry_records r").
		Joins("JOIN game_provider_callback_logs c ON c.id=r.source_id AND c.tenant_id=r.tenant_id AND c.site_id=r.site_id AND c.deleted_at IS NULL").
		Where("r.source_type='PROVIDER_CALLBACK' AND r.status IN ('PENDING','FAILED') AND r.deleted_at IS NULL AND c.provider_code=?", code)
	rdb, _ = operationalScopeAlias(rdb, "r", actor, tenantID, siteID)
	_ = rdb.Count(&retryCount).Error
	launchTotal := launchSuccess + launchFailure
	callbackTotal := callbackSuccess + callbackFailure
	launchRate, callbackRate := float64(0), float64(0)
	if launchTotal > 0 {
		launchRate = float64(launchSuccess) * 100 / float64(launchTotal)
	}
	if callbackTotal > 0 {
		callbackRate = float64(callbackSuccess) * 100 / float64(callbackTotal)
	}
	status := "UNKNOWN"
	if maintenance > 0 {
		status = "MAINTENANCE"
	} else if launchTotal+callbackTotal > 0 {
		combined := float64(launchSuccess+callbackSuccess) * 100 / float64(launchTotal+callbackTotal)
		if combined >= 95 {
			status = "HEALTHY"
		} else if combined >= 80 {
			status = "DEGRADED"
		} else {
			status = "UNAVAILABLE"
		}
	}
	var average any = nil
	if len(latencies) > 0 {
		sum := float64(0)
		for _, v := range latencies {
			sum += v
		}
		average = sum / float64(len(latencies))
	}
	return map[string]any{"providerCode": code, "providerName": name, "enabledTenantCount": games.EnabledTenantCount, "enabledGameCount": games.EnabledGameCount, "launchSuccessRate": launchRate, "callbackSuccessRate": callbackRate, "averageLatencyMS": average, "lastSuccessfulRequest": lastSuccess, "lastFailedRequest": lastFailure, "failureCount": launchFailure + callbackFailure, "signatureFailureCount": signatureFailures, "retryQueueCount": retryCount, "maintenance": maintenance > 0, "healthStatus": status}, nil
}

func (s *Service) ProviderDiagnostics(code string, q Query, actor Actor) (map[string]any, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, NewError("PROVIDER_REQUIRED", "provider is required")
	}
	from, to, err := supportTimeRange(value(q, "from"), value(q, "to"))
	if err != nil {
		return nil, err
	}
	health, err := s.providerHealth(code, code, q, actor, from, to)
	if err != nil {
		return nil, err
	}
	launches := []map[string]any{}
	ldb := s.db.Table("game_game_launch_logs l").Select("l.id,l.tenant_id tenantID,l.site_id siteID,l.user_id playerID,l.launch_no launchNo,l.launch_status status,l.correlation_id correlationID,l.error_message errorMessage,l.created_at createdAt").Joins("JOIN game_game_providers p ON p.id=l.provider_id AND p.provider_code=? AND p.deleted_at IS NULL", code).Where("l.created_at BETWEEN ? AND ? AND l.deleted_at IS NULL", from, to)
	ldb, _ = operationalScopeAlias(ldb, "l", actor, value(q, "tenantID"), value(q, "siteID"))
	if country := strings.ToUpper(value(q, "country")); country != "" {
		ldb = ldb.Where("l.country=?", country)
	}
	if currency := strings.ToUpper(value(q, "currency")); currency != "" {
		ldb = ldb.Where("l.currency=?", currency)
	}
	_ = ldb.Order("l.id DESC").Limit(100).Find(&launches).Error
	callbacks := []map[string]any{}
	cdb := s.db.Table("game_provider_callback_logs c").Select("c.id,c.tenant_id tenantID,c.site_id siteID,c.callback_id callbackID,c.operation,c.provider_transaction_id providerTransactionID,c.provider_round_id providerRoundID,c.signature_result signatureResult,c.processing_result processingResult,c.http_status httpStatus,c.duration_ms durationMS,c.correlation_id correlationID,c.failure_reason failureReason,c.created_at createdAt").Where("c.provider_code=? AND c.created_at BETWEEN ? AND ? AND c.deleted_at IS NULL", code, from, to)
	cdb, _ = operationalScopeAlias(cdb, "c", actor, value(q, "tenantID"), value(q, "siteID"))
	if currency := strings.ToUpper(value(q, "currency")); currency != "" {
		cdb = cdb.Where("c.currency=?", currency)
	}
	if country := strings.ToUpper(value(q, "country")); country != "" {
		cdb = cdb.Where("EXISTS (SELECT 1 FROM game_users u WHERE u.id=c.user_id AND u.tenant_id=c.tenant_id AND u.site_id=c.site_id AND u.country_code=?)", country)
	}
	_ = cdb.Order("c.id DESC").Limit(200).Find(&callbacks).Error
	transactions := []map[string]any{}
	tdb := s.db.Table("game_provider_transactions t").Select("id,tenant_id tenantID,site_id siteID,user_id playerID,provider_transaction_id providerTransactionID,provider_round_id providerRoundID,operation,amount_minor amountMinor,currency,status,related_transaction_id relatedTransactionID,created_at createdAt").Where("provider_code=? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", code, from, to)
	tdb, _ = operationalScopeAlias(tdb, "t", actor, value(q, "tenantID"), value(q, "siteID"))
	if currency := strings.ToUpper(value(q, "currency")); currency != "" {
		tdb = tdb.Where("t.currency=?", currency)
	}
	if country := strings.ToUpper(value(q, "country")); country != "" {
		tdb = tdb.Where("EXISTS (SELECT 1 FROM game_users u WHERE u.id=t.user_id AND u.tenant_id=t.tenant_id AND u.site_id=t.site_id AND u.country_code=?)", country)
	}
	_ = tdb.Order("id DESC").Limit(200).Find(&transactions).Error

	duplicates := []map[string]any{}
	signatureFailures := []map[string]any{}
	errorSamples := []map[string]any{}
	type errorBucket struct {
		Operation string `json:"operation"`
		Result    string `json:"result"`
		Count     int    `json:"count"`
	}
	buckets := map[string]*errorBucket{}
	transactionCounts := map[string]int{}
	for _, callback := range callbacks {
		transactionID := strings.TrimSpace(fmt.Sprint(callback["providerTransactionID"]))
		if transactionID != "" {
			transactionCounts[transactionID]++
		}
		if signature := strings.ToUpper(fmt.Sprint(callback["signatureResult"])); signature != "" && signature != "VALID" {
			signatureFailures = append(signatureFailures, callback)
		}
		result := strings.ToUpper(fmt.Sprint(callback["processingResult"]))
		if result != "SUCCESS" {
			errorSamples = append(errorSamples, callback)
			operation := strings.ToUpper(fmt.Sprint(callback["operation"]))
			key := operation + "|" + result
			if buckets[key] == nil {
				buckets[key] = &errorBucket{Operation: operation, Result: result}
			}
			buckets[key].Count++
		}
	}
	for transactionID, count := range transactionCounts {
		if count > 1 {
			duplicates = append(duplicates, map[string]any{"providerTransactionID": transactionID, "count": count})
		}
	}
	errorDistribution := make([]errorBucket, 0, len(buckets))
	for _, bucket := range buckets {
		errorDistribution = append(errorDistribution, *bucket)
	}
	sort.Slice(duplicates, func(i, j int) bool {
		return fmt.Sprint(duplicates[i]["providerTransactionID"]) < fmt.Sprint(duplicates[j]["providerTransactionID"])
	})
	sort.Slice(errorDistribution, func(i, j int) bool {
		if errorDistribution[i].Count == errorDistribution[j].Count {
			return errorDistribution[i].Operation < errorDistribution[j].Operation
		}
		return errorDistribution[i].Count > errorDistribution[j].Count
	})

	retryHistory := []map[string]any{}
	rdb := s.db.Table("game_system_retry_records r").
		Select("r.id,r.tenant_id tenantID,r.site_id siteID,r.source_id sourceID,r.status,r.attempt_count attemptCount,r.max_attempts maxAttempts,r.next_attempt_at nextAttemptAt,r.last_error lastError,r.created_at createdAt,r.updated_at updatedAt").
		Joins("JOIN game_provider_callback_logs c ON c.id=r.source_id AND c.tenant_id=r.tenant_id AND c.site_id=r.site_id AND c.deleted_at IS NULL").
		Where("r.source_type='PROVIDER_CALLBACK' AND c.provider_code=? AND r.deleted_at IS NULL", code)
	rdb, _ = operationalScopeAlias(rdb, "r", actor, value(q, "tenantID"), value(q, "siteID"))
	_ = rdb.Order("r.id DESC").Limit(200).Find(&retryHistory).Error

	return map[string]any{
		"health": health, "launchRequests": launches, "callbackRecords": callbacks, "transactions": transactions,
		"duplicateCallbacks": duplicates, "signatureFailures": signatureFailures, "recentErrors": errorSamples,
		"errorDistribution": errorDistribution, "retryHistory": retryHistory,
	}, nil
}

func later(current *time.Time, candidate time.Time) *time.Time {
	if current == nil || candidate.After(*current) {
		copy := candidate
		return &copy
	}
	return current
}

func (s *Service) AddOperationalNote(input OperationalNoteInput, actor Actor) (uint64, error) {
	if err := requireReason(input.Reason); err != nil {
		return 0, err
	}
	if strings.TrimSpace(input.Note) == "" || strings.TrimSpace(input.TargetType) == "" || strings.TrimSpace(input.TargetID) == "" {
		return 0, NewError("INVALID_NOTE", "target and note are required")
	}
	tenantID, siteID, err := operationalTargetScope(s.db, actor, input.TenantID, input.SiteID)
	if err != nil {
		return 0, err
	}
	var id uint64
	err = s.db.Transaction(func(tx *gorm.DB) error {
		row := map[string]any{"tenant_id": tenantID, "site_id": siteID, "target_type": strings.ToUpper(input.TargetType), "target_id": strings.TrimSpace(input.TargetID), "note": strings.TrimSpace(input.Note), "author_id": actor.ID, "author_username": actor.Username, "created_at": now(), "updated_at": now()}
		if err := tx.Table("game_operational_notes").Create(row).Error; err != nil {
			return err
		}
		if err := tx.Table("game_operational_notes").Select("id").Where("tenant_id=? AND site_id=? AND author_id=?", tenantID, siteID, actor.ID).Order("id DESC").Take(&id).Error; err != nil {
			return err
		}
		return audit(tx, targetAuditActor(actor, tenantID, siteID), "operations:note:write", "operations.note.create", "operational_note", id, nil, map[string]any{"targetType": input.TargetType, "targetID": input.TargetID}, input.Reason)
	})
	return id, err
}
