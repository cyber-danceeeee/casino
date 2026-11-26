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
	gameBetRecordsCachePrefixKey = "gameBetRecords:"
	// GameBetRecordsExpireTime expire time
	GameBetRecordsExpireTime = 5 * time.Minute
)

var _ GameBetRecordsCache = (*gameBetRecordsCache)(nil)

// GameBetRecordsCache cache interface
type GameBetRecordsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameBetRecords, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameBetRecords, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameBetRecords, error)
	MultiSet(ctx context.Context, data []*model.GameBetRecords, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameBetRecordsCache define a cache struct
type gameBetRecordsCache struct {
	cache cache.Cache
}

// NewGameBetRecordsCache new a cache
func NewGameBetRecordsCache(cacheType *database.CacheType) GameBetRecordsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameBetRecords{}
		})
		return &gameBetRecordsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameBetRecords{}
		})
		return &gameBetRecordsCache{cache: c}
	}

	return nil // no cache
}

// GetGameBetRecordsCacheKey cache key
func (c *gameBetRecordsCache) GetGameBetRecordsCacheKey(id uint64) string {
	return gameBetRecordsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameBetRecordsCache) Set(ctx context.Context, id uint64, data *model.GameBetRecords, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameBetRecordsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameBetRecordsCache) Get(ctx context.Context, id uint64) (*model.GameBetRecords, error) {
	var data *model.GameBetRecords
	cacheKey := c.GetGameBetRecordsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameBetRecordsCache) MultiSet(ctx context.Context, data []*model.GameBetRecords, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameBetRecordsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameBetRecordsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameBetRecords, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameBetRecordsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameBetRecords)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameBetRecords)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameBetRecordsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameBetRecordsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameBetRecordsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameBetRecordsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameBetRecordsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameBetRecordsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
