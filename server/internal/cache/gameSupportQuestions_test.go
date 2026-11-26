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

func newGameSupportQuestionsCache() *gotest.Cache {
	record1 := &model.GameSupportQuestions{}
	record1.ID = 1
	record2 := &model.GameSupportQuestions{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewGameSupportQuestionsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_gameSupportQuestionsCache_Set(t *testing.T) {
	c := newGameSupportQuestionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportQuestions)
	err := c.ICache.(GameSupportQuestionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(GameSupportQuestionsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_gameSupportQuestionsCache_Get(t *testing.T) {
	c := newGameSupportQuestionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportQuestions)
	err := c.ICache.(GameSupportQuestionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameSupportQuestionsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(GameSupportQuestionsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_gameSupportQuestionsCache_MultiGet(t *testing.T) {
	c := newGameSupportQuestionsCache()
	defer c.Close()

	var testData []*model.GameSupportQuestions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameSupportQuestions))
	}

	err := c.ICache.(GameSupportQuestionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(GameSupportQuestionsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.GameSupportQuestions))
	}
}

func Test_gameSupportQuestionsCache_MultiSet(t *testing.T) {
	c := newGameSupportQuestionsCache()
	defer c.Close()

	var testData []*model.GameSupportQuestions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.GameSupportQuestions))
	}

	err := c.ICache.(GameSupportQuestionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameSupportQuestionsCache_Del(t *testing.T) {
	c := newGameSupportQuestionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportQuestions)
	err := c.ICache.(GameSupportQuestionsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_gameSupportQuestionsCache_SetCacheWithNotFound(t *testing.T) {
	c := newGameSupportQuestionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.GameSupportQuestions)
	err := c.ICache.(GameSupportQuestionsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(GameSupportQuestionsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewGameSupportQuestionsCache(t *testing.T) {
	c := NewGameSupportQuestionsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewGameSupportQuestionsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewGameSupportQuestionsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
