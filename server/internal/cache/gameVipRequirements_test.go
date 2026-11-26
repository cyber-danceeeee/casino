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

func newGameVipRequirementsCache() *gotest.Cache {
	record1 := &model.GameVipRequirements{}
	record1.ID = 1
	record2 := &model.GameVipRequirements{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameVipRequirementsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameVipRequirementsCache_Set(t *testing.T) {
	c := newGameVipRequirementsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameVipRequirements)
	err := c.ICache.(GameVipRequirementsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameVipRequirementsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameVipRequirementsCache_Get(t *testing.T) {
	c := newGameVipRequirementsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameVipRequirements)
	err := c.ICache.(GameVipRequirementsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameVipRequirementsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameVipRequirementsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameVipRequirementsCache_MultiGet(t *testing.T) {
	c := newGameVipRequirementsCache()
	defer c.Close()

	var testData []*model.GameVipRequirements
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameVipRequirements))
	}

	err := c.ICache.(GameVipRequirementsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameVipRequirementsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameVipRequirements))
	}
}

func Test_gameVipRequirementsCache_MultiSet(t *testing.T) {
	c := newGameVipRequirementsCache()
	defer c.Close()

	var testData []*model.GameVipRequirements
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameVipRequirements))
	}

	err := c.ICache.(GameVipRequirementsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameVipRequirementsCache_Del(t *testing.T) {
	c := newGameVipRequirementsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameVipRequirements)
	err := c.ICache.(GameVipRequirementsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameVipRequirementsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameVipRequirementsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameVipRequirements)
	err := c.ICache.(GameVipRequirementsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameVipRequirementsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameVipRequirementsCache(t *testing.T) {
	c := NewGameVipRequirementsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameVipRequirementsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameVipRequirementsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
