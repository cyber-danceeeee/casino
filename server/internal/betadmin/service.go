package betadmin

import (
	"crypto/rand"
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

	"casino/internal/database"
	"casino/internal/model"
	"casino/internal/tenancy"
)

type Service struct{ db *gorm.DB }

func NewService() *Service { return &Service{db: database.GetDB()} }

func normalizePage(query ListQuery) ListQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	return query
}

func money(minor int64) string {
	return decimal.NewFromInt(minor).Div(decimal.NewFromInt(100)).StringFixed(2)
}

func fillBetMoney(rows []BetRow) {
	for i := range rows {
		rows[i].BetAmount = money(rows[i].BetAmountMinor)
		rows[i].ValidBetAmount = money(rows[i].ValidBetAmountMinor)
		rows[i].WinAmount = money(rows[i].WinAmountMinor)
		rows[i].PlatformProfit = money(rows[i].PlatformProfitMinor)
	}
}
func setBetMoney(row *BetRow) {
	row.BetAmount = money(row.BetAmountMinor)
	row.ValidBetAmount = money(row.ValidBetAmountMinor)
	row.WinAmount = money(row.WinAmountMinor)
	row.PlatformProfit = money(row.PlatformProfitMinor)
}
func fillRoundMoney(rows []RoundRow) {
	for i := range rows {
		rows[i].TotalBet = money(rows[i].TotalBetMinor)
		rows[i].TotalWin = money(rows[i].TotalWinMinor)
		rows[i].PlatformProfit = money(rows[i].PlatformProfitMinor)
	}
}
func setRoundMoney(row *RoundRow) {
	row.TotalBet = money(row.TotalBetMinor)
	row.TotalWin = money(row.TotalWinMinor)
	row.PlatformProfit = money(row.PlatformProfitMinor)
}
func fillRollbackMoney(rows []RollbackRow) {
	for i := range rows {
		rows[i].OriginalAmount = money(rows[i].OriginalAmountMinor)
		rows[i].RollbackAmount = money(rows[i].RollbackAmountMinor)
	}
}
func fillAbnormalMoney(rows []AbnormalRow) {
	for i := range rows {
		rows[i].BetAmount = money(rows[i].BetAmountMinor)
		rows[i].WinAmount = money(rows[i].WinAmountMinor)
	}
}
func fillCallbackMoney(rows []CallbackRow) {
	for i := range rows {
		rows[i].Amount = money(rows[i].AmountMinor)
	}
}

func paginate(db *gorm.DB, query ListQuery, target any) (PageResult, error) {
	query = normalizePage(query)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return PageResult{}, err
	}
	if err := db.Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Scan(target).Error; err != nil {
		return PageResult{}, err
	}
	return PageResult{List: target, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func contains(db *gorm.DB, column, value string) *gorm.DB {
	value = strings.TrimSpace(value)
	if value == "" {
		return db
	}
	return db.Where(column+" LIKE ?", "%"+value+"%")
}

func equals(db *gorm.DB, column, value string) *gorm.DB {
	value = strings.TrimSpace(value)
	if value == "" {
		return db
	}
	return db.Where(column+" = ?", value)
}

func timeRange(db *gorm.DB, column, from, to string) *gorm.DB {
	if strings.TrimSpace(from) != "" {
		db = db.Where(column+" >= ?", from)
	}
	if strings.TrimSpace(to) != "" {
		db = db.Where(column+" <= ?", to)
	}
	return db
}

func minorRange(db *gorm.DB, column, minValue, maxValue string) *gorm.DB {
	if value, ok := parseMinor(minValue); ok {
		db = db.Where(column+" >= ?", value)
	}
	if value, ok := parseMinor(maxValue); ok {
		db = db.Where(column+" <= ?", value)
	}
	return db
}

func parseMinor(value string) (int64, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return 0, false
	}
	return parsed.Mul(decimal.NewFromInt(100)).IntPart(), true
}

func userPhoneDisplay(alias string) string {
	return fmt.Sprintf("CASE WHEN NULLIF(%s.phone,'') IS NOT NULL THEN TRIM(CONCAT_WS(' ',NULLIF(%s.phone_country_code,''),%s.phone)) ELSE COALESCE(%s.username,'') END", alias, alias, alias, alias)
}

func baseBetSelect() string {
	return fmt.Sprintf(`b.id,b.platform_bet_id,b.user_id,u.user_no,%s username,b.game_id,COALESCE(g.name,'') game_name,COALESCE(NULLIF(g.local_thumbnail_url,''),NULLIF(g.cover_url,''),'') game_icon,b.provider_code,b.provider_game_id,b.provider_round_id,b.provider_bet_transaction_id,b.provider_win_transaction_id,b.game_type,b.bet_amount_minor,b.valid_bet_amount_minor,b.win_amount_minor,b.platform_profit_minor,b.currency,b.bet_status status,b.rollback_status,b.failure_reason,b.bet_at,b.settled_at,b.created_at,b.updated_at`, userPhoneDisplay("u"))
}

func callbackUserIDSQL() string {
	return "COALESCE(NULLIF(c.user_id,0),pp.user_id,pt.user_id,br.user_id,0)"
}

func callbackProviderRoundSQL() string {
	return "COALESCE(NULLIF(c.provider_round_id,''),NULLIF(pt.provider_round_id,''),NULLIF(br.provider_round_id,''),'')"
}

func callbackProviderGameSQL() string {
	return "COALESCE(NULLIF(c.provider_game_id,''),NULLIF(pt.provider_game_id,''),NULLIF(br.provider_game_id,''),'')"
}

func callbackSelectSQL(includeRequest bool) string {
	requestFields := ""
	if includeRequest {
		requestFields = ",c.request_path,c.request_headers,c.query_parameters,c.request_body,c.timestamp_validation,c.normalized_command,c.wallet_result,c.provider_response,c.duration_ms,c.correlation_id"
	}
	return fmt.Sprintf(`
		c.id,c.callback_id,c.provider_code,c.operation,
		%s user_id,
		u.user_no,%s username,%s user_phone,
		COALESCE(g.id,0) game_id,COALESCE(g.name,'') game_name,
		c.provider_player_id,c.provider_transaction_id,%s provider_round_id,%s provider_game_id,
		c.amount_minor,c.currency,c.http_method%s,c.signature_result,c.processing_result,c.http_status,
		c.source_ip,c.failure_reason,c.marked_for_investigation,c.created_at`,
		callbackUserIDSQL(), userPhoneDisplay("u"), userPhoneDisplay("u"),
		callbackProviderRoundSQL(), callbackProviderGameSQL(), requestFields)
}

func callbackJoins(db *gorm.DB) *gorm.DB {
	return db.
		Joins("LEFT JOIN game_provider_players pp ON pp.provider_code=c.provider_code AND pp.provider_player_id=c.provider_player_id AND pp.deleted_at IS NULL").
		Joins(`LEFT JOIN game_provider_transactions pt ON pt.id = (
			SELECT pt2.id FROM game_provider_transactions pt2
			WHERE pt2.provider_code=c.provider_code AND pt2.provider_transaction_id=c.provider_transaction_id AND pt2.deleted_at IS NULL
			ORDER BY pt2.id DESC LIMIT 1
		)`).
		Joins(`LEFT JOIN game_bet_records br ON br.id = (
			SELECT br2.id FROM game_bet_records br2
			WHERE br2.provider_code=c.provider_code AND br2.deleted_at IS NULL AND (
				br2.provider_bet_transaction_id=c.provider_transaction_id OR
				br2.provider_win_transaction_id=c.provider_transaction_id OR
				(COALESCE(NULLIF(c.provider_round_id,''),NULLIF(pt.provider_round_id,''),'') <> '' AND br2.provider_round_id=COALESCE(NULLIF(c.provider_round_id,''),NULLIF(pt.provider_round_id,''),''))
			)
			ORDER BY br2.id DESC LIMIT 1
		)`).
		Joins("LEFT JOIN game_users u ON u.id=" + callbackUserIDSQL() + " AND u.deleted_at IS NULL").
		Joins("LEFT JOIN game_game_providers p ON p.provider_code=c.provider_code AND p.deleted_at IS NULL").
		Joins("LEFT JOIN game_games g ON g.id=COALESCE(NULLIF(br.game_id,0),(SELECT gg.id FROM game_games gg WHERE gg.provider_id=p.id AND gg.provider_game_code=" + callbackProviderGameSQL() + " AND gg.deleted_at IS NULL ORDER BY gg.id DESC LIMIT 1)) AND g.deleted_at IS NULL")
}

func (s *Service) ListBets(query ListQuery) (PageResult, error) {
	rows := []BetRow{}
	v := query.Values
	db := s.db.Table("game_bet_records AS b").Select(baseBetSelect()).Joins("LEFT JOIN game_users u ON u.id=b.user_id").Joins("LEFT JOIN game_games g ON g.id=b.game_id").Where("b.deleted_at IS NULL")
	db = tenancy.ApplySite(db, "b", query.Scope)
	db = contains(db, "b.platform_bet_id", v["platformBetID"])
	db = contains(db, "b.provider_bet_transaction_id", v["providerTransactionID"])
	db = contains(db, "b.provider_round_id", v["providerRoundID"])
	db = contains(db, "u.user_no", v["userNo"])
	if username := strings.TrimSpace(v["username"]); username != "" {
		like := "%" + username + "%"
		db = db.Where("u.phone LIKE ? OR u.username LIKE ?", like, like)
	}
	db = equals(db, "b.provider_code", v["providerCode"])
	db = equals(db, "b.game_id", v["gameID"])
	db = contains(db, "b.provider_game_id", v["providerGameID"])
	db = equals(db, "b.game_type", v["gameType"])
	db = equals(db, "b.currency", v["currency"])
	db = equals(db, "b.bet_status", v["status"])
	db = equals(db, "b.rollback_status", v["rollbackStatus"])
	db = minorRange(db, "b.bet_amount_minor", v["amountMin"], v["amountMax"])
	db = minorRange(db, "b.win_amount_minor", v["winAmountMin"], v["winAmountMax"])
	db = timeRange(db, "b.created_at", v["createdFrom"], v["createdTo"])
	db = timeRange(db, "b.settled_at", v["settledFrom"], v["settledTo"])
	db = db.Order("b.created_at DESC,b.id DESC")
	result, err := paginate(db, query, &rows)
	fillBetMoney(rows)
	result.List = rows
	return result, err
}

func (s *Service) ListRounds(query ListQuery) (PageResult, error) {
	rows := []RoundRow{}
	v := query.Values
	db := s.db.Table("game_rounds r").Select("r.id,r.provider_round_id,r.user_id,u.user_no,u.username,r.game_id,COALESCE(g.name,'') game_name,r.provider_code,r.provider_game_id,r.total_bet_minor,r.total_win_minor,r.platform_profit_minor,r.currency,r.transaction_count,r.status,r.started_at,r.completed_at,r.created_at,r.updated_at").Joins("LEFT JOIN game_users u ON u.id=r.user_id").Joins("LEFT JOIN game_games g ON g.id=r.game_id").Where("r.deleted_at IS NULL")
	db = tenancy.ApplySite(db, "r", query.Scope)
	db = contains(db, "r.provider_round_id", v["providerRoundID"])
	db = contains(db, "u.user_no", v["userNo"])
	db = contains(db, "u.username", v["username"])
	db = equals(db, "r.provider_code", v["providerCode"])
	db = contains(db, "g.name", v["game"])
	db = equals(db, "r.currency", v["currency"])
	db = equals(db, "r.status", v["status"])
	db = minorRange(db, "r.total_bet_minor", v["betAmountMin"], v["betAmountMax"])
	db = minorRange(db, "r.total_win_minor", v["winAmountMin"], v["winAmountMax"])
	db = timeRange(db, "r.started_at", v["startedFrom"], v["startedTo"])
	db = timeRange(db, "r.completed_at", v["completedFrom"], v["completedTo"])
	db = db.Order("r.created_at DESC,r.id DESC")
	result, err := paginate(db, query, &rows)
	fillRoundMoney(rows)
	result.List = rows
	return result, err
}

func (s *Service) ListRollbacks(query ListQuery) (PageResult, error) {
	rows := []RollbackRow{}
	v := query.Values
	db := s.db.Table("game_rollbacks rb").Select("rb.id,rb.rollback_id,rb.original_provider_transaction_id,rb.provider_transaction_id,rb.provider_round_id,rb.user_id,u.user_no,u.username,rb.provider_code,rb.game_id,COALESCE(g.name,'') game_name,rb.original_operation,rb.original_amount_minor,rb.rollback_amount_minor,rb.currency,rb.status,rb.failure_reason,rb.created_at,rb.completed_at").Joins("LEFT JOIN game_users u ON u.id=rb.user_id").Joins("LEFT JOIN game_games g ON g.id=rb.game_id").Where("rb.deleted_at IS NULL")
	db = tenancy.ApplySite(db, "rb", query.Scope)
	db = contains(db, "rb.rollback_id", v["rollbackID"])
	db = contains(db, "rb.original_provider_transaction_id", v["originalTransactionID"])
	db = contains(db, "rb.provider_transaction_id", v["providerTransactionID"])
	db = contains(db, "rb.provider_round_id", v["providerRoundID"])
	db = contains(db, "u.user_no", v["userNo"])
	db = contains(db, "u.username", v["username"])
	db = equals(db, "rb.provider_code", v["providerCode"])
	db = equals(db, "rb.original_operation", v["operation"])
	db = equals(db, "rb.currency", v["currency"])
	db = equals(db, "rb.status", v["status"])
	db = timeRange(db, "rb.created_at", v["createdFrom"], v["createdTo"])
	db = timeRange(db, "rb.completed_at", v["completedFrom"], v["completedTo"])
	db = db.Order("rb.created_at DESC,rb.id DESC")
	result, err := paginate(db, query, &rows)
	fillRollbackMoney(rows)
	result.List = rows
	return result, err
}

func (s *Service) ListAbnormal(query ListQuery) (PageResult, error) {
	rows := []AbnormalRow{}
	v := query.Values
	db := s.db.Table("game_abnormal_bet_events a").Select("a.id,a.risk_event_id,a.bet_record_id,a.user_id,u.user_no,u.username,a.provider_code,a.game_id,COALESCE(g.name,'') game_name,a.provider_round_id,a.risk_type,a.risk_level,a.review_status,a.bet_amount_minor,a.win_amount_minor,a.trigger_count,a.assigned_reviewer,a.latest_reason,a.created_at,a.updated_at").Joins("LEFT JOIN game_users u ON u.id=a.user_id").Joins("LEFT JOIN game_games g ON g.id=a.game_id").Where("a.deleted_at IS NULL")
	db = tenancy.ApplySite(db, "a", query.Scope)
	db = contains(db, "u.user_no", v["userNo"])
	db = contains(db, "u.username", v["username"])
	db = equals(db, "a.provider_code", v["providerCode"])
	db = contains(db, "g.name", v["game"])
	db = equals(db, "a.risk_level", v["riskLevel"])
	db = equals(db, "a.risk_type", v["riskType"])
	db = equals(db, "a.review_status", v["reviewStatus"])
	db = contains(db, "a.assigned_reviewer", v["assignedReviewer"])
	db = timeRange(db, "a.created_at", v["createdFrom"], v["createdTo"])
	db = timeRange(db, "a.updated_at", v["updatedFrom"], v["updatedTo"])
	db = db.Order("FIELD(a.risk_level,'CRITICAL','HIGH','MEDIUM','LOW'),a.updated_at DESC,a.id DESC")
	result, err := paginate(db, query, &rows)
	fillAbnormalMoney(rows)
	result.List = rows
	return result, err
}

func (s *Service) ListCallbacks(query ListQuery) (PageResult, error) {
	rows := []CallbackRow{}
	v := query.Values
	db := callbackJoins(s.db.Table("game_provider_callback_logs c").Select(callbackSelectSQL(false))).
		Where("c.deleted_at IS NULL")
	db = tenancy.ApplySite(db, "c", query.Scope)
	db = equals(db, "c.provider_code", v["providerCode"])
	db = equals(db, "c.operation", v["operation"])
	db = contains(db, "u.user_no", v["userNo"])
	db = contains(db, "u.phone", v["userPhone"])
	db = contains(db, "c.provider_player_id", v["providerPlayerID"])
	db = contains(db, "c.provider_transaction_id", v["providerTransactionID"])
	db = contains(db, callbackProviderRoundSQL(), v["providerRoundID"])
	db = contains(db, callbackProviderGameSQL(), v["gameID"])
	db = equals(db, "c.http_method", v["httpMethod"])
	db = equals(db, "c.signature_result", v["signatureResult"])
	db = equals(db, "c.processing_result", v["processingResult"])
	db = contains(db, "c.source_ip", v["sourceIP"])
	db = timeRange(db, "c.created_at", v["createdFrom"], v["createdTo"])
	db = db.Order("c.created_at DESC,c.id DESC")
	result, err := paginate(db, query, &rows)
	fillCallbackMoney(rows)
	result.List = rows
	return result, err
}

func (s *Service) BetDetail(id uint64, scope tenancy.Scope) (map[string]any, error) {
	row, err := s.singleBet(id, scope)
	if err != nil {
		return nil, notFound(err)
	}
	return map[string]any{"overview": row, "transactionTimeline": s.providerTransactions(row.ProviderCode, row.ProviderRoundID, row.UserID, scope), "walletTransactions": s.walletTransactions(row.UserID, "game_provider_transaction", 0, scope), "providerRequests": s.callbackLogs(row.ProviderCode, row.ProviderRoundID, row.UserID, scope), "roundInformation": s.roundByProvider(row.ProviderCode, row.ProviderRoundID, row.UserID, scope), "operationLogs": s.operationLogs("bet", id, scope)}, nil
}

func (s *Service) RoundDetail(id uint64, scope tenancy.Scope) (map[string]any, error) {
	var row RoundRow
	db := s.db.Table("game_rounds r").Select("r.id,r.provider_round_id,r.user_id,u.user_no,u.username,r.game_id,COALESCE(g.name,'') game_name,r.provider_code,r.provider_game_id,r.total_bet_minor,r.total_win_minor,r.platform_profit_minor,r.currency,r.transaction_count,r.status,r.started_at,r.completed_at,r.created_at,r.updated_at").Joins("LEFT JOIN game_users u ON u.id=r.user_id").Joins("LEFT JOIN game_games g ON g.id=r.game_id").Where("r.id=? AND r.deleted_at IS NULL", id)
	db = tenancy.ApplySite(db, "r", scope).Take(&row)
	if db.Error != nil {
		return nil, notFound(db.Error)
	}
	setRoundMoney(&row)
	bets := []BetRow{}
	betDB := s.db.Table("game_bet_records b").Select(baseBetSelect()).Joins("LEFT JOIN game_users u ON u.id=b.user_id").Joins("LEFT JOIN game_games g ON g.id=b.game_id").Where("b.provider_code=? AND b.provider_round_id=? AND b.user_id=? AND b.deleted_at IS NULL", row.ProviderCode, row.ProviderRoundID, row.UserID)
	_ = tenancy.ApplySite(betDB, "b", scope).Order("b.id").Scan(&bets).Error
	fillBetMoney(bets)
	return map[string]any{"roundSummary": row, "bettingRecords": bets, "walletTransactions": s.providerTransactions(row.ProviderCode, row.ProviderRoundID, row.UserID, scope), "providerCallbacks": s.callbackLogs(row.ProviderCode, row.ProviderRoundID, row.UserID, scope), "operationLogs": s.operationLogs("round", id, scope)}, nil
}

func (s *Service) RollbackDetail(id uint64, scope tenancy.Scope) (map[string]any, error) {
	rows, err := s.rollbackRows(scope, "rb.id = ?", id)
	if err != nil || len(rows) == 0 {
		return nil, newError("ROLLBACK_NOT_FOUND", "rollback record not found")
	}
	row := rows[0]
	return map[string]any{"overview": row, "originalTransaction": s.providerTransaction(row.ProviderCode, row.OriginalProviderTransactionID, scope), "rollbackTransaction": s.providerTransaction(row.ProviderCode, row.ProviderTransactionID, scope), "walletEffect": s.walletTransactions(row.UserID, "game_provider_transaction", 0, scope), "providerRequest": s.callbackLogs(row.ProviderCode, row.ProviderRoundID, row.UserID, scope), "operationLogs": s.operationLogs("rollback", id, scope)}, nil
}

func (s *Service) AbnormalDetail(id uint64, scope tenancy.Scope) (map[string]any, error) {
	rows, err := s.abnormalRows(scope, "a.id = ?", id)
	if err != nil || len(rows) == 0 {
		return nil, newError("ABNORMAL_BET_NOT_FOUND", "abnormal bet not found")
	}
	row := rows[0]
	var history, notes []map[string]any
	_ = tenancy.ApplySite(s.db.Table("game_abnormal_bet_review_histories"), "", scope).Where("abnormal_bet_id=? AND deleted_at IS NULL", id).Order("created_at DESC").Find(&history).Error
	_ = tenancy.ApplySite(s.db.Table("game_abnormal_bet_notes"), "", scope).Where("abnormal_bet_id=? AND deleted_at IS NULL", id).Order("created_at DESC").Find(&notes).Error
	bet, _ := s.singleBet(row.BetRecordID, scope)
	return map[string]any{"riskSummary": row, "betInformation": bet, "relatedTransactions": s.providerTransactions(row.ProviderCode, row.ProviderRoundID, row.UserID, scope), "relatedRounds": s.roundByProvider(row.ProviderCode, row.ProviderRoundID, row.UserID, scope), "relatedDevicesAndIPs": s.userDeviceSummary(row.UserID, scope), "reviewHistory": history, "internalNotes": notes, "operationLogs": s.operationLogs("abnormal_bet", id, scope)}, nil
}

func (s *Service) CallbackDetail(id uint64, scope tenancy.Scope) (map[string]any, error) {
	var row map[string]any
	result := callbackJoins(tenancy.ApplySite(s.db.Table("game_provider_callback_logs c").Select(callbackSelectSQL(true)), "c", scope)).
		Where("c.id=? AND c.deleted_at IS NULL", id).Take(&row)
	if result.Error != nil {
		return nil, notFound(result.Error)
	}
	return map[string]any{"overview": safeMap(row, []string{"id", "callback_id", "provider_code", "operation", "user_no", "username", "user_phone", "provider_player_id", "provider_transaction_id", "provider_round_id", "game_id", "game_name", "provider_game_id", "amount_minor", "currency", "http_method", "request_path", "source_ip", "processing_result", "duration_ms", "correlation_id", "failure_reason", "created_at"}), "request": safeMap(row, []string{"http_method", "request_path", "request_headers", "query_parameters", "request_body", "source_ip"}), "validation": safeMap(row, []string{"signature_result", "timestamp_validation", "processing_result", "failure_reason"}), "normalizedCommand": jsonObject(row["normalized_command"]), "walletResult": jsonObject(row["wallet_result"]), "providerResponse": jsonObject(row["provider_response"]), "operationLogs": s.operationLogs("provider_callback", id, scope)}, nil
}

func (s *Service) singleBet(id uint64, scope tenancy.Scope) (BetRow, error) {
	var row BetRow
	db := s.db.Table("game_bet_records b").Select(baseBetSelect()).Joins("LEFT JOIN game_users u ON u.id=b.user_id").Joins("LEFT JOIN game_games g ON g.id=b.game_id").Where("b.id=? AND b.deleted_at IS NULL", id)
	db = tenancy.ApplySite(db, "b", scope).Take(&row)
	setBetMoney(&row)
	return row, db.Error
}
func (s *Service) rollbackRows(scope tenancy.Scope, where string, args ...any) ([]RollbackRow, error) {
	rows := []RollbackRow{}
	db := s.db.Table("game_rollbacks rb").Select("rb.id,rb.rollback_id,rb.original_provider_transaction_id,rb.provider_transaction_id,rb.provider_round_id,rb.user_id,u.user_no,u.username,rb.provider_code,rb.game_id,COALESCE(g.name,'') game_name,rb.original_operation,rb.original_amount_minor,rb.rollback_amount_minor,rb.currency,rb.status,rb.failure_reason,rb.created_at,rb.completed_at").Joins("LEFT JOIN game_users u ON u.id=rb.user_id").Joins("LEFT JOIN game_games g ON g.id=rb.game_id").Where(where, args...).Where("rb.deleted_at IS NULL")
	db = tenancy.ApplySite(db, "rb", scope).Scan(&rows)
	fillRollbackMoney(rows)
	return rows, db.Error
}
func (s *Service) abnormalRows(scope tenancy.Scope, where string, args ...any) ([]AbnormalRow, error) {
	rows := []AbnormalRow{}
	db := s.db.Table("game_abnormal_bet_events a").Select("a.id,a.risk_event_id,a.bet_record_id,a.user_id,u.user_no,u.username,a.provider_code,a.game_id,COALESCE(g.name,'') game_name,a.provider_round_id,a.risk_type,a.risk_level,a.review_status,a.bet_amount_minor,a.win_amount_minor,a.trigger_count,a.assigned_reviewer,a.latest_reason,a.created_at,a.updated_at").Joins("LEFT JOIN game_users u ON u.id=a.user_id").Joins("LEFT JOIN game_games g ON g.id=a.game_id").Where(where, args...).Where("a.deleted_at IS NULL")
	db = tenancy.ApplySite(db, "a", scope).Scan(&rows)
	fillAbnormalMoney(rows)
	return rows, db.Error
}
func (s *Service) providerTransactions(provider, round string, userID uint64, scope tenancy.Scope) []map[string]any {
	var rows []map[string]any
	db := s.db.Table("game_provider_transactions").Select("id,provider_code,provider_transaction_id,provider_round_id,provider_game_id,user_id,wallet_transaction_id,operation,amount_minor,currency,status,related_transaction_id,balance_before_minor,balance_after_minor,created_at").Where("provider_code=? AND provider_round_id=? AND user_id=? AND deleted_at IS NULL", provider, round, userID)
	_ = tenancy.ApplySite(db, "", scope).Order("created_at,id").Find(&rows).Error
	return rows
}
func (s *Service) providerTransaction(provider, transaction string, scope tenancy.Scope) map[string]any {
	var row map[string]any
	db := s.db.Table("game_provider_transactions").Select("id,provider_code,provider_transaction_id,provider_round_id,provider_game_id,user_id,wallet_transaction_id,operation,amount_minor,currency,status,related_transaction_id,balance_before_minor,balance_after_minor,created_at").Where("provider_code=? AND provider_transaction_id=? AND deleted_at IS NULL", provider, transaction)
	_ = tenancy.ApplySite(db, "", scope).Take(&row).Error
	return row
}
func (s *Service) walletTransactions(userID uint64, source string, sourceID uint64, scope tenancy.Scope) []map[string]any {
	var rows []map[string]any
	db := s.db.Table("game_wallet_transactions").Select("id,txn_no,external_txn_no,user_id,wallet_id,wallet_type,direction,category,source_type,source_id,provider_code,order_no,amount,currency,status,balance_before,balance_after,created_at,completed_at").Where("user_id=? AND deleted_at IS NULL", userID)
	db = tenancy.ApplySite(db, "", scope)
	if source != "" {
		db = db.Where("source_type=?", source)
	}
	if sourceID > 0 {
		db = db.Where("source_id=?", sourceID)
	}
	_ = db.Order("created_at,id").Find(&rows).Error
	return rows
}
func (s *Service) callbackLogs(provider, round string, userID uint64, scope tenancy.Scope) []map[string]any {
	var rows []map[string]any
	db := s.db.Table("game_provider_callback_logs").Select("id,callback_id,provider_code,operation,user_id,provider_transaction_id,provider_round_id,provider_game_id,amount_minor,currency,signature_result,timestamp_validation,processing_result,http_status,source_ip,duration_ms,failure_reason,created_at").Where("provider_code=? AND deleted_at IS NULL", provider)
	db = tenancy.ApplySite(db, "", scope)
	if round != "" {
		db = db.Where("provider_round_id=?", round)
	}
	if userID > 0 {
		db = db.Where("user_id=?", userID)
	}
	_ = db.Order("created_at,id").Find(&rows).Error
	return rows
}
func (s *Service) roundByProvider(provider, round string, userID uint64, scope tenancy.Scope) map[string]any {
	var row map[string]any
	db := s.db.Table("game_rounds").Where("provider_code=? AND provider_round_id=? AND user_id=? AND deleted_at IS NULL", provider, round, userID)
	_ = tenancy.ApplySite(db, "", scope).Take(&row).Error
	return row
}
func (s *Service) operationLogs(target string, id uint64, scope tenancy.Scope) []map[string]any {
	var rows []map[string]any
	db := s.db.Table("game_admin_operation_logs").Select("id,operator_id,operator_username,permission,action,before_state,after_state,reason,ip_address,correlation_id,created_at").Where("target_type=? AND target_id=? AND deleted_at IS NULL", target, id)
	_ = tenancy.ApplySite(db, "", scope).Order("created_at DESC").Find(&rows).Error
	return rows
}
func (s *Service) userDeviceSummary(userID uint64, scope tenancy.Scope) map[string]any {
	var devices, logins []map[string]any
	deviceDB := s.db.Table("game_user_devices").Select("id,device_id,platform,device_model,os_name,first_ip,last_ip,first_seen_at,last_seen_at,risk_status").Where("user_id=? AND deleted_at IS NULL", userID)
	_ = tenancy.ApplySite(deviceDB, "", scope).Find(&devices).Error
	loginDB := s.db.Table("game_user_login_records").Select("ip_address,country,region,device_id,platform,created_at").Where("user_id=? AND deleted_at IS NULL", userID)
	_ = tenancy.ApplySite(loginDB, "", scope).Order("created_at DESC").Limit(20).Find(&logins).Error
	return map[string]any{"devices": devices, "recentLogins": logins}
}
func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return newError("NOT_FOUND", "record not found")
	}
	return err
}
func safeMap(row map[string]any, keys []string) map[string]any {
	out := map[string]any{}
	for _, key := range keys {
		out[key] = row[key]
	}
	return out
}
func jsonObject(value any) any {
	switch typed := value.(type) {
	case string:
		var out any
		if json.Unmarshal([]byte(typed), &out) == nil {
			return out
		}
	case []byte:
		var out any
		if json.Unmarshal(typed, &out) == nil {
			return out
		}
	}
	return map[string]any{}
}

func (s *Service) MarkBetForReview(id uint64, command ReasonCommand, actor Actor) error {
	return s.createReviewEvent(id, 0, 0, "MANUAL_REVIEW", "MEDIUM", command.Reason, "bet:review", actor)
}
func (s *Service) MarkRollbackForInvestigation(id uint64, command ReasonCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var rollback model.GameRollbacks
		if err := tx.Where("id = ?", id).First(&rollback).Error; err != nil {
			return notFound(err)
		}
		var bet model.GameBetOrders
		if rollback.BetRecordID > 0 {
			_ = tx.Where("id = ?", rollback.BetRecordID).First(&bet).Error
		}
		event := model.GameAbnormalBetEvents{RiskEventID: randomID("RISK"), BetRecordID: bet.ID, UserID: rollback.UserID, ProviderCode: rollback.ProviderCode, GameID: rollback.GameID, ProviderRoundID: rollback.ProviderRoundID, RiskType: "FREQUENT_ROLLBACK", RiskLevel: "HIGH", ReviewStatus: "PENDING", BetAmountMinor: bet.BetAmountMinor, WinAmountMinor: bet.WinAmountMinor, LatestReason: command.Reason}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		return audit(tx, actor, "bet:rollback:review", "mark_for_investigation", "rollback", id, rollback.UserID, nil, map[string]any{"riskEventID": event.RiskEventID}, command.Reason)
	})
}
func (s *Service) MarkCallbackForInvestigation(id uint64, command ReasonCommand, actor Actor) error {
	var row model.GameProviderCallbackLogs
	if err := s.db.Where("id=?", id).First(&row).Error; err != nil {
		return notFound(err)
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&row).Update("marked_for_investigation", true).Error; err != nil {
			return err
		}
		event := model.GameAbnormalBetEvents{RiskEventID: randomID("RISK"), CallbackLogID: id, UserID: row.UserID, ProviderCode: row.ProviderCode, ProviderRoundID: row.ProviderRoundID, RiskType: "PROVIDER_CALLBACK_ANOMALY", RiskLevel: "HIGH", ReviewStatus: "PENDING", LatestReason: command.Reason}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		return audit(tx, actor, "bet:callback-log:view", "mark_callback_investigation", "provider_callback", id, row.UserID, map[string]any{"marked": false}, map[string]any{"marked": true, "riskEventID": event.RiskEventID}, command.Reason)
	})
}

func (s *Service) createReviewEvent(targetID, roundID, callbackID uint64, riskType, riskLevel, reason, permission string, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var bet model.GameBetOrders
		if targetID > 0 {
			if err := tx.Where("id=?", targetID).First(&bet).Error; err != nil {
				return notFound(err)
			}
		}
		event := model.GameAbnormalBetEvents{RiskEventID: randomID("RISK"), BetRecordID: bet.ID, RoundID: roundID, CallbackLogID: callbackID, UserID: bet.UserID, ProviderCode: bet.ProviderCode, GameID: bet.GameID, ProviderRoundID: bet.ProviderRoundID, RiskType: riskType, RiskLevel: riskLevel, ReviewStatus: "PENDING", BetAmountMinor: bet.BetAmountMinor, WinAmountMinor: bet.WinAmountMinor, LatestReason: reason}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, "mark_for_review", "bet", targetID, bet.UserID, nil, map[string]any{"riskEventID": event.RiskEventID}, reason)
	})
}

func (s *Service) AssignAbnormal(id uint64, command AssignCommand, actor Actor) error {
	return s.updateAbnormal(id, "ASSIGN", "", command.Reason, actor, func(row *model.GameAbnormalBetEvents) map[string]any {
		return map[string]any{"assigned_reviewer_id": command.ReviewerID, "assigned_reviewer": command.Reviewer, "review_status": "UNDER_REVIEW"}
	})
}
func (s *Service) ReviewAbnormal(id uint64, command ReviewCommand, actor Actor) error {
	allowed := map[string]bool{"UNDER_REVIEW": true, "CONFIRMED": true, "RESTRICTED": true}
	status := strings.ToUpper(strings.TrimSpace(command.Status))
	if !allowed[status] {
		return newError("INVALID_REVIEW_STATUS", "invalid review status")
	}
	return s.updateAbnormal(id, "REVIEW", status, command.Reason, actor, func(*model.GameAbnormalBetEvents) map[string]any { return map[string]any{"review_status": status} })
}
func (s *Service) ClearAbnormal(id uint64, command ReasonCommand, actor Actor) error {
	return s.updateAbnormal(id, "CLEAR", "CLEARED", command.Reason, actor, func(*model.GameAbnormalBetEvents) map[string]any { return map[string]any{"review_status": "CLEARED"} })
}
func (s *Service) RestrictAbnormal(id uint64, command ReasonCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameAbnormalBetEvents
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", id).First(&row).Error; err != nil {
			return notFound(err)
		}
		before := row.ReviewStatus
		if err := tx.Model(&row).Updates(map[string]any{"review_status": "RESTRICTED", "latest_reason": command.Reason}).Error; err != nil {
			return err
		}
		if err := tx.Table("game_users").Where("id=?", row.UserID).Updates(map[string]any{"game_access_enabled": 0, "risk_review_required": 1}).Error; err != nil {
			return err
		}
		if err := history(tx, &row, "RESTRICT", before, "RESTRICTED", command.Reason, actor); err != nil {
			return err
		}
		return audit(tx, actor, "bet:abnormal:restrict", "restrict_user", "abnormal_bet", id, row.UserID, map[string]any{"reviewStatus": before}, map[string]any{"reviewStatus": "RESTRICTED", "gameAccessEnabled": false}, command.Reason)
	})
}
func (s *Service) AddAbnormalNote(id uint64, command NoteCommand, actor Actor) error {
	content := strings.TrimSpace(command.Content)
	if content == "" {
		return newError("INVALID_REQUEST", "note content is required")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameAbnormalBetEvents
		if err := tx.Where("id=?", id).First(&row).Error; err != nil {
			return notFound(err)
		}
		note := model.GameAbnormalBetNotes{AbnormalBetID: id, OperatorID: actor.ID, OperatorUsername: actor.Username, Content: content}
		if err := tx.Create(&note).Error; err != nil {
			return err
		}
		return audit(tx, actor, "bet:abnormal:note", "add_note", "abnormal_bet", id, row.UserID, nil, map[string]any{"noteID": note.ID}, "internal note added")
	})
}
func (s *Service) updateAbnormal(id uint64, action, status, reason string, actor Actor, updates func(*model.GameAbnormalBetEvents) map[string]any) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameAbnormalBetEvents
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", id).First(&row).Error; err != nil {
			return notFound(err)
		}
		before := row.ReviewStatus
		values := updates(&row)
		values["latest_reason"] = reason
		if err := tx.Model(&row).Updates(values).Error; err != nil {
			return err
		}
		after := status
		if after == "" {
			if value, ok := values["review_status"].(string); ok {
				after = value
			} else {
				after = before
			}
		}
		if err := history(tx, &row, action, before, after, reason, actor); err != nil {
			return err
		}
		return audit(tx, actor, "bet:abnormal:review", strings.ToLower(action), "abnormal_bet", id, row.UserID, map[string]any{"reviewStatus": before}, map[string]any{"reviewStatus": after}, reason)
	})
}
func history(tx *gorm.DB, row *model.GameAbnormalBetEvents, action, before, after, reason string, actor Actor) error {
	return tx.Create(&model.GameAbnormalBetReviewHistories{TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID, AbnormalBetID: row.ID, Action: action, BeforeStatus: before, AfterStatus: after, OperatorID: actor.ID, OperatorUsername: actor.Username, Reason: reason}).Error
}
func audit(tx *gorm.DB, actor Actor, permission, action, target string, targetID, userID uint64, before, after any, reason string) error {
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)
	return tx.Create(&model.GameAdminOperationLogs{TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID, OperatorID: actor.ID, OperatorUsername: actor.Username, Permission: permission, Action: action, TargetType: target, TargetID: targetID, TargetUserID: userID, BeforeState: string(beforeJSON), AfterState: string(afterJSON), Reason: reason, IPAddress: actor.IP, UserAgent: actor.UserAgent, CorrelationID: actor.CorrelationID}).Error
}
func randomID(prefix string) string {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err == nil {
		return prefix + hex.EncodeToString(b)
	}
	return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano())
}
func parseUint(value string) uint64 {
	parsed, _ := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return parsed
}
