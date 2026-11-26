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

func newGameUserWalletsCache() *gotest.Cache {
	record1 := &model.GameUserWallets{}
	record1.ID = 1
	record2 := &model.GameUserWallets{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameUserWalletsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameUserWalletsCache_Set(t *testing.T) {
	c := newGameUserWalletsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserWallets)
	err := c.ICache.(GameUserWalletsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameUserWalletsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameUserWalletsCache_Get(t *testing.T) {
	c := newGameUserWalletsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserWallets)
	err := c.ICache.(GameUserWalletsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserWalletsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameUserWalletsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameUserWalletsCache_MultiGet(t *testing.T) {
	c := newGameUserWalletsCache()
	defer c.Close()

	var testData []*model.GameUserWallets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserWallets))
	}

	err := c.ICache.(GameUserWalletsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserWalletsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameUserWallets))
	}
}

func Test_gameUserWalletsCache_MultiSet(t *testing.T) {
	c := newGameUserWalletsCache()
	defer c.Close()

	var testData []*model.GameUserWallets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserWallets))
	}

	err := c.ICache.(GameUserWalletsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserWalletsCache_Del(t *testing.T) {
	c := newGameUserWalletsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserWallets)
	err := c.ICache.(GameUserWalletsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserWalletsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameUserWalletsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserWallets)
	err := c.ICache.(GameUserWalletsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameUserWalletsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameUserWalletsCache(t *testing.T) {
	c := NewGameUserWalletsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameUserWalletsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameUserWalletsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
