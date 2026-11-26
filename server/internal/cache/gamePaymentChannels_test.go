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

func newGamePaymentChannelsCache() *gotest.Cache {
	record1 := &model.GamePaymentChannels{}
	record1.ID = 1
	record2 := &model.GamePaymentChannels{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGamePaymentChannelsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gamePaymentChannelsCache_Set(t *testing.T) {
	c := newGamePaymentChannelsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GamePaymentChannels)
	err := c.ICache.(GamePaymentChannelsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GamePaymentChannelsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gamePaymentChannelsCache_Get(t *testing.T) {
	c := newGamePaymentChannelsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GamePaymentChannels)
	err := c.ICache.(GamePaymentChannelsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GamePaymentChannelsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GamePaymentChannelsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gamePaymentChannelsCache_MultiGet(t *testing.T) {
	c := newGamePaymentChannelsCache()
	defer c.Close()

	var testData []*model.GamePaymentChannels
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GamePaymentChannels))
	}

	err := c.ICache.(GamePaymentChannelsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GamePaymentChannelsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GamePaymentChannels))
	}
}

func Test_gamePaymentChannelsCache_MultiSet(t *testing.T) {
	c := newGamePaymentChannelsCache()
	defer c.Close()

	var testData []*model.GamePaymentChannels
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GamePaymentChannels))
	}

	err := c.ICache.(GamePaymentChannelsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gamePaymentChannelsCache_Del(t *testing.T) {
	c := newGamePaymentChannelsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GamePaymentChannels)
	err := c.ICache.(GamePaymentChannelsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gamePaymentChannelsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGamePaymentChannelsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GamePaymentChannels)
	err := c.ICache.(GamePaymentChannelsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GamePaymentChannelsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGamePaymentChannelsCache(t *testing.T) {
	c := NewGamePaymentChannelsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGamePaymentChannelsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGamePaymentChannelsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
