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
	gameAgentCommissionRecordsCachePrefixKey = "gameAgentCommissionRecords:"
	// GameAgentCommissionRecordsExpireTime expire time
	GameAgentCommissionRecordsExpireTime = 5 * time.Minute
)

var _ GameAgentCommissionRecordsCache = (*gameAgentCommissionRecordsCache)(nil)

// GameAgentCommissionRecordsCache cache interface
type GameAgentCommissionRecordsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameAgentCommissionRecords, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameAgentCommissionRecords, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameAgentCommissionRecords, error)
	MultiSet(ctx context.Context, data []*model.GameAgentCommissionRecords, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameAgentCommissionRecordsCache define a cache struct
type gameAgentCommissionRecordsCache struct {
	cache cache.Cache
}

// NewGameAgentCommissionRecordsCache new a cache
func NewGameAgentCommissionRecordsCache(cacheType *database.CacheType) GameAgentCommissionRecordsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameAgentCommissionRecords{}
		})
		return &gameAgentCommissionRecordsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameAgentCommissionRecords{}
		})
		return &gameAgentCommissionRecordsCache{cache: c}
	}

	return nil // no cache
}

// GetGameAgentCommissionRecordsCacheKey cache key
func (c *gameAgentCommissionRecordsCache) GetGameAgentCommissionRecordsCacheKey(id uint64) string {
	return gameAgentCommissionRecordsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameAgentCommissionRecordsCache) Set(ctx context.Context, id uint64, data *model.GameAgentCommissionRecords, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameAgentCommissionRecordsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameAgentCommissionRecordsCache) Get(ctx context.Context, id uint64) (*model.GameAgentCommissionRecords, error) {
	var data *model.GameAgentCommissionRecords
	cacheKey := c.GetGameAgentCommissionRecordsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameAgentCommissionRecordsCache) MultiSet(ctx context.Context, data []*model.GameAgentCommissionRecords, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameAgentCommissionRecordsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameAgentCommissionRecordsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameAgentCommissionRecords, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameAgentCommissionRecordsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameAgentCommissionRecords)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameAgentCommissionRecords)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameAgentCommissionRecordsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameAgentCommissionRecordsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameAgentCommissionRecordsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameAgentCommissionRecordsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameAgentCommissionRecordsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameAgentCommissionRecordsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
