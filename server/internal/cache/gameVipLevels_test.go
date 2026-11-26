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

func newGameVipLevelsCache() *gotest.Cache {
	record1 := &model.GameVipLevels{}
	record1.ID = 1
	record2 := &model.GameVipLevels{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameVipLevelsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameVipLevelsCache_Set(t *testing.T) {
	c := newGameVipLevelsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameVipLevels)
	err := c.ICache.(GameVipLevelsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameVipLevelsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameVipLevelsCache_Get(t *testing.T) {
	c := newGameVipLevelsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameVipLevels)
	err := c.ICache.(GameVipLevelsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameVipLevelsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameVipLevelsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameVipLevelsCache_MultiGet(t *testing.T) {
	c := newGameVipLevelsCache()
	defer c.Close()

	var testData []*model.GameVipLevels
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameVipLevels))
	}

	err := c.ICache.(GameVipLevelsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameVipLevelsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameVipLevels))
	}
}

func Test_gameVipLevelsCache_MultiSet(t *testing.T) {
	c := newGameVipLevelsCache()
	defer c.Close()

	var testData []*model.GameVipLevels
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameVipLevels))
	}

	err := c.ICache.(GameVipLevelsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameVipLevelsCache_Del(t *testing.T) {
	c := newGameVipLevelsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameVipLevels)
	err := c.ICache.(GameVipLevelsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameVipLevelsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameVipLevelsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameVipLevels)
	err := c.ICache.(GameVipLevelsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameVipLevelsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameVipLevelsCache(t *testing.T) {
	c := NewGameVipLevelsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameVipLevelsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameVipLevelsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
