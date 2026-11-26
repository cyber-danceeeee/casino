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

func newGameUserActivityRecordsCache() *gotest.Cache {
	record1 := &model.GameUserActivityRecords{}
	record1.ID = 1
	record2 := &model.GameUserActivityRecords{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameUserActivityRecordsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameUserActivityRecordsCache_Set(t *testing.T) {
	c := newGameUserActivityRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserActivityRecords)
	err := c.ICache.(GameUserActivityRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameUserActivityRecordsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameUserActivityRecordsCache_Get(t *testing.T) {
	c := newGameUserActivityRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserActivityRecords)
	err := c.ICache.(GameUserActivityRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserActivityRecordsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameUserActivityRecordsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameUserActivityRecordsCache_MultiGet(t *testing.T) {
	c := newGameUserActivityRecordsCache()
	defer c.Close()

	var testData []*model.GameUserActivityRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserActivityRecords))
	}

	err := c.ICache.(GameUserActivityRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserActivityRecordsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameUserActivityRecords))
	}
}

func Test_gameUserActivityRecordsCache_MultiSet(t *testing.T) {
	c := newGameUserActivityRecordsCache()
	defer c.Close()

	var testData []*model.GameUserActivityRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserActivityRecords))
	}

	err := c.ICache.(GameUserActivityRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserActivityRecordsCache_Del(t *testing.T) {
	c := newGameUserActivityRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserActivityRecords)
	err := c.ICache.(GameUserActivityRecordsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserActivityRecordsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameUserActivityRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserActivityRecords)
	err := c.ICache.(GameUserActivityRecordsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameUserActivityRecordsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameUserActivityRecordsCache(t *testing.T) {
	c := NewGameUserActivityRecordsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameUserActivityRecordsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameUserActivityRecordsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
