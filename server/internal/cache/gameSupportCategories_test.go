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

func newGameSupportCategoriesCache() *gotest.Cache {
	record1 := &model.GameSupportCategories{}
	record1.ID = 1
	record2 := &model.GameSupportCategories{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameSupportCategoriesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameSupportCategoriesCache_Set(t *testing.T) {
	c := newGameSupportCategoriesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportCategories)
	err := c.ICache.(GameSupportCategoriesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameSupportCategoriesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameSupportCategoriesCache_Get(t *testing.T) {
	c := newGameSupportCategoriesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportCategories)
	err := c.ICache.(GameSupportCategoriesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameSupportCategoriesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameSupportCategoriesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameSupportCategoriesCache_MultiGet(t *testing.T) {
	c := newGameSupportCategoriesCache()
	defer c.Close()

	var testData []*model.GameSupportCategories
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameSupportCategories))
	}

	err := c.ICache.(GameSupportCategoriesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameSupportCategoriesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameSupportCategories))
	}
}

func Test_gameSupportCategoriesCache_MultiSet(t *testing.T) {
	c := newGameSupportCategoriesCache()
	defer c.Close()

	var testData []*model.GameSupportCategories
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameSupportCategories))
	}

	err := c.ICache.(GameSupportCategoriesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameSupportCategoriesCache_Del(t *testing.T) {
	c := newGameSupportCategoriesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportCategories)
	err := c.ICache.(GameSupportCategoriesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameSupportCategoriesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameSupportCategoriesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportCategories)
	err := c.ICache.(GameSupportCategoriesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameSupportCategoriesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameSupportCategoriesCache(t *testing.T) {
	c := NewGameSupportCategoriesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameSupportCategoriesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameSupportCategoriesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
