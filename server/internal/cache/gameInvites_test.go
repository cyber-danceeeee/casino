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

func newGameInvitesCache() *gotest.Cache {
	record1 := &model.GameInvites{}
	record1.ID = 1
	record2 := &model.GameInvites{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameInvitesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameInvitesCache_Set(t *testing.T) {
	c := newGameInvitesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvites)
	err := c.ICache.(GameInvitesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameInvitesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameInvitesCache_Get(t *testing.T) {
	c := newGameInvitesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvites)
	err := c.ICache.(GameInvitesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameInvitesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameInvitesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameInvitesCache_MultiGet(t *testing.T) {
	c := newGameInvitesCache()
	defer c.Close()

	var testData []*model.GameInvites
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameInvites))
	}

	err := c.ICache.(GameInvitesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameInvitesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameInvites))
	}
}

func Test_gameInvitesCache_MultiSet(t *testing.T) {
	c := newGameInvitesCache()
	defer c.Close()

	var testData []*model.GameInvites
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameInvites))
	}

	err := c.ICache.(GameInvitesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameInvitesCache_Del(t *testing.T) {
	c := newGameInvitesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvites)
	err := c.ICache.(GameInvitesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameInvitesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameInvitesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameInvites)
	err := c.ICache.(GameInvitesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameInvitesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameInvitesCache(t *testing.T) {
	c := NewGameInvitesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameInvitesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameInvitesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
