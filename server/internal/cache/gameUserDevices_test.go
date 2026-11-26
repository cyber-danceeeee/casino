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

func newGameUserDevicesCache() *gotest.Cache {
	record1 := &model.GameUserDevices{}
	record1.ID = 1
	record2 := &model.GameUserDevices{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameUserDevicesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameUserDevicesCache_Set(t *testing.T) {
	c := newGameUserDevicesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserDevices)
	err := c.ICache.(GameUserDevicesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameUserDevicesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameUserDevicesCache_Get(t *testing.T) {
	c := newGameUserDevicesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserDevices)
	err := c.ICache.(GameUserDevicesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserDevicesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameUserDevicesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameUserDevicesCache_MultiGet(t *testing.T) {
	c := newGameUserDevicesCache()
	defer c.Close()

	var testData []*model.GameUserDevices
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserDevices))
	}

	err := c.ICache.(GameUserDevicesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserDevicesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameUserDevices))
	}
}

func Test_gameUserDevicesCache_MultiSet(t *testing.T) {
	c := newGameUserDevicesCache()
	defer c.Close()

	var testData []*model.GameUserDevices
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserDevices))
	}

	err := c.ICache.(GameUserDevicesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserDevicesCache_Del(t *testing.T) {
	c := newGameUserDevicesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserDevices)
	err := c.ICache.(GameUserDevicesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserDevicesCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameUserDevicesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserDevices)
	err := c.ICache.(GameUserDevicesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameUserDevicesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameUserDevicesCache(t *testing.T) {
	c := NewGameUserDevicesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameUserDevicesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameUserDevicesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
