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

func newGameNoticesCache() *gotest.Cache {
	record1 := &model.GameNotices{}
	record1.ID = 1
	record2 := &model.GameNotices{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameNoticesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameNoticesCache_Set(t *testing.T) {
	c := newGameNoticesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameNotices)
	err := c.ICache.(GameNoticesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameNoticesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameNoticesCache_Get(t *testing.T) {
	c := newGameNoticesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameNotices)
	err := c.ICache.(GameNoticesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameNoticesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameNoticesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameNoticesCache_MultiGet(t *testing.T) {
	c := newGameNoticesCache()
	defer c.Close()

	var testData []*model.GameNotices
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameNotices))
	}

	err := c.ICache.(GameNoticesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameNoticesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameNotices))
	}
}

func Test_gameNoticesCache_MultiSet(t *testing.T) {
	c := newGameNoticesCache()
	defer c.Close()

	var testData []*model.GameNotices
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameNotices))
	}

	err := c.ICache.(GameNoticesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameNoticesCache_Del(t *testing.T) {
	c := newGameNoticesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameNotices)
	err := c.ICache.(GameNoticesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameNoticesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameNoticesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameNotices)
	err := c.ICache.(GameNoticesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameNoticesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameNoticesCache(t *testing.T) {
	c := NewGameNoticesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameNoticesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameNoticesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
