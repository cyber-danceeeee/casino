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

func newGameUserWelfareCardsCache() *gotest.Cache {
	record1 := &model.GameUserWelfareCards{}
	record1.ID = 1
	record2 := &model.GameUserWelfareCards{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameUserWelfareCardsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameUserWelfareCardsCache_Set(t *testing.T) {
	c := newGameUserWelfareCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserWelfareCards)
	err := c.ICache.(GameUserWelfareCardsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameUserWelfareCardsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameUserWelfareCardsCache_Get(t *testing.T) {
	c := newGameUserWelfareCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserWelfareCards)
	err := c.ICache.(GameUserWelfareCardsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserWelfareCardsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameUserWelfareCardsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameUserWelfareCardsCache_MultiGet(t *testing.T) {
	c := newGameUserWelfareCardsCache()
	defer c.Close()

	var testData []*model.GameUserWelfareCards
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserWelfareCards))
	}

	err := c.ICache.(GameUserWelfareCardsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserWelfareCardsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameUserWelfareCards))
	}
}

func Test_gameUserWelfareCardsCache_MultiSet(t *testing.T) {
	c := newGameUserWelfareCardsCache()
	defer c.Close()

	var testData []*model.GameUserWelfareCards
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserWelfareCards))
	}

	err := c.ICache.(GameUserWelfareCardsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserWelfareCardsCache_Del(t *testing.T) {
	c := newGameUserWelfareCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserWelfareCards)
	err := c.ICache.(GameUserWelfareCardsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserWelfareCardsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameUserWelfareCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserWelfareCards)
	err := c.ICache.(GameUserWelfareCardsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameUserWelfareCardsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameUserWelfareCardsCache(t *testing.T) {
	c := NewGameUserWelfareCardsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameUserWelfareCardsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameUserWelfareCardsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
