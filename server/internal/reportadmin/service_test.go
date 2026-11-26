package reportadmin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"casino/internal/content"
	"casino/internal/eventbus"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPercentageZeroDenominator(t *testing.T) {
	require.Nil(t, percentage(decimal.NewFromInt(12), decimal.Zero))
}

func TestPercentageUsesPreciseDecimal(t *testing.T) {
	value := percentage(decimal.NewFromInt(1), decimal.NewFromInt(3))
	require.NotNil(t, value)
	require.Equal(t, "33.3333", *value)
}

func TestPaginateRows(t *testing.T) {
	rows, total, page, size := paginateRows([]int{1, 2, 3, 4, 5}, Query{Page: 2, PageSize: 2})
	require.Equal(t, []int{3, 4}, rows)
	require.Equal(t, 5, total)
	require.Equal(t, 2, page)
	require.Equal(t, 2, size)
}

func TestTimezoneOffset(t *testing.T) {
	delta, zone, err := timezoneOffset("+07:00")
	require.NoError(t, err)
	require.Equal(t, -60, delta)
	require.Equal(t, "+07:00", zone)
}

func TestSensitiveExportFieldsAreExcluded(t *testing.T) {
	require.True(t, sensitiveExportKey("accessToken"))
	require.True(t, sensitiveExportKey("rawMetadata"))
	require.False(t, sensitiveExportKey("totalBet"))
}

func TestReportTypeValidation(t *testing.T) {
	value, err := reportType("financial")
	require.NoError(t, err)
	require.Equal(t, "FINANCIAL", value)
	_, err = reportType("wallet-secret")
	require.Error(t, err)
}

func TestDailySummaryTablesAreUniqueAndPrefixed(t *testing.T) {
	seen := map[string]bool{}
	for _, table := range dailySummaryTables() {
		require.Contains(t, table, "game_daily_")
		require.False(t, seen[table])
		seen[table] = true
	}
	require.Len(t, seen, 7)
}

func TestFinancialChartAlignsCurrenciesAcrossPeriods(t *testing.T) {
	chart := financialChart("cashFlow", []FinancialRow{
		{Period: "2026-07-01", Currency: "PKR", DepositAmount: decimal.NewFromInt(100)},
		{Period: "2026-07-02", Currency: "USD", DepositAmount: decimal.NewFromInt(20)},
	}, []string{"depositAmount"})

	require.Equal(t, []string{"2026-07-01", "2026-07-02"}, chart.Categories)
	require.Len(t, chart.Series, 2)
	require.Equal(t, []string{"100.00", "0.00"}, chart.Series[0].Data)
	require.Equal(t, []string{"0.00", "20.00"}, chart.Series[1].Data)
}

func TestReportScopeSQLAddsH5AndSite(t *testing.T) {
	where, args := reportScopeSQL(Query{Scope: content.Scope{SiteID: 2}}, "r")
	require.Equal(t, " AND r.site_id IN (0, ?)", where)
	require.Equal(t, []any{uint64(2)}, args)
}

func TestFinancialFactsScopeUsesFactTableAlias(t *testing.T) {
	where, args := userFactWhere(Query{Scope: content.Scope{SiteID: 4}, Values: map[string]string{"dateFrom": "2026-07-01", "dateTo": "2026-07-02"}}, "d.paid_at", "d.currency", "u", "d", true)

	require.Contains(t, where, "d.site_id IN (0, ?)")

	require.Equal(t, []any{"2026-07-01", "2026-07-02", uint64(4)}, args)
}

func TestExportWorkerScopeComesFromTaskRecord(t *testing.T) {
	filters := map[string]string{"siteID": "888", "currency": "PKR"}
	scope := exportTaskScope(3)
	q := exportPageQuery(filters, scope)

	require.Equal(t, uint64(3), q.Scope.SiteID)
	require.False(t, q.Scope.Platform)
	require.Equal(t, "888", q.Values["siteID"])
}

func TestExportTaskScopePlatformWhenH5IsZero(t *testing.T) {
	scope := exportTaskScope(0)
	require.True(t, scope.Platform)
	require.Equal(t, uint64(0), scope.SiteID)
	require.Equal(t, uint64(0), scope.SiteID)
}

func TestExportTaskLookupIncludesH5ScopeForIDOR(t *testing.T) {
	db := dryRunMySQL(t)
	stmt := exportTaskDB(db, content.Scope{SiteID: 3}).
		Where("id=? AND deleted_at IS NULL", uint64(99)).
		Find(&[]ExportRow{}).Statement

	sql := stmt.SQL.String()
	require.Contains(t, sql, "FROM `game_report_export_tasks`")

	require.Contains(t, sql, "site_id IN (0, ?)")
	require.Equal(t, []any{uint64(3), uint64(99)}, stmt.Vars)
}

func TestPlatformExportTaskLookupIsUnscoped(t *testing.T) {
	db := dryRunMySQL(t)
	stmt := exportTaskDB(db, content.Scope{Platform: true}).
		Where("id=? AND deleted_at IS NULL", uint64(99)).
		Find(&[]ExportRow{}).Statement

	sql := stmt.SQL.String()

	require.NotContains(t, sql, "site_id IN")
	require.Equal(t, []any{uint64(99)}, stmt.Vars)
}

func TestDailySummaryIndexesIncludeSiteAndCurrency(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "database", "schema.sql"))
	require.NoError(t, err)
	schema := string(data)
	for _, table := range dailySummaryTables() {
		require.Contains(t, schema, "CREATE TABLE `"+table+"`")
	}
	require.Contains(t, schema, "`site_id`,`business_date`,`currency`")
}

func TestReportExportRetryPublishesOutboxEvent(t *testing.T) {
	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	require.NoError(t, eventbus.Migrate(db))
	require.NoError(t, db.Exec("CREATE TABLE game_report_export_tasks (id INTEGER PRIMARY KEY, site_id INTEGER NOT NULL, status TEXT NOT NULL, failure_reason TEXT NOT NULL DEFAULT '', completed_at DATETIME, updated_at DATETIME, deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("CREATE TABLE game_admin_operation_logs (id INTEGER PRIMARY KEY AUTOINCREMENT, site_id INTEGER, operator_id INTEGER, operator_username TEXT, permission TEXT, action TEXT, target_type TEXT, target_id INTEGER, before_state TEXT, after_state TEXT, reason TEXT, ip_address TEXT, user_agent TEXT, correlation_id TEXT, currency TEXT, created_at DATETIME, updated_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO game_report_export_tasks (id, site_id, status) VALUES (91, 3, 'FAILED')").Error)

	actor := Actor{ID: 12, Username: "finance", CorrelationID: "retry-trace", Scope: content.Scope{SiteID: 3}}
	require.NoError(t, New(db).Export.Retry(91, "retry after storage recovery", actor))

	var status string
	require.NoError(t, db.Table("game_report_export_tasks").Select("status").Where("id=91").Scan(&status).Error)
	require.Equal(t, "PENDING", status)
	var count int64
	require.NoError(t, db.Table("event_outbox").Where("event_type = ? AND aggregate_id = ? AND site_id = ?", eventbus.EventReportExportRequested, "91", 3).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func dryRunMySQL(t *testing.T) *gorm.DB {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	mock.ExpectClose()
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DryRun: true})
	require.NoError(t, err)
	return db
}
