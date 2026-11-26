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

func newGameWelfareCardsCache() *gotest.Cache {
	record1 := &model.GameWelfareCards{}
	record1.ID = 1
	record2 := &model.GameWelfareCards{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameWelfareCardsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameWelfareCardsCache_Set(t *testing.T) {
	c := newGameWelfareCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWelfareCards)
	err := c.ICache.(GameWelfareCardsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameWelfareCardsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameWelfareCardsCache_Get(t *testing.T) {
	c := newGameWelfareCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWelfareCards)
	err := c.ICache.(GameWelfareCardsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameWelfareCardsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameWelfareCardsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameWelfareCardsCache_MultiGet(t *testing.T) {
	c := newGameWelfareCardsCache()
	defer c.Close()

	var testData []*model.GameWelfareCards
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameWelfareCards))
	}

	err := c.ICache.(GameWelfareCardsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameWelfareCardsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameWelfareCards))
	}
}

func Test_gameWelfareCardsCache_MultiSet(t *testing.T) {
	c := newGameWelfareCardsCache()
	defer c.Close()

	var testData []*model.GameWelfareCards
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameWelfareCards))
	}

	err := c.ICache.(GameWelfareCardsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameWelfareCardsCache_Del(t *testing.T) {
	c := newGameWelfareCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWelfareCards)
	err := c.ICache.(GameWelfareCardsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameWelfareCardsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameWelfareCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWelfareCards)
	err := c.ICache.(GameWelfareCardsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameWelfareCardsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameWelfareCardsCache(t *testing.T) {
	c := NewGameWelfareCardsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameWelfareCardsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameWelfareCardsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
