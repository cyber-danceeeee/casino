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

func newGameUserSettingsCache() *gotest.Cache {
	record1 := &model.GameUserSettings{}
	record1.ID = 1
	record2 := &model.GameUserSettings{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameUserSettingsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameUserSettingsCache_Set(t *testing.T) {
	c := newGameUserSettingsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserSettings)
	err := c.ICache.(GameUserSettingsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameUserSettingsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameUserSettingsCache_Get(t *testing.T) {
	c := newGameUserSettingsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserSettings)
	err := c.ICache.(GameUserSettingsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserSettingsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameUserSettingsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameUserSettingsCache_MultiGet(t *testing.T) {
	c := newGameUserSettingsCache()
	defer c.Close()

	var testData []*model.GameUserSettings
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserSettings))
	}

	err := c.ICache.(GameUserSettingsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameUserSettingsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameUserSettings))
	}
}

func Test_gameUserSettingsCache_MultiSet(t *testing.T) {
	c := newGameUserSettingsCache()
	defer c.Close()

	var testData []*model.GameUserSettings
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameUserSettings))
	}

	err := c.ICache.(GameUserSettingsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserSettingsCache_Del(t *testing.T) {
	c := newGameUserSettingsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserSettings)
	err := c.ICache.(GameUserSettingsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameUserSettingsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameUserSettingsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameUserSettings)
	err := c.ICache.(GameUserSettingsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameUserSettingsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameUserSettingsCache(t *testing.T) {
	c := NewGameUserSettingsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameUserSettingsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameUserSettingsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
