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

func newGamePaymentMethodsCache() *gotest.Cache {
	record1 := &model.GamePaymentMethods{}
	record1.ID = 1
	record2 := &model.GamePaymentMethods{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGamePaymentMethodsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gamePaymentMethodsCache_Set(t *testing.T) {
	c := newGamePaymentMethodsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GamePaymentMethods)
	err := c.ICache.(GamePaymentMethodsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GamePaymentMethodsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gamePaymentMethodsCache_Get(t *testing.T) {
	c := newGamePaymentMethodsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GamePaymentMethods)
	err := c.ICache.(GamePaymentMethodsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GamePaymentMethodsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GamePaymentMethodsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gamePaymentMethodsCache_MultiGet(t *testing.T) {
	c := newGamePaymentMethodsCache()
	defer c.Close()

	var testData []*model.GamePaymentMethods
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GamePaymentMethods))
	}

	err := c.ICache.(GamePaymentMethodsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GamePaymentMethodsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GamePaymentMethods))
	}
}

func Test_gamePaymentMethodsCache_MultiSet(t *testing.T) {
	c := newGamePaymentMethodsCache()
	defer c.Close()

	var testData []*model.GamePaymentMethods
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GamePaymentMethods))
	}

	err := c.ICache.(GamePaymentMethodsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gamePaymentMethodsCache_Del(t *testing.T) {
	c := newGamePaymentMethodsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GamePaymentMethods)
	err := c.ICache.(GamePaymentMethodsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gamePaymentMethodsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGamePaymentMethodsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GamePaymentMethods)
	err := c.ICache.(GamePaymentMethodsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GamePaymentMethodsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGamePaymentMethodsCache(t *testing.T) {
	c := NewGamePaymentMethodsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGamePaymentMethodsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGamePaymentMethodsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
