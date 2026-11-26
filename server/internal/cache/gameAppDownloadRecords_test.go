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

func newGameAppDownloadRecordsCache() *gotest.Cache {
	record1 := &model.GameAppDownloadRecords{}
	record1.ID = 1
	record2 := &model.GameAppDownloadRecords{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameAppDownloadRecordsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameAppDownloadRecordsCache_Set(t *testing.T) {
	c := newGameAppDownloadRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAppDownloadRecords)
	err := c.ICache.(GameAppDownloadRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameAppDownloadRecordsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameAppDownloadRecordsCache_Get(t *testing.T) {
	c := newGameAppDownloadRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAppDownloadRecords)
	err := c.ICache.(GameAppDownloadRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameAppDownloadRecordsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameAppDownloadRecordsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameAppDownloadRecordsCache_MultiGet(t *testing.T) {
	c := newGameAppDownloadRecordsCache()
	defer c.Close()

	var testData []*model.GameAppDownloadRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameAppDownloadRecords))
	}

	err := c.ICache.(GameAppDownloadRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameAppDownloadRecordsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameAppDownloadRecords))
	}
}

func Test_gameAppDownloadRecordsCache_MultiSet(t *testing.T) {
	c := newGameAppDownloadRecordsCache()
	defer c.Close()

	var testData []*model.GameAppDownloadRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameAppDownloadRecords))
	}

	err := c.ICache.(GameAppDownloadRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameAppDownloadRecordsCache_Del(t *testing.T) {
	c := newGameAppDownloadRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAppDownloadRecords)
	err := c.ICache.(GameAppDownloadRecordsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameAppDownloadRecordsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameAppDownloadRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAppDownloadRecords)
	err := c.ICache.(GameAppDownloadRecordsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameAppDownloadRecordsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameAppDownloadRecordsCache(t *testing.T) {
	c := NewGameAppDownloadRecordsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameAppDownloadRecordsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameAppDownloadRecordsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
