package dao

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-dev-frame/sponge/pkg/gotest"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
	"github.com/go-dev-frame/sponge/pkg/utils"
	"github.com/stretchr/testify/assert"

	"casino/internal/cache"
	"casino/internal/database"
	"casino/internal/model"
)

func newGameGameLaunchLogsDao() *gotest.Dao {
	testData := &model.GameGameLaunchLogs{}
	testData.ID = 1
	// you can set the other fields of testData here, such as:
	//testData.CreatedAt = time.Now()
	//testData.UpdatedAt = testData.CreatedAt

	// init mock cache
	//c := gotest.NewCache(map[string]interface{}{"no cache": testData}) // to test mysql, disable caching
	c := gotest.NewCache(map[string]interface{}{utils.Uint64ToStr(testData.ID): testData})
	c.ICache = cache.NewGameGameLaunchLogsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})

	// init mock dao
	d := gotest.NewDao(c, testData)
	d.IDao = NewGameGameLaunchLogsDao(d.DB, c.ICache.(cache.GameGameLaunchLogsCache))

	return d
}

func Test_gameGameLaunchLogsDao_Create(t *testing.T) {
	d := newGameGameLaunchLogsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameGameLaunchLogs)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(GameGameLaunchLogsDao).Create(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGameLaunchLogsDao_DeleteByID(t *testing.T) {
	d := newGameGameLaunchLogsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameGameLaunchLogs)
	expectedSQLForDeletion := "UPDATE .*"
	expectedArgsForDeletionTime := d.AnyTime

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(expectedArgsForDeletionTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(GameGameLaunchLogsDao).DeleteByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(GameGameLaunchLogsDao).DeleteByID(d.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameGameLaunchLogsDao_UpdateByID(t *testing.T) {
	d := newGameGameLaunchLogsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameGameLaunchLogs)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WithArgs(d.AnyTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(GameGameLaunchLogsDao).UpdateByID(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(GameGameLaunchLogsDao).UpdateByID(d.Ctx, &model.GameGameLaunchLogs{})
	assert.Error(t, err)
	
}

func Test_gameGameLaunchLogsDao_GetByID(t *testing.T) {
	d := newGameGameLaunchLogsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameGameLaunchLogs)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID, 1).
		WillReturnRows(rows)

	_, err := d.IDao.(GameGameLaunchLogsDao).GetByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// error test
	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(2).
		WillReturnRows(rows)
	_, err = d.IDao.(GameGameLaunchLogsDao).GetByID(d.Ctx, 2)
	assert.Error(t, err)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(3, 4).
		WillReturnRows(rows)
	_, err = d.IDao.(GameGameLaunchLogsDao).GetByID(d.Ctx, 4)
	assert.Error(t, err)
}

func Test_gameGameLaunchLogsDao_GetByColumns(t *testing.T) {
	d := newGameGameLaunchLogsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameGameLaunchLogs)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	_, _, err := d.IDao.(GameGameLaunchLogsDao).GetByColumns(d.Ctx, &query.Params{
		Page:  0,
		Limit: 10,
		Sort:  "ignore count", // ignore test count(*)
	})
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// err test
	_, _, err = d.IDao.(GameGameLaunchLogsDao).GetByColumns(d.Ctx, &query.Params{
		Page:  0,
		Limit: 10,
		Columns: []query.Column{
			{
				Name:  "id",
				Exp:   "<",
				Value: 0,
			},
		},
	})
	assert.Error(t, err)

	// error test
	dao := &gameGameLaunchLogsDao{}
	_, _, err = dao.GetByColumns(context.Background(), &query.Params{Columns: []query.Column{{}}})
	t.Log(err)
}

func Test_gameGameLaunchLogsDao_CreateByTx(t *testing.T) {
	d := newGameGameLaunchLogsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameGameLaunchLogs)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	_, err := d.IDao.(GameGameLaunchLogsDao).CreateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGameLaunchLogsDao_DeleteByTx(t *testing.T) {
	d := newGameGameLaunchLogsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameGameLaunchLogs)
	expectedSQLForDeletion := "UPDATE .*"
	expectedArgsForDeletionTime := d.AnyTime

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(expectedArgsForDeletionTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(GameGameLaunchLogsDao).DeleteByTx(d.Ctx, d.DB, testData.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGameLaunchLogsDao_UpdateByTx(t *testing.T) {
	d := newGameGameLaunchLogsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameGameLaunchLogs)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WithArgs(d.AnyTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(GameGameLaunchLogsDao).UpdateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}
