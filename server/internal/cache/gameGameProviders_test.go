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

func newGameGameProvidersCache() *gotest.Cache {
	record1 := &model.GameGameProviders{}
	record1.ID = 1
	record2 := &model.GameGameProviders{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameGameProvidersCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameGameProvidersCache_Set(t *testing.T) {
	c := newGameGameProvidersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameProviders)
	err := c.ICache.(GameGameProvidersCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameGameProvidersCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameGameProvidersCache_Get(t *testing.T) {
	c := newGameGameProvidersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameProviders)
	err := c.ICache.(GameGameProvidersCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameGameProvidersCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameGameProvidersCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameGameProvidersCache_MultiGet(t *testing.T) {
	c := newGameGameProvidersCache()
	defer c.Close()

	var testData []*model.GameGameProviders
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameGameProviders))
	}

	err := c.ICache.(GameGameProvidersCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameGameProvidersCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameGameProviders))
	}
}

func Test_gameGameProvidersCache_MultiSet(t *testing.T) {
	c := newGameGameProvidersCache()
	defer c.Close()

	var testData []*model.GameGameProviders
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameGameProviders))
	}

	err := c.ICache.(GameGameProvidersCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGameProvidersCache_Del(t *testing.T) {
	c := newGameGameProvidersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameProviders)
	err := c.ICache.(GameGameProvidersCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGameProvidersCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameGameProvidersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameProviders)
	err := c.ICache.(GameGameProvidersCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameGameProvidersCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameGameProvidersCache(t *testing.T) {
	c := NewGameGameProvidersCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameGameProvidersCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameGameProvidersCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
