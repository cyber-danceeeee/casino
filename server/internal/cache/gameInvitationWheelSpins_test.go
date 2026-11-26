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

func newGameInvitationWheelSpinsCache() *gotest.Cache {
	record1 := &model.GameInvitationWheelSpins{}
	record1.ID = 1
	record2 := &model.GameInvitationWheelSpins{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameInvitationWheelSpinsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameInvitationWheelSpinsCache_Set(t *testing.T) {
	c := newGameInvitationWheelSpinsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelSpins)
	err := c.ICache.(GameInvitationWheelSpinsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameInvitationWheelSpinsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameInvitationWheelSpinsCache_Get(t *testing.T) {
	c := newGameInvitationWheelSpinsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelSpins)
	err := c.ICache.(GameInvitationWheelSpinsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameInvitationWheelSpinsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameInvitationWheelSpinsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameInvitationWheelSpinsCache_MultiGet(t *testing.T) {
	c := newGameInvitationWheelSpinsCache()
	defer c.Close()

	var testData []*model.GameInvitationWheelSpins
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameInvitationWheelSpins))
	}

	err := c.ICache.(GameInvitationWheelSpinsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameInvitationWheelSpinsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameInvitationWheelSpins))
	}
}

func Test_gameInvitationWheelSpinsCache_MultiSet(t *testing.T) {
	c := newGameInvitationWheelSpinsCache()
	defer c.Close()

	var testData []*model.GameInvitationWheelSpins
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameInvitationWheelSpins))
	}

	err := c.ICache.(GameInvitationWheelSpinsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameInvitationWheelSpinsCache_Del(t *testing.T) {
	c := newGameInvitationWheelSpinsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelSpins)
	err := c.ICache.(GameInvitationWheelSpinsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameInvitationWheelSpinsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameInvitationWheelSpinsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvitationWheelSpins)
	err := c.ICache.(GameInvitationWheelSpinsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameInvitationWheelSpinsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameInvitationWheelSpinsCache(t *testing.T) {
	c := NewGameInvitationWheelSpinsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameInvitationWheelSpinsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameInvitationWheelSpinsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
