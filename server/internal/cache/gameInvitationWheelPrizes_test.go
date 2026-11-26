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

func newGameInvitationWheelPrizesCache() *gotest.Cache {
	record1 := &model.GameInvitationWheelPrizes{}
	record1.ID = 1
	record2 := &model.GameInvitationWheelPrizes{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameInvitationWheelPrizesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameInvitationWheelPrizesCache_Set(t *testing.T) {
	c := newGameInvitationWheelPrizesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelPrizes)
	err := c.ICache.(GameInvitationWheelPrizesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameInvitationWheelPrizesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameInvitationWheelPrizesCache_Get(t *testing.T) {
	c := newGameInvitationWheelPrizesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelPrizes)
	err := c.ICache.(GameInvitationWheelPrizesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameInvitationWheelPrizesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameInvitationWheelPrizesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameInvitationWheelPrizesCache_MultiGet(t *testing.T) {
	c := newGameInvitationWheelPrizesCache()
	defer c.Close()

	var testData []*model.GameInvitationWheelPrizes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameInvitationWheelPrizes))
	}

	err := c.ICache.(GameInvitationWheelPrizesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameInvitationWheelPrizesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameInvitationWheelPrizes))
	}
}

func Test_gameInvitationWheelPrizesCache_MultiSet(t *testing.T) {
	c := newGameInvitationWheelPrizesCache()
	defer c.Close()

	var testData []*model.GameInvitationWheelPrizes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameInvitationWheelPrizes))
	}

	err := c.ICache.(GameInvitationWheelPrizesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameInvitationWheelPrizesCache_Del(t *testing.T) {
	c := newGameInvitationWheelPrizesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelPrizes)
	err := c.ICache.(GameInvitationWheelPrizesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameInvitationWheelPrizesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameInvitationWheelPrizesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelPrizes)
	err := c.ICache.(GameInvitationWheelPrizesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameInvitationWheelPrizesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameInvitationWheelPrizesCache(t *testing.T) {
	c := NewGameInvitationWheelPrizesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameInvitationWheelPrizesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameInvitationWheelPrizesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
