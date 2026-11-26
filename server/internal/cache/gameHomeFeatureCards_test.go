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

func newGameHomeFeatureCardsCache() *gotest.Cache {
	record1 := &model.GameHomeFeatureCards{}
	record1.ID = 1
	record2 := &model.GameHomeFeatureCards{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameHomeFeatureCardsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameHomeFeatureCardsCache_Set(t *testing.T) {
	c := newGameHomeFeatureCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameHomeFeatureCards)
	err := c.ICache.(GameHomeFeatureCardsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameHomeFeatureCardsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameHomeFeatureCardsCache_Get(t *testing.T) {
	c := newGameHomeFeatureCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameHomeFeatureCards)
	err := c.ICache.(GameHomeFeatureCardsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameHomeFeatureCardsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameHomeFeatureCardsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameHomeFeatureCardsCache_MultiGet(t *testing.T) {
	c := newGameHomeFeatureCardsCache()
	defer c.Close()

	var testData []*model.GameHomeFeatureCards
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameHomeFeatureCards))
	}

	err := c.ICache.(GameHomeFeatureCardsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameHomeFeatureCardsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameHomeFeatureCards))
	}
}

func Test_gameHomeFeatureCardsCache_MultiSet(t *testing.T) {
	c := newGameHomeFeatureCardsCache()
	defer c.Close()

	var testData []*model.GameHomeFeatureCards
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameHomeFeatureCards))
	}

	err := c.ICache.(GameHomeFeatureCardsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameHomeFeatureCardsCache_Del(t *testing.T) {
	c := newGameHomeFeatureCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameHomeFeatureCards)
	err := c.ICache.(GameHomeFeatureCardsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameHomeFeatureCardsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameHomeFeatureCardsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameHomeFeatureCards)
	err := c.ICache.(GameHomeFeatureCardsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameHomeFeatureCardsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameHomeFeatureCardsCache(t *testing.T) {
	c := NewGameHomeFeatureCardsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameHomeFeatureCardsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameHomeFeatureCardsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
