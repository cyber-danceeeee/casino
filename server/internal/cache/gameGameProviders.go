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
	gameGameProvidersCachePrefixKey = "gameGameProviders:"
	// GameGameProvidersExpireTime expire time
	GameGameProvidersExpireTime = 5 * time.Minute
)

var _ GameGameProvidersCache = (*gameGameProvidersCache)(nil)

// GameGameProvidersCache cache interface
type GameGameProvidersCache interface {
	Set(ctx context.Context, id uint64, data *model.GameGameProviders, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameGameProviders, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameGameProviders, error)
	MultiSet(ctx context.Context, data []*model.GameGameProviders, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameGameProvidersCache define a cache struct
type gameGameProvidersCache struct {
	cache cache.Cache
}

// NewGameGameProvidersCache new a cache
func NewGameGameProvidersCache(cacheType *database.CacheType) GameGameProvidersCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameGameProviders{}
		})
		return &gameGameProvidersCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameGameProviders{}
		})
		return &gameGameProvidersCache{cache: c}
	}

	return nil // no cache
}

// GetGameGameProvidersCacheKey cache key
func (c *gameGameProvidersCache) GetGameGameProvidersCacheKey(id uint64) string {
	return gameGameProvidersCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameGameProvidersCache) Set(ctx context.Context, id uint64, data *model.GameGameProviders, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameGameProvidersCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameGameProvidersCache) Get(ctx context.Context, id uint64) (*model.GameGameProviders, error) {
	var data *model.GameGameProviders
	cacheKey := c.GetGameGameProvidersCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameGameProvidersCache) MultiSet(ctx context.Context, data []*model.GameGameProviders, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameGameProvidersCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameGameProvidersCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameGameProviders, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameGameProvidersCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameGameProviders)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameGameProviders)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameGameProvidersCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameGameProvidersCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameGameProvidersCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameGameProvidersCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameGameProvidersCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameGameProvidersCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
