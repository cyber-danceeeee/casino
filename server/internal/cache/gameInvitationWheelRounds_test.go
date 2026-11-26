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

func newGameInvitationWheelRoundsCache() *gotest.Cache {
	record1 := &model.GameInvitationWheelRounds{}
	record1.ID = 1
	record2 := &model.GameInvitationWheelRounds{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameInvitationWheelRoundsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameInvitationWheelRoundsCache_Set(t *testing.T) {
	c := newGameInvitationWheelRoundsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelRounds)
	err := c.ICache.(GameInvitationWheelRoundsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameInvitationWheelRoundsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameInvitationWheelRoundsCache_Get(t *testing.T) {
	c := newGameInvitationWheelRoundsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelRounds)
	err := c.ICache.(GameInvitationWheelRoundsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameInvitationWheelRoundsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameInvitationWheelRoundsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameInvitationWheelRoundsCache_MultiGet(t *testing.T) {
	c := newGameInvitationWheelRoundsCache()
	defer c.Close()

	var testData []*model.GameInvitationWheelRounds
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameInvitationWheelRounds))
	}

	err := c.ICache.(GameInvitationWheelRoundsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameInvitationWheelRoundsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameInvitationWheelRounds))
	}
}

func Test_gameInvitationWheelRoundsCache_MultiSet(t *testing.T) {
	c := newGameInvitationWheelRoundsCache()
	defer c.Close()

	var testData []*model.GameInvitationWheelRounds
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameInvitationWheelRounds))
	}

	err := c.ICache.(GameInvitationWheelRoundsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameInvitationWheelRoundsCache_Del(t *testing.T) {
	c := newGameInvitationWheelRoundsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelRounds)
	err := c.ICache.(GameInvitationWheelRoundsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameInvitationWheelRoundsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameInvitationWheelRoundsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelRounds)
	err := c.ICache.(GameInvitationWheelRoundsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameInvitationWheelRoundsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameInvitationWheelRoundsCache(t *testing.T) {
	c := NewGameInvitationWheelRoundsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameInvitationWheelRoundsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameInvitationWheelRoundsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
