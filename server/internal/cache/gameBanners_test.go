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

func newGameBannersCache() *gotest.Cache {
	record1 := &model.GameBanners{}
	record1.ID = 1
	record2 := &model.GameBanners{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameBannersCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameBannersCache_Set(t *testing.T) {
	c := newGameBannersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBanners)
	err := c.ICache.(GameBannersCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameBannersCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameBannersCache_Get(t *testing.T) {
	c := newGameBannersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBanners)
	err := c.ICache.(GameBannersCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameBannersCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameBannersCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameBannersCache_MultiGet(t *testing.T) {
	c := newGameBannersCache()
	defer c.Close()

	var testData []*model.GameBanners
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameBanners))
	}

	err := c.ICache.(GameBannersCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameBannersCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameBanners))
	}
}

func Test_gameBannersCache_MultiSet(t *testing.T) {
	c := newGameBannersCache()
	defer c.Close()

	var testData []*model.GameBanners
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameBanners))
	}

	err := c.ICache.(GameBannersCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameBannersCache_Del(t *testing.T) {
	c := newGameBannersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBanners)
	err := c.ICache.(GameBannersCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameBannersCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameBannersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameBanners)
	err := c.ICache.(GameBannersCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameBannersCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameBannersCache(t *testing.T) {
	c := NewGameBannersCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameBannersCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameBannersCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
