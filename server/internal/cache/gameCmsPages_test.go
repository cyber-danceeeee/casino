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

func newGameCmsPagesCache() *gotest.Cache {
	record1 := &model.GameCmsPages{}
	record1.ID = 1
	record2 := &model.GameCmsPages{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameCmsPagesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameCmsPagesCache_Set(t *testing.T) {
	c := newGameCmsPagesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameCmsPages)
	err := c.ICache.(GameCmsPagesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameCmsPagesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameCmsPagesCache_Get(t *testing.T) {
	c := newGameCmsPagesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameCmsPages)
	err := c.ICache.(GameCmsPagesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameCmsPagesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameCmsPagesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameCmsPagesCache_MultiGet(t *testing.T) {
	c := newGameCmsPagesCache()
	defer c.Close()

	var testData []*model.GameCmsPages
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameCmsPages))
	}

	err := c.ICache.(GameCmsPagesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameCmsPagesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameCmsPages))
	}
}

func Test_gameCmsPagesCache_MultiSet(t *testing.T) {
	c := newGameCmsPagesCache()
	defer c.Close()

	var testData []*model.GameCmsPages
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameCmsPages))
	}

	err := c.ICache.(GameCmsPagesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameCmsPagesCache_Del(t *testing.T) {
	c := newGameCmsPagesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameCmsPages)
	err := c.ICache.(GameCmsPagesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameCmsPagesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameCmsPagesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameCmsPages)
	err := c.ICache.(GameCmsPagesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameCmsPagesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameCmsPagesCache(t *testing.T) {
	c := NewGameCmsPagesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameCmsPagesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameCmsPagesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
