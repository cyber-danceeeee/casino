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

func newGameInvitationChestConfigsCache() *gotest.Cache {
	record1 := &model.GameInvitationChestConfigs{}
	record1.ID = 1
	record2 := &model.GameInvitationChestConfigs{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameInvitationChestConfigsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameInvitationChestConfigsCache_Set(t *testing.T) {
	c := newGameInvitationChestConfigsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationChestConfigs)
	err := c.ICache.(GameInvitationChestConfigsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameInvitationChestConfigsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameInvitationChestConfigsCache_Get(t *testing.T) {
	c := newGameInvitationChestConfigsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationChestConfigs)
	err := c.ICache.(GameInvitationChestConfigsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameInvitationChestConfigsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameInvitationChestConfigsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameInvitationChestConfigsCache_MultiGet(t *testing.T) {
	c := newGameInvitationChestConfigsCache()
	defer c.Close()

	var testData []*model.GameInvitationChestConfigs
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameInvitationChestConfigs))
	}

	err := c.ICache.(GameInvitationChestConfigsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameInvitationChestConfigsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameInvitationChestConfigs))
	}
}

func Test_gameInvitationChestConfigsCache_MultiSet(t *testing.T) {
	c := newGameInvitationChestConfigsCache()
	defer c.Close()

	var testData []*model.GameInvitationChestConfigs
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameInvitationChestConfigs))
	}

	err := c.ICache.(GameInvitationChestConfigsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameInvitationChestConfigsCache_Del(t *testing.T) {
	c := newGameInvitationChestConfigsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationChestConfigs)
	err := c.ICache.(GameInvitationChestConfigsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameInvitationChestConfigsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameInvitationChestConfigsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationChestConfigs)
	err := c.ICache.(GameInvitationChestConfigsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameInvitationChestConfigsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameInvitationChestConfigsCache(t *testing.T) {
	c := NewGameInvitationChestConfigsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameInvitationChestConfigsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameInvitationChestConfigsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
