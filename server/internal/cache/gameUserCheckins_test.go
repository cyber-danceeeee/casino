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

func newGameUserCheckinsCache() *gotest.Cache {
	record1 := &model.GameUserCheckins{}
	record1.ID = 1
	record2 := &model.GameUserCheckins{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameUserCheckinsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameUserCheckinsCache_Set(t *testing.T) {
	c := newGameUserCheckinsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserCheckins)
	err := c.ICache.(GameUserCheckinsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameUserCheckinsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameUserCheckinsCache_Get(t *testing.T) {
	c := newGameUserCheckinsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserCheckins)
	err := c.ICache.(GameUserCheckinsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserCheckinsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameUserCheckinsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameUserCheckinsCache_MultiGet(t *testing.T) {
	c := newGameUserCheckinsCache()
	defer c.Close()

	var testData []*model.GameUserCheckins
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserCheckins))
	}

	err := c.ICache.(GameUserCheckinsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserCheckinsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameUserCheckins))
	}
}

func Test_gameUserCheckinsCache_MultiSet(t *testing.T) {
	c := newGameUserCheckinsCache()
	defer c.Close()

	var testData []*model.GameUserCheckins
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserCheckins))
	}

	err := c.ICache.(GameUserCheckinsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserCheckinsCache_Del(t *testing.T) {
	c := newGameUserCheckinsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserCheckins)
	err := c.ICache.(GameUserCheckinsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserCheckinsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameUserCheckinsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserCheckins)
	err := c.ICache.(GameUserCheckinsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameUserCheckinsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameUserCheckinsCache(t *testing.T) {
	c := NewGameUserCheckinsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameUserCheckinsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameUserCheckinsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
