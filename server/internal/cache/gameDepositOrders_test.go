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

func newGameDepositOrdersCache() *gotest.Cache {
	record1 := &model.GameDepositOrders{}
	record1.ID = 1
	record2 := &model.GameDepositOrders{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameDepositOrdersCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameDepositOrdersCache_Set(t *testing.T) {
	c := newGameDepositOrdersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameDepositOrders)
	err := c.ICache.(GameDepositOrdersCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameDepositOrdersCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameDepositOrdersCache_Get(t *testing.T) {
	c := newGameDepositOrdersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameDepositOrders)
	err := c.ICache.(GameDepositOrdersCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameDepositOrdersCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameDepositOrdersCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameDepositOrdersCache_MultiGet(t *testing.T) {
	c := newGameDepositOrdersCache()
	defer c.Close()

	var testData []*model.GameDepositOrders
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameDepositOrders))
	}

	err := c.ICache.(GameDepositOrdersCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameDepositOrdersCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameDepositOrders))
	}
}

func Test_gameDepositOrdersCache_MultiSet(t *testing.T) {
	c := newGameDepositOrdersCache()
	defer c.Close()

	var testData []*model.GameDepositOrders
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameDepositOrders))
	}

	err := c.ICache.(GameDepositOrdersCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameDepositOrdersCache_Del(t *testing.T) {
	c := newGameDepositOrdersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameDepositOrders)
	err := c.ICache.(GameDepositOrdersCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameDepositOrdersCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameDepositOrdersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameDepositOrders)
	err := c.ICache.(GameDepositOrdersCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameDepositOrdersCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameDepositOrdersCache(t *testing.T) {
	c := NewGameDepositOrdersCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameDepositOrdersCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameDepositOrdersCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
