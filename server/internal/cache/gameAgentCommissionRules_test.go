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

func newGameAgentCommissionRulesCache() *gotest.Cache {
	record1 := &model.GameAgentCommissionRules{}
	record1.ID = 1
	record2 := &model.GameAgentCommissionRules{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameAgentCommissionRulesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameAgentCommissionRulesCache_Set(t *testing.T) {
	c := newGameAgentCommissionRulesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAgentCommissionRules)
	err := c.ICache.(GameAgentCommissionRulesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameAgentCommissionRulesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameAgentCommissionRulesCache_Get(t *testing.T) {
	c := newGameAgentCommissionRulesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAgentCommissionRules)
	err := c.ICache.(GameAgentCommissionRulesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameAgentCommissionRulesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameAgentCommissionRulesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameAgentCommissionRulesCache_MultiGet(t *testing.T) {
	c := newGameAgentCommissionRulesCache()
	defer c.Close()

	var testData []*model.GameAgentCommissionRules
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameAgentCommissionRules))
	}

	err := c.ICache.(GameAgentCommissionRulesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameAgentCommissionRulesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameAgentCommissionRules))
	}
}

func Test_gameAgentCommissionRulesCache_MultiSet(t *testing.T) {
	c := newGameAgentCommissionRulesCache()
	defer c.Close()

	var testData []*model.GameAgentCommissionRules
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameAgentCommissionRules))
	}

	err := c.ICache.(GameAgentCommissionRulesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameAgentCommissionRulesCache_Del(t *testing.T) {
	c := newGameAgentCommissionRulesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAgentCommissionRules)
	err := c.ICache.(GameAgentCommissionRulesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameAgentCommissionRulesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameAgentCommissionRulesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameAgentCommissionRules)
	err := c.ICache.(GameAgentCommissionRulesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameAgentCommissionRulesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameAgentCommissionRulesCache(t *testing.T) {
	c := NewGameAgentCommissionRulesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameAgentCommissionRulesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameAgentCommissionRulesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
