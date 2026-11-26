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

func newGameSupportTicketsCache() *gotest.Cache {
	record1 := &model.GameSupportTickets{}
	record1.ID = 1
	record2 := &model.GameSupportTickets{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameSupportTicketsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameSupportTicketsCache_Set(t *testing.T) {
	c := newGameSupportTicketsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportTickets)
	err := c.ICache.(GameSupportTicketsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameSupportTicketsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameSupportTicketsCache_Get(t *testing.T) {
	c := newGameSupportTicketsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportTickets)
	err := c.ICache.(GameSupportTicketsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameSupportTicketsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameSupportTicketsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameSupportTicketsCache_MultiGet(t *testing.T) {
	c := newGameSupportTicketsCache()
	defer c.Close()

	var testData []*model.GameSupportTickets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameSupportTickets))
	}

	err := c.ICache.(GameSupportTicketsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameSupportTicketsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameSupportTickets))
	}
}

func Test_gameSupportTicketsCache_MultiSet(t *testing.T) {
	c := newGameSupportTicketsCache()
	defer c.Close()

	var testData []*model.GameSupportTickets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameSupportTickets))
	}

	err := c.ICache.(GameSupportTicketsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameSupportTicketsCache_Del(t *testing.T) {
	c := newGameSupportTicketsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportTickets)
	err := c.ICache.(GameSupportTicketsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameSupportTicketsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameSupportTicketsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportTickets)
	err := c.ICache.(GameSupportTicketsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameSupportTicketsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameSupportTicketsCache(t *testing.T) {
	c := NewGameSupportTicketsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameSupportTicketsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameSupportTicketsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
