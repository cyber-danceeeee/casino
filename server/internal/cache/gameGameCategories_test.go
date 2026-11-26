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

func newGameGameCategoriesCache() *gotest.Cache {
	record1 := &model.GameGameCategories{}
	record1.ID = 1
	record2 := &model.GameGameCategories{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameGameCategoriesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameGameCategoriesCache_Set(t *testing.T) {
	c := newGameGameCategoriesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameCategories)
	err := c.ICache.(GameGameCategoriesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameGameCategoriesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameGameCategoriesCache_Get(t *testing.T) {
	c := newGameGameCategoriesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameCategories)
	err := c.ICache.(GameGameCategoriesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameGameCategoriesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameGameCategoriesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameGameCategoriesCache_MultiGet(t *testing.T) {
	c := newGameGameCategoriesCache()
	defer c.Close()

	var testData []*model.GameGameCategories
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameGameCategories))
	}

	err := c.ICache.(GameGameCategoriesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameGameCategoriesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameGameCategories))
	}
}

func Test_gameGameCategoriesCache_MultiSet(t *testing.T) {
	c := newGameGameCategoriesCache()
	defer c.Close()

	var testData []*model.GameGameCategories
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameGameCategories))
	}

	err := c.ICache.(GameGameCategoriesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGameCategoriesCache_Del(t *testing.T) {
	c := newGameGameCategoriesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameCategories)
	err := c.ICache.(GameGameCategoriesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGameCategoriesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameGameCategoriesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameCategories)
	err := c.ICache.(GameGameCategoriesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameGameCategoriesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameGameCategoriesCache(t *testing.T) {
	c := NewGameGameCategoriesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameGameCategoriesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameGameCategoriesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
