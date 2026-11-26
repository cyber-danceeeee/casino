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
	gameVipLevelsCachePrefixKey = "gameVipLevels:"
	// GameVipLevelsExpireTime expire time
	GameVipLevelsExpireTime = 5 * time.Minute
)

var _ GameVipLevelsCache = (*gameVipLevelsCache)(nil)

// GameVipLevelsCache cache interface
type GameVipLevelsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameVipLevels, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameVipLevels, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameVipLevels, error)
	MultiSet(ctx context.Context, data []*model.GameVipLevels, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameVipLevelsCache define a cache struct
type gameVipLevelsCache struct {
	cache cache.Cache
}

// NewGameVipLevelsCache new a cache
func NewGameVipLevelsCache(cacheType *database.CacheType) GameVipLevelsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameVipLevels{}
		})
		return &gameVipLevelsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameVipLevels{}
		})
		return &gameVipLevelsCache{cache: c}
	}

	return nil // no cache
}

// GetGameVipLevelsCacheKey cache key
func (c *gameVipLevelsCache) GetGameVipLevelsCacheKey(id uint64) string {
	return gameVipLevelsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameVipLevelsCache) Set(ctx context.Context, id uint64, data *model.GameVipLevels, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameVipLevelsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameVipLevelsCache) Get(ctx context.Context, id uint64) (*model.GameVipLevels, error) {
	var data *model.GameVipLevels
	cacheKey := c.GetGameVipLevelsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameVipLevelsCache) MultiSet(ctx context.Context, data []*model.GameVipLevels, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameVipLevelsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameVipLevelsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameVipLevels, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameVipLevelsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameVipLevels)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameVipLevels)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameVipLevelsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameVipLevelsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameVipLevelsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameVipLevelsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameVipLevelsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameVipLevelsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
