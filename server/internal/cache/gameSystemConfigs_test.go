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

func newGameSystemConfigsCache() *gotest.Cache {
	record1 := &model.GameSystemConfigs{}
	record1.ID = 1
	record2 := &model.GameSystemConfigs{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameSystemConfigsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameSystemConfigsCache_Set(t *testing.T) {
	c := newGameSystemConfigsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSystemConfigs)
	err := c.ICache.(GameSystemConfigsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameSystemConfigsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameSystemConfigsCache_Get(t *testing.T) {
	c := newGameSystemConfigsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSystemConfigs)
	err := c.ICache.(GameSystemConfigsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameSystemConfigsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameSystemConfigsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameSystemConfigsCache_MultiGet(t *testing.T) {
	c := newGameSystemConfigsCache()
	defer c.Close()

	var testData []*model.GameSystemConfigs
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameSystemConfigs))
	}

	err := c.ICache.(GameSystemConfigsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameSystemConfigsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameSystemConfigs))
	}
}

func Test_gameSystemConfigsCache_MultiSet(t *testing.T) {
	c := newGameSystemConfigsCache()
	defer c.Close()

	var testData []*model.GameSystemConfigs
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameSystemConfigs))
	}

	err := c.ICache.(GameSystemConfigsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameSystemConfigsCache_Del(t *testing.T) {
	c := newGameSystemConfigsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSystemConfigs)
	err := c.ICache.(GameSystemConfigsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameSystemConfigsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameSystemConfigsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSystemConfigs)
	err := c.ICache.(GameSystemConfigsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameSystemConfigsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameSystemConfigsCache(t *testing.T) {
	c := NewGameSystemConfigsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameSystemConfigsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameSystemConfigsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
