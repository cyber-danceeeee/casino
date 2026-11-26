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

func newGameUserFavoriteGamesCache() *gotest.Cache {
	record1 := &model.GameUserFavoriteGames{}
	record1.ID = 1
	record2 := &model.GameUserFavoriteGames{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameUserFavoriteGamesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameUserFavoriteGamesCache_Set(t *testing.T) {
	c := newGameUserFavoriteGamesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserFavoriteGames)
	err := c.ICache.(GameUserFavoriteGamesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameUserFavoriteGamesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameUserFavoriteGamesCache_Get(t *testing.T) {
	c := newGameUserFavoriteGamesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserFavoriteGames)
	err := c.ICache.(GameUserFavoriteGamesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserFavoriteGamesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameUserFavoriteGamesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameUserFavoriteGamesCache_MultiGet(t *testing.T) {
	c := newGameUserFavoriteGamesCache()
	defer c.Close()

	var testData []*model.GameUserFavoriteGames
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserFavoriteGames))
	}

	err := c.ICache.(GameUserFavoriteGamesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserFavoriteGamesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameUserFavoriteGames))
	}
}

func Test_gameUserFavoriteGamesCache_MultiSet(t *testing.T) {
	c := newGameUserFavoriteGamesCache()
	defer c.Close()

	var testData []*model.GameUserFavoriteGames
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserFavoriteGames))
	}

	err := c.ICache.(GameUserFavoriteGamesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserFavoriteGamesCache_Del(t *testing.T) {
	c := newGameUserFavoriteGamesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserFavoriteGames)
	err := c.ICache.(GameUserFavoriteGamesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserFavoriteGamesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameUserFavoriteGamesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserFavoriteGames)
	err := c.ICache.(GameUserFavoriteGamesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameUserFavoriteGamesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameUserFavoriteGamesCache(t *testing.T) {
	c := NewGameUserFavoriteGamesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameUserFavoriteGamesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameUserFavoriteGamesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
