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
	gameActivityEventsCachePrefixKey = "gameActivityEvents:"
	// GameActivityEventsExpireTime expire time
	GameActivityEventsExpireTime = 5 * time.Minute
)

var _ GameActivityEventsCache = (*gameActivityEventsCache)(nil)

// GameActivityEventsCache cache interface
type GameActivityEventsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameActivityEvents, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameActivityEvents, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameActivityEvents, error)
	MultiSet(ctx context.Context, data []*model.GameActivityEvents, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameActivityEventsCache define a cache struct
type gameActivityEventsCache struct {
	cache cache.Cache
}

// NewGameActivityEventsCache new a cache
func NewGameActivityEventsCache(cacheType *database.CacheType) GameActivityEventsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameActivityEvents{}
		})
		return &gameActivityEventsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameActivityEvents{}
		})
		return &gameActivityEventsCache{cache: c}
	}

	return nil // no cache
}

// GetGameActivityEventsCacheKey cache key
func (c *gameActivityEventsCache) GetGameActivityEventsCacheKey(id uint64) string {
	return gameActivityEventsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameActivityEventsCache) Set(ctx context.Context, id uint64, data *model.GameActivityEvents, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameActivityEventsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameActivityEventsCache) Get(ctx context.Context, id uint64) (*model.GameActivityEvents, error) {
	var data *model.GameActivityEvents
	cacheKey := c.GetGameActivityEventsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameActivityEventsCache) MultiSet(ctx context.Context, data []*model.GameActivityEvents, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameActivityEventsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameActivityEventsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameActivityEvents, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameActivityEventsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameActivityEvents)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameActivityEvents)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameActivityEventsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameActivityEventsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameActivityEventsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameActivityEventsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameActivityEventsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameActivityEventsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
