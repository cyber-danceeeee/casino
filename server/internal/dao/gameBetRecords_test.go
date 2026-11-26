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

func newGameBetRecordsDao() *gotest.Dao {
	testData := &model.GameBetRecords{}
	testData.ID = 1
	// you can set the other fields of testData here, such as:
	//testData.CreatedAt = time.Now()
	//testData.UpdatedAt = testData.CreatedAt

	// init mock cache
	//c := gotest.NewCache(map[string]interface{}{"no cache": testData}) // to test mysql, disable caching
	c := gotest.NewCache(map[string]interface{}{utils.Uint64ToStr(testData.ID): testData})
	c.ICache = cache.NewGameBetRecordsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})

	// init mock dao
	d := gotest.NewDao(c, testData)
	d.IDao = NewGameBetRecordsDao(d.DB, c.ICache.(cache.GameBetRecordsCache))

	return d
}

func Test_gameBetRecordsDao_Create(t *testing.T) {
	d := newGameBetRecordsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameBetRecords)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(GameBetRecordsDao).Create(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameBetRecordsDao_DeleteByID(t *testing.T) {
	d := newGameBetRecordsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameBetRecords)
	expectedSQLForDeletion := "UPDATE .*"
	expectedArgsForDeletionTime := d.AnyTime

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(expectedArgsForDeletionTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(GameBetRecordsDao).DeleteByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(GameBetRecordsDao).DeleteByID(d.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameBetRecordsDao_UpdateByID(t *testing.T) {
	d := newGameBetRecordsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameBetRecords)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WithArgs(d.AnyTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(GameBetRecordsDao).UpdateByID(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(GameBetRecordsDao).UpdateByID(d.Ctx, &model.GameBetRecords{})
	assert.Error(t, err)

}

func Test_gameBetRecordsDao_GetByID(t *testing.T) {
	d := newGameBetRecordsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameBetRecords)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID, 1).
		WillReturnRows(rows)

	_, err := d.IDao.(GameBetRecordsDao).GetByID(d.Ctx, testData.ID)
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
	_, err = d.IDao.(GameBetRecordsDao).GetByID(d.Ctx, 2)
	assert.Error(t, err)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(3, 4).
		WillReturnRows(rows)
	_, err = d.IDao.(GameBetRecordsDao).GetByID(d.Ctx, 4)
	assert.Error(t, err)
}

func Test_gameBetRecordsDao_GetByColumns(t *testing.T) {
	d := newGameBetRecordsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameBetRecords)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	_, _, err := d.IDao.(GameBetRecordsDao).GetByColumns(d.Ctx, &query.Params{
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
	_, _, err = d.IDao.(GameBetRecordsDao).GetByColumns(d.Ctx, &query.Params{
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
	dao := &gameBetRecordsDao{}
	_, _, err = dao.GetByColumns(context.Background(), &query.Params{Columns: []query.Column{{}}})
	t.Log(err)
}

func Test_gameBetRecordsDao_CreateByTx(t *testing.T) {
	d := newGameBetRecordsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameBetRecords)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	_, err := d.IDao.(GameBetRecordsDao).CreateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameBetRecordsDao_DeleteByTx(t *testing.T) {
	d := newGameBetRecordsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameBetRecords)
	expectedSQLForDeletion := "UPDATE .*"
	expectedArgsForDeletionTime := d.AnyTime

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(expectedArgsForDeletionTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(GameBetRecordsDao).DeleteByTx(d.Ctx, d.DB, testData.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameBetRecordsDao_UpdateByTx(t *testing.T) {
	d := newGameBetRecordsDao()
	defer d.Close()
	testData := d.TestData.(*model.GameBetRecords)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WithArgs(d.AnyTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(GameBetRecordsDao).UpdateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}
