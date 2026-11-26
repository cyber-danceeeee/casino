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

func newGameAgentCommissionRecordsCache() *gotest.Cache {
	record1 := &model.GameAgentCommissionRecords{}
	record1.ID = 1
	record2 := &model.GameAgentCommissionRecords{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameAgentCommissionRecordsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameAgentCommissionRecordsCache_Set(t *testing.T) {
	c := newGameAgentCommissionRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAgentCommissionRecords)
	err := c.ICache.(GameAgentCommissionRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameAgentCommissionRecordsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameAgentCommissionRecordsCache_Get(t *testing.T) {
	c := newGameAgentCommissionRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAgentCommissionRecords)
	err := c.ICache.(GameAgentCommissionRecordsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameAgentCommissionRecordsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameAgentCommissionRecordsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameAgentCommissionRecordsCache_MultiGet(t *testing.T) {
	c := newGameAgentCommissionRecordsCache()
	defer c.Close()

	var testData []*model.GameAgentCommissionRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameAgentCommissionRecords))
	}

	err := c.ICache.(GameAgentCommissionRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameAgentCommissionRecordsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameAgentCommissionRecords))
	}
}

func Test_gameAgentCommissionRecordsCache_MultiSet(t *testing.T) {
	c := newGameAgentCommissionRecordsCache()
	defer c.Close()

	var testData []*model.GameAgentCommissionRecords
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameAgentCommissionRecords))
	}

	err := c.ICache.(GameAgentCommissionRecordsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameAgentCommissionRecordsCache_Del(t *testing.T) {
	c := newGameAgentCommissionRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAgentCommissionRecords)
	err := c.ICache.(GameAgentCommissionRecordsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameAgentCommissionRecordsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameAgentCommissionRecordsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAgentCommissionRecords)
	err := c.ICache.(GameAgentCommissionRecordsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameAgentCommissionRecordsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameAgentCommissionRecordsCache(t *testing.T) {
	c := NewGameAgentCommissionRecordsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameAgentCommissionRecordsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameAgentCommissionRecordsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
