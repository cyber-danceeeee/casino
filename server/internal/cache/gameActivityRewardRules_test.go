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

func newGameActivityRewardRulesCache() *gotest.Cache {
	record1 := &model.GameActivityRewardRules{}
	record1.ID = 1
	record2 := &model.GameActivityRewardRules{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameActivityRewardRulesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameActivityRewardRulesCache_Set(t *testing.T) {
	c := newGameActivityRewardRulesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameActivityRewardRules)
	err := c.ICache.(GameActivityRewardRulesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameActivityRewardRulesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameActivityRewardRulesCache_Get(t *testing.T) {
	c := newGameActivityRewardRulesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameActivityRewardRules)
	err := c.ICache.(GameActivityRewardRulesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameActivityRewardRulesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameActivityRewardRulesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameActivityRewardRulesCache_MultiGet(t *testing.T) {
	c := newGameActivityRewardRulesCache()
	defer c.Close()

	var testData []*model.GameActivityRewardRules
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameActivityRewardRules))
	}

	err := c.ICache.(GameActivityRewardRulesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameActivityRewardRulesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameActivityRewardRules))
	}
}

func Test_gameActivityRewardRulesCache_MultiSet(t *testing.T) {
	c := newGameActivityRewardRulesCache()
	defer c.Close()

	var testData []*model.GameActivityRewardRules
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameActivityRewardRules))
	}

	err := c.ICache.(GameActivityRewardRulesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameActivityRewardRulesCache_Del(t *testing.T) {
	c := newGameActivityRewardRulesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameActivityRewardRules)
	err := c.ICache.(GameActivityRewardRulesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameActivityRewardRulesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameActivityRewardRulesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameActivityRewardRules)
	err := c.ICache.(GameActivityRewardRulesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameActivityRewardRulesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameActivityRewardRulesCache(t *testing.T) {
	c := NewGameActivityRewardRulesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameActivityRewardRulesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameActivityRewardRulesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
