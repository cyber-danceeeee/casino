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
	gameRankingRecordsCachePrefixKey = "gameRankingRecords:"
	// GameRankingRecordsExpireTime expire time
	GameRankingRecordsExpireTime = 5 * time.Minute
)

var _ GameRankingRecordsCache = (*gameRankingRecordsCache)(nil)

// GameRankingRecordsCache cache interface
type GameRankingRecordsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameRankingRecords, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameRankingRecords, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameRankingRecords, error)
	MultiSet(ctx context.Context, data []*model.GameRankingRecords, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameRankingRecordsCache define a cache struct
type gameRankingRecordsCache struct {
	cache cache.Cache
}

// NewGameRankingRecordsCache new a cache
func NewGameRankingRecordsCache(cacheType *database.CacheType) GameRankingRecordsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameRankingRecords{}
		})
		return &gameRankingRecordsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameRankingRecords{}
		})
		return &gameRankingRecordsCache{cache: c}
	}

	return nil // no cache
}

// GetGameRankingRecordsCacheKey cache key
func (c *gameRankingRecordsCache) GetGameRankingRecordsCacheKey(id uint64) string {
	return gameRankingRecordsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameRankingRecordsCache) Set(ctx context.Context, id uint64, data *model.GameRankingRecords, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameRankingRecordsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameRankingRecordsCache) Get(ctx context.Context, id uint64) (*model.GameRankingRecords, error) {
	var data *model.GameRankingRecords
	cacheKey := c.GetGameRankingRecordsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameRankingRecordsCache) MultiSet(ctx context.Context, data []*model.GameRankingRecords, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameRankingRecordsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameRankingRecordsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameRankingRecords, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameRankingRecordsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameRankingRecords)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameRankingRecords)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameRankingRecordsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameRankingRecordsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameRankingRecordsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameRankingRecordsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameRankingRecordsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameRankingRecordsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
