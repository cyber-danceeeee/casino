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

func newGameGamesCache() *gotest.Cache {
	record1 := &model.GameGames{}
	record1.ID = 1
	record2 := &model.GameGames{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameGamesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameGamesCache_Set(t *testing.T) {
	c := newGameGamesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGames)
	err := c.ICache.(GameGamesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameGamesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameGamesCache_Get(t *testing.T) {
	c := newGameGamesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGames)
	err := c.ICache.(GameGamesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameGamesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameGamesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameGamesCache_MultiGet(t *testing.T) {
	c := newGameGamesCache()
	defer c.Close()

	var testData []*model.GameGames
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameGames))
	}

	err := c.ICache.(GameGamesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameGamesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameGames))
	}
}

func Test_gameGamesCache_MultiSet(t *testing.T) {
	c := newGameGamesCache()
	defer c.Close()

	var testData []*model.GameGames
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameGames))
	}

	err := c.ICache.(GameGamesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGamesCache_Del(t *testing.T) {
	c := newGameGamesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGames)
	err := c.ICache.(GameGamesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGamesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameGamesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGames)
	err := c.ICache.(GameGamesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameGamesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameGamesCache(t *testing.T) {
	c := NewGameGamesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameGamesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameGamesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
