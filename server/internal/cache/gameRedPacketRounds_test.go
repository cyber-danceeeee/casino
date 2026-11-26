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

func newGameRedPacketRoundsCache() *gotest.Cache {
	record1 := &model.GameRedPacketRounds{}
	record1.ID = 1
	record2 := &model.GameRedPacketRounds{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameRedPacketRoundsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameRedPacketRoundsCache_Set(t *testing.T) {
	c := newGameRedPacketRoundsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRedPacketRounds)
	err := c.ICache.(GameRedPacketRoundsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameRedPacketRoundsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameRedPacketRoundsCache_Get(t *testing.T) {
	c := newGameRedPacketRoundsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRedPacketRounds)
	err := c.ICache.(GameRedPacketRoundsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameRedPacketRoundsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameRedPacketRoundsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameRedPacketRoundsCache_MultiGet(t *testing.T) {
	c := newGameRedPacketRoundsCache()
	defer c.Close()

	var testData []*model.GameRedPacketRounds
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameRedPacketRounds))
	}

	err := c.ICache.(GameRedPacketRoundsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameRedPacketRoundsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameRedPacketRounds))
	}
}

func Test_gameRedPacketRoundsCache_MultiSet(t *testing.T) {
	c := newGameRedPacketRoundsCache()
	defer c.Close()

	var testData []*model.GameRedPacketRounds
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameRedPacketRounds))
	}

	err := c.ICache.(GameRedPacketRoundsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameRedPacketRoundsCache_Del(t *testing.T) {
	c := newGameRedPacketRoundsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRedPacketRounds)
	err := c.ICache.(GameRedPacketRoundsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameRedPacketRoundsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameRedPacketRoundsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameRedPacketRounds)
	err := c.ICache.(GameRedPacketRoundsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameRedPacketRoundsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameRedPacketRoundsCache(t *testing.T) {
	c := NewGameRedPacketRoundsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameRedPacketRoundsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameRedPacketRoundsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
