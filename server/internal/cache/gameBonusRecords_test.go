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

func newGameBonusRecordsCache() *gotest.Cache {
	record1 := &model.GameBonusRecords{}
	record1.ID = 1
	record2 := &model.GameBonusRecords{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameBonusRecordsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameBonusRecordsCache_Set(t *testing.T) {
	c := newGameBonusRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBonusRecords)
	err := c.ICache.(GameBonusRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameBonusRecordsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameBonusRecordsCache_Get(t *testing.T) {
	c := newGameBonusRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBonusRecords)
	err := c.ICache.(GameBonusRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameBonusRecordsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameBonusRecordsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameBonusRecordsCache_MultiGet(t *testing.T) {
	c := newGameBonusRecordsCache()
	defer c.Close()

	var testData []*model.GameBonusRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameBonusRecords))
	}

	err := c.ICache.(GameBonusRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameBonusRecordsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameBonusRecords))
	}
}

func Test_gameBonusRecordsCache_MultiSet(t *testing.T) {
	c := newGameBonusRecordsCache()
	defer c.Close()

	var testData []*model.GameBonusRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameBonusRecords))
	}

	err := c.ICache.(GameBonusRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameBonusRecordsCache_Del(t *testing.T) {
	c := newGameBonusRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBonusRecords)
	err := c.ICache.(GameBonusRecordsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameBonusRecordsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameBonusRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBonusRecords)
	err := c.ICache.(GameBonusRecordsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameBonusRecordsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameBonusRecordsCache(t *testing.T) {
	c := NewGameBonusRecordsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameBonusRecordsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameBonusRecordsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
