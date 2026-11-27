package systemadmin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLoginLogResultsAndFilters(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, _ := db.DB()
	t.Cleanup(func() { raw.Close() })
	require.NoError(t, db.Exec(`CREATE TABLE game_admin_login_logs(id INTEGER PRIMARY KEY,username TEXT,success INTEGER,created_at DATETIME,deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO game_admin_login_logs VALUES(1,'test',1,CURRENT_TIMESTAMP,NULL),(2,'test',0,CURRENT_TIMESTAMP,NULL)`).Error)
	svc := New(db)
	all, err := svc.ListLoginLogs(Query{})
	require.NoError(t, err)
	rows := all.List.([]LoginLogRow)
	require.Len(t, rows, 2)
	require.Equal(t, "FAILED", rows[0].LoginResult)
	require.Equal(t, "SUCCESS", rows[1].LoginResult)
	filtered, err := svc.ListLoginLogs(Query{Values: map[string]string{"loginResult": "FAILED"}})
	require.NoError(t, err)
	require.Equal(t, int64(1), filtered.Total)
	data, err := svc.LoginLogDetail(1)
	require.NoError(t, err)
	require.Equal(t, "SUCCESS", data["overview"].(LoginLogRow).LoginResult)
}
func TestAuditLogListQueryAndDetailRedaction(t *testing.T) {
	svc, db := newSiteContentTestService(t)
	require.NoError(t, db.Exec(`ALTER TABLE game_admin_operation_logs ADD COLUMN request_info TEXT`).Error)
	require.NoError(t, db.Exec(`ALTER TABLE game_admin_operation_logs ADD COLUMN error_info TEXT`).Error)
	require.NoError(t, db.Exec(`INSERT INTO game_admin_operation_logs(operator_username,permission,module,action,result,created_at,before_state,after_state) VALUES('operator','game:update','','save','SUCCESS',CURRENT_TIMESTAMP,'{"password":"secret"}','{"enabled":true}')`).Error)
	result, err := svc.ListAuditLogs(Query{Values: map[string]string{"module": "game"}})
	require.NoError(t, err)
	rows := result.List.(*[]AuditLogRow)
	require.Len(t, *rows, 1)
	require.Equal(t, "game", (*rows)[0].Module)
	detail, err := svc.AuditLogDetail((*rows)[0].ID)
	require.NoError(t, err)
	encoded, err := json.Marshal(detail)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "secret")
	require.Contains(t, string(encoded), "enabled")
}
