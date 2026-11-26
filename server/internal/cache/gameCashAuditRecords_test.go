package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-dev-frame/sponge/pkg/gotest"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"casino/internal/database"
	"casino/internal/model"
)

func newGameCashAuditRecordsCache() *gotest.Cache {
	record1 := &model.GameCashAuditRecords{}
	record1.ID = 1
	record2 := &model.GameCashAuditRecords{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameCashAuditRecordsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameCashAuditRecordsCache_Set(t *testing.T) {
	c := newGameCashAuditRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameCashAuditRecords)
	err := c.ICache.(GameCashAuditRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameCashAuditRecordsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameCashAuditRecordsCache_Get(t *testing.T) {
	c := newGameCashAuditRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameCashAuditRecords)
	err := c.ICache.(GameCashAuditRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameCashAuditRecordsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameCashAuditRecordsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameCashAuditRecordsCache_MultiGet(t *testing.T) {
	c := newGameCashAuditRecordsCache()
	defer c.Close()

	var testData []*model.GameCashAuditRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameCashAuditRecords))
	}

	err := c.ICache.(GameCashAuditRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameCashAuditRecordsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameCashAuditRecords))
	}
}

func Test_gameCashAuditRecordsCache_MultiSet(t *testing.T) {
	c := newGameCashAuditRecordsCache()
	defer c.Close()

	var testData []*model.GameCashAuditRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameCashAuditRecords))
	}

	err := c.ICache.(GameCashAuditRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameCashAuditRecordsCache_Del(t *testing.T) {
	c := newGameCashAuditRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameCashAuditRecords)
	err := c.ICache.(GameCashAuditRecordsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameCashAuditRecordsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameCashAuditRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameCashAuditRecords)
	err := c.ICache.(GameCashAuditRecordsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameCashAuditRecordsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameCashAuditRecordsCache(t *testing.T) {
	c := NewGameCashAuditRecordsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameCashAuditRecordsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameCashAuditRecordsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
