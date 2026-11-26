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

func newGameWalletTransactionsCache() *gotest.Cache {
	record1 := &model.GameWalletTransactions{}
	record1.ID = 1
	record2 := &model.GameWalletTransactions{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameWalletTransactionsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameWalletTransactionsCache_Set(t *testing.T) {
	c := newGameWalletTransactionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWalletTransactions)
	err := c.ICache.(GameWalletTransactionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameWalletTransactionsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameWalletTransactionsCache_Get(t *testing.T) {
	c := newGameWalletTransactionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWalletTransactions)
	err := c.ICache.(GameWalletTransactionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameWalletTransactionsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameWalletTransactionsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameWalletTransactionsCache_MultiGet(t *testing.T) {
	c := newGameWalletTransactionsCache()
	defer c.Close()

	var testData []*model.GameWalletTransactions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameWalletTransactions))
	}

	err := c.ICache.(GameWalletTransactionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameWalletTransactionsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameWalletTransactions))
	}
}

func Test_gameWalletTransactionsCache_MultiSet(t *testing.T) {
	c := newGameWalletTransactionsCache()
	defer c.Close()

	var testData []*model.GameWalletTransactions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameWalletTransactions))
	}

	err := c.ICache.(GameWalletTransactionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameWalletTransactionsCache_Del(t *testing.T) {
	c := newGameWalletTransactionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWalletTransactions)
	err := c.ICache.(GameWalletTransactionsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameWalletTransactionsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameWalletTransactionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameWalletTransactions)
	err := c.ICache.(GameWalletTransactionsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameWalletTransactionsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameWalletTransactionsCache(t *testing.T) {
	c := NewGameWalletTransactionsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameWalletTransactionsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameWalletTransactionsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
