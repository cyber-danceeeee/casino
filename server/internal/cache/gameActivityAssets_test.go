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

func newGameActivityAssetsCache() *gotest.Cache {
	record1 := &model.GameActivityAssets{}
	record1.ID = 1
	record2 := &model.GameActivityAssets{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameActivityAssetsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameActivityAssetsCache_Set(t *testing.T) {
	c := newGameActivityAssetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameActivityAssets)
	err := c.ICache.(GameActivityAssetsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameActivityAssetsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameActivityAssetsCache_Get(t *testing.T) {
	c := newGameActivityAssetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameActivityAssets)
	err := c.ICache.(GameActivityAssetsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameActivityAssetsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameActivityAssetsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameActivityAssetsCache_MultiGet(t *testing.T) {
	c := newGameActivityAssetsCache()
	defer c.Close()

	var testData []*model.GameActivityAssets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameActivityAssets))
	}

	err := c.ICache.(GameActivityAssetsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameActivityAssetsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameActivityAssets))
	}
}

func Test_gameActivityAssetsCache_MultiSet(t *testing.T) {
	c := newGameActivityAssetsCache()
	defer c.Close()

	var testData []*model.GameActivityAssets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameActivityAssets))
	}

	err := c.ICache.(GameActivityAssetsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameActivityAssetsCache_Del(t *testing.T) {
	c := newGameActivityAssetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameActivityAssets)
	err := c.ICache.(GameActivityAssetsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameActivityAssetsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameActivityAssetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameActivityAssets)
	err := c.ICache.(GameActivityAssetsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameActivityAssetsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameActivityAssetsCache(t *testing.T) {
	c := NewGameActivityAssetsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameActivityAssetsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameActivityAssetsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
