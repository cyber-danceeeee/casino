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

func newGameRankingPeriodsCache() *gotest.Cache {
	record1 := &model.GameRankingPeriods{}
	record1.ID = 1
	record2 := &model.GameRankingPeriods{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameRankingPeriodsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameRankingPeriodsCache_Set(t *testing.T) {
	c := newGameRankingPeriodsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRankingPeriods)
	err := c.ICache.(GameRankingPeriodsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameRankingPeriodsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameRankingPeriodsCache_Get(t *testing.T) {
	c := newGameRankingPeriodsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRankingPeriods)
	err := c.ICache.(GameRankingPeriodsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameRankingPeriodsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameRankingPeriodsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameRankingPeriodsCache_MultiGet(t *testing.T) {
	c := newGameRankingPeriodsCache()
	defer c.Close()

	var testData []*model.GameRankingPeriods
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameRankingPeriods))
	}

	err := c.ICache.(GameRankingPeriodsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameRankingPeriodsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameRankingPeriods))
	}
}

func Test_gameRankingPeriodsCache_MultiSet(t *testing.T) {
	c := newGameRankingPeriodsCache()
	defer c.Close()

	var testData []*model.GameRankingPeriods
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameRankingPeriods))
	}

	err := c.ICache.(GameRankingPeriodsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameRankingPeriodsCache_Del(t *testing.T) {
	c := newGameRankingPeriodsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRankingPeriods)
	err := c.ICache.(GameRankingPeriodsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameRankingPeriodsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameRankingPeriodsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRankingPeriods)
	err := c.ICache.(GameRankingPeriodsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameRankingPeriodsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameRankingPeriodsCache(t *testing.T) {
	c := NewGameRankingPeriodsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameRankingPeriodsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameRankingPeriodsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
