package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-dev-frame/sponge/pkg/cache"
	"github.com/go-dev-frame/sponge/pkg/encoding"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"casino/internal/database"
	"casino/internal/model"
)

const (
	// cache prefix key, must end with a colon
	gameUserActivityRecordsCachePrefixKey = "gameUserActivityRecords:"
	// GameUserActivityRecordsExpireTime expire time
	GameUserActivityRecordsExpireTime = 5 * time.Minute
)

var _ GameUserActivityRecordsCache = (*gameUserActivityRecordsCache)(nil)

// GameUserActivityRecordsCache cache interface
type GameUserActivityRecordsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameUserActivityRecords, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameUserActivityRecords, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserActivityRecords, error)
	MultiSet(ctx context.Context, data []*model.GameUserActivityRecords, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameUserActivityRecordsCache define a cache struct
type gameUserActivityRecordsCache struct {
	cache cache.Cache
}

// NewGameUserActivityRecordsCache new a cache
func NewGameUserActivityRecordsCache(cacheType *database.CacheType) GameUserActivityRecordsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserActivityRecords{}
		})
		return &gameUserActivityRecordsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserActivityRecords{}
		})
		return &gameUserActivityRecordsCache{cache: c}
	}

	return nil // no cache
}

// GetGameUserActivityRecordsCacheKey cache key
func (c *gameUserActivityRecordsCache) GetGameUserActivityRecordsCacheKey(id uint64) string {
	return gameUserActivityRecordsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameUserActivityRecordsCache) Set(ctx context.Context, id uint64, data *model.GameUserActivityRecords, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameUserActivityRecordsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameUserActivityRecordsCache) Get(ctx context.Context, id uint64) (*model.GameUserActivityRecords, error) {
	var data *model.GameUserActivityRecords
	cacheKey := c.GetGameUserActivityRecordsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameUserActivityRecordsCache) MultiSet(ctx context.Context, data []*model.GameUserActivityRecords, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameUserActivityRecordsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameUserActivityRecordsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserActivityRecords, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameUserActivityRecordsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameUserActivityRecords)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameUserActivityRecords)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameUserActivityRecordsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameUserActivityRecordsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserActivityRecordsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameUserActivityRecordsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserActivityRecordsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameUserActivityRecordsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
