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

func newGameGameLaunchLogsCache() *gotest.Cache {
	record1 := &model.GameGameLaunchLogs{}
	record1.ID = 1
	record2 := &model.GameGameLaunchLogs{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameGameLaunchLogsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameGameLaunchLogsCache_Set(t *testing.T) {
	c := newGameGameLaunchLogsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameLaunchLogs)
	err := c.ICache.(GameGameLaunchLogsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameGameLaunchLogsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameGameLaunchLogsCache_Get(t *testing.T) {
	c := newGameGameLaunchLogsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameLaunchLogs)
	err := c.ICache.(GameGameLaunchLogsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameGameLaunchLogsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameGameLaunchLogsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameGameLaunchLogsCache_MultiGet(t *testing.T) {
	c := newGameGameLaunchLogsCache()
	defer c.Close()

	var testData []*model.GameGameLaunchLogs
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameGameLaunchLogs))
	}

	err := c.ICache.(GameGameLaunchLogsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameGameLaunchLogsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameGameLaunchLogs))
	}
}

func Test_gameGameLaunchLogsCache_MultiSet(t *testing.T) {
	c := newGameGameLaunchLogsCache()
	defer c.Close()

	var testData []*model.GameGameLaunchLogs
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameGameLaunchLogs))
	}

	err := c.ICache.(GameGameLaunchLogsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGameLaunchLogsCache_Del(t *testing.T) {
	c := newGameGameLaunchLogsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameLaunchLogs)
	err := c.ICache.(GameGameLaunchLogsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameGameLaunchLogsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameGameLaunchLogsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameGameLaunchLogs)
	err := c.ICache.(GameGameLaunchLogsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameGameLaunchLogsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameGameLaunchLogsCache(t *testing.T) {
	c := NewGameGameLaunchLogsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameGameLaunchLogsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameGameLaunchLogsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
