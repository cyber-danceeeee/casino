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

func newGameWithdrawAccountsCache() *gotest.Cache {
	record1 := &model.GameWithdrawAccounts{}
	record1.ID = 1
	record2 := &model.GameWithdrawAccounts{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameWithdrawAccountsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameWithdrawAccountsCache_Set(t *testing.T) {
	c := newGameWithdrawAccountsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWithdrawAccounts)
	err := c.ICache.(GameWithdrawAccountsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameWithdrawAccountsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameWithdrawAccountsCache_Get(t *testing.T) {
	c := newGameWithdrawAccountsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWithdrawAccounts)
	err := c.ICache.(GameWithdrawAccountsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameWithdrawAccountsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameWithdrawAccountsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameWithdrawAccountsCache_MultiGet(t *testing.T) {
	c := newGameWithdrawAccountsCache()
	defer c.Close()

	var testData []*model.GameWithdrawAccounts
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameWithdrawAccounts))
	}

	err := c.ICache.(GameWithdrawAccountsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameWithdrawAccountsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameWithdrawAccounts))
	}
}

func Test_gameWithdrawAccountsCache_MultiSet(t *testing.T) {
	c := newGameWithdrawAccountsCache()
	defer c.Close()

	var testData []*model.GameWithdrawAccounts
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameWithdrawAccounts))
	}

	err := c.ICache.(GameWithdrawAccountsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameWithdrawAccountsCache_Del(t *testing.T) {
	c := newGameWithdrawAccountsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWithdrawAccounts)
	err := c.ICache.(GameWithdrawAccountsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameWithdrawAccountsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameWithdrawAccountsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWithdrawAccounts)
	err := c.ICache.(GameWithdrawAccountsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameWithdrawAccountsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameWithdrawAccountsCache(t *testing.T) {
	c := NewGameWithdrawAccountsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameWithdrawAccountsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameWithdrawAccountsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
