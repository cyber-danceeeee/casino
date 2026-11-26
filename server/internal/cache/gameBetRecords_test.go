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

func newGameBetRecordsCache() *gotest.Cache {
	record1 := &model.GameBetRecords{}
	record1.ID = 1
	record2 := &model.GameBetRecords{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameBetRecordsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameBetRecordsCache_Set(t *testing.T) {
	c := newGameBetRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBetRecords)
	err := c.ICache.(GameBetRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameBetRecordsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameBetRecordsCache_Get(t *testing.T) {
	c := newGameBetRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBetRecords)
	err := c.ICache.(GameBetRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameBetRecordsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameBetRecordsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameBetRecordsCache_MultiGet(t *testing.T) {
	c := newGameBetRecordsCache()
	defer c.Close()

	var testData []*model.GameBetRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameBetRecords))
	}

	err := c.ICache.(GameBetRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameBetRecordsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameBetRecords))
	}
}

func Test_gameBetRecordsCache_MultiSet(t *testing.T) {
	c := newGameBetRecordsCache()
	defer c.Close()

	var testData []*model.GameBetRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameBetRecords))
	}

	err := c.ICache.(GameBetRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameBetRecordsCache_Del(t *testing.T) {
	c := newGameBetRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBetRecords)
	err := c.ICache.(GameBetRecordsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameBetRecordsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameBetRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBetRecords)
	err := c.ICache.(GameBetRecordsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameBetRecordsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameBetRecordsCache(t *testing.T) {
	c := NewGameBetRecordsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameBetRecordsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameBetRecordsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
