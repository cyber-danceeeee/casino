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

func newGameRankingRecordsCache() *gotest.Cache {
	record1 := &model.GameRankingRecords{}
	record1.ID = 1
	record2 := &model.GameRankingRecords{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameRankingRecordsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameRankingRecordsCache_Set(t *testing.T) {
	c := newGameRankingRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRankingRecords)
	err := c.ICache.(GameRankingRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameRankingRecordsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameRankingRecordsCache_Get(t *testing.T) {
	c := newGameRankingRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRankingRecords)
	err := c.ICache.(GameRankingRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameRankingRecordsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameRankingRecordsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameRankingRecordsCache_MultiGet(t *testing.T) {
	c := newGameRankingRecordsCache()
	defer c.Close()

	var testData []*model.GameRankingRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameRankingRecords))
	}

	err := c.ICache.(GameRankingRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameRankingRecordsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameRankingRecords))
	}
}

func Test_gameRankingRecordsCache_MultiSet(t *testing.T) {
	c := newGameRankingRecordsCache()
	defer c.Close()

	var testData []*model.GameRankingRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameRankingRecords))
	}

	err := c.ICache.(GameRankingRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameRankingRecordsCache_Del(t *testing.T) {
	c := newGameRankingRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRankingRecords)
	err := c.ICache.(GameRankingRecordsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameRankingRecordsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameRankingRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRankingRecords)
	err := c.ICache.(GameRankingRecordsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameRankingRecordsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameRankingRecordsCache(t *testing.T) {
	c := NewGameRankingRecordsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameRankingRecordsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameRankingRecordsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
