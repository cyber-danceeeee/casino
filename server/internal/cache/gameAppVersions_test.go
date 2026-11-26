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

func newGameAppVersionsCache() *gotest.Cache {
	record1 := &model.GameAppVersions{}
	record1.ID = 1
	record2 := &model.GameAppVersions{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameAppVersionsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameAppVersionsCache_Set(t *testing.T) {
	c := newGameAppVersionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAppVersions)
	err := c.ICache.(GameAppVersionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameAppVersionsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameAppVersionsCache_Get(t *testing.T) {
	c := newGameAppVersionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAppVersions)
	err := c.ICache.(GameAppVersionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameAppVersionsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameAppVersionsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameAppVersionsCache_MultiGet(t *testing.T) {
	c := newGameAppVersionsCache()
	defer c.Close()

	var testData []*model.GameAppVersions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameAppVersions))
	}

	err := c.ICache.(GameAppVersionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameAppVersionsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameAppVersions))
	}
}

func Test_gameAppVersionsCache_MultiSet(t *testing.T) {
	c := newGameAppVersionsCache()
	defer c.Close()

	var testData []*model.GameAppVersions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameAppVersions))
	}

	err := c.ICache.(GameAppVersionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameAppVersionsCache_Del(t *testing.T) {
	c := newGameAppVersionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAppVersions)
	err := c.ICache.(GameAppVersionsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameAppVersionsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameAppVersionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAppVersions)
	err := c.ICache.(GameAppVersionsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameAppVersionsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameAppVersionsCache(t *testing.T) {
	c := NewGameAppVersionsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameAppVersionsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameAppVersionsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
