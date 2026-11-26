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

func newGameGameCategoryRelationsCache() *gotest.Cache {
	record1 := &model.GameGameCategoryRelations{}
	record1.ID = 1
	record2 := &model.GameGameCategoryRelations{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameGameCategoryRelationsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameGameCategoryRelationsCache_Set(t *testing.T) {
	c := newGameGameCategoryRelationsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameCategoryRelations)
	err := c.ICache.(GameGameCategoryRelationsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameGameCategoryRelationsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameGameCategoryRelationsCache_Get(t *testing.T) {
	c := newGameGameCategoryRelationsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameCategoryRelations)
	err := c.ICache.(GameGameCategoryRelationsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameGameCategoryRelationsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameGameCategoryRelationsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameGameCategoryRelationsCache_MultiGet(t *testing.T) {
	c := newGameGameCategoryRelationsCache()
	defer c.Close()

	var testData []*model.GameGameCategoryRelations
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameGameCategoryRelations))
	}

	err := c.ICache.(GameGameCategoryRelationsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameGameCategoryRelationsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameGameCategoryRelations))
	}
}

func Test_gameGameCategoryRelationsCache_MultiSet(t *testing.T) {
	c := newGameGameCategoryRelationsCache()
	defer c.Close()

	var testData []*model.GameGameCategoryRelations
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameGameCategoryRelations))
	}

	err := c.ICache.(GameGameCategoryRelationsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGameCategoryRelationsCache_Del(t *testing.T) {
	c := newGameGameCategoryRelationsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameCategoryRelations)
	err := c.ICache.(GameGameCategoryRelationsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGameCategoryRelationsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameGameCategoryRelationsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameCategoryRelations)
	err := c.ICache.(GameGameCategoryRelationsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameGameCategoryRelationsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameGameCategoryRelationsCache(t *testing.T) {
	c := NewGameGameCategoryRelationsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameGameCategoryRelationsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameGameCategoryRelationsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
