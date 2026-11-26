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

func newGameSmsVerificationCodesCache() *gotest.Cache {
	record1 := &model.GameSmsVerificationCodes{}
	record1.ID = 1
	record2 := &model.GameSmsVerificationCodes{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameSmsVerificationCodesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameSmsVerificationCodesCache_Set(t *testing.T) {
	c := newGameSmsVerificationCodesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSmsVerificationCodes)
	err := c.ICache.(GameSmsVerificationCodesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameSmsVerificationCodesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameSmsVerificationCodesCache_Get(t *testing.T) {
	c := newGameSmsVerificationCodesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSmsVerificationCodes)
	err := c.ICache.(GameSmsVerificationCodesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameSmsVerificationCodesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameSmsVerificationCodesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameSmsVerificationCodesCache_MultiGet(t *testing.T) {
	c := newGameSmsVerificationCodesCache()
	defer c.Close()

	var testData []*model.GameSmsVerificationCodes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameSmsVerificationCodes))
	}

	err := c.ICache.(GameSmsVerificationCodesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameSmsVerificationCodesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameSmsVerificationCodes))
	}
}

func Test_gameSmsVerificationCodesCache_MultiSet(t *testing.T) {
	c := newGameSmsVerificationCodesCache()
	defer c.Close()

	var testData []*model.GameSmsVerificationCodes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameSmsVerificationCodes))
	}

	err := c.ICache.(GameSmsVerificationCodesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameSmsVerificationCodesCache_Del(t *testing.T) {
	c := newGameSmsVerificationCodesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSmsVerificationCodes)
	err := c.ICache.(GameSmsVerificationCodesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameSmsVerificationCodesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameSmsVerificationCodesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSmsVerificationCodes)
	err := c.ICache.(GameSmsVerificationCodesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameSmsVerificationCodesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameSmsVerificationCodesCache(t *testing.T) {
	c := NewGameSmsVerificationCodesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameSmsVerificationCodesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameSmsVerificationCodesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
