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
	gameActivityAssetsCachePrefixKey = "gameActivityAssets:"
	// GameActivityAssetsExpireTime expire time
	GameActivityAssetsExpireTime = 5 * time.Minute
)

var _ GameActivityAssetsCache = (*gameActivityAssetsCache)(nil)

// GameActivityAssetsCache cache interface
type GameActivityAssetsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameActivityAssets, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameActivityAssets, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameActivityAssets, error)
	MultiSet(ctx context.Context, data []*model.GameActivityAssets, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameActivityAssetsCache define a cache struct
type gameActivityAssetsCache struct {
	cache cache.Cache
}

// NewGameActivityAssetsCache new a cache
func NewGameActivityAssetsCache(cacheType *database.CacheType) GameActivityAssetsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameActivityAssets{}
		})
		return &gameActivityAssetsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameActivityAssets{}
		})
		return &gameActivityAssetsCache{cache: c}
	}

	return nil // no cache
}

// GetGameActivityAssetsCacheKey cache key
func (c *gameActivityAssetsCache) GetGameActivityAssetsCacheKey(id uint64) string {
	return gameActivityAssetsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameActivityAssetsCache) Set(ctx context.Context, id uint64, data *model.GameActivityAssets, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameActivityAssetsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameActivityAssetsCache) Get(ctx context.Context, id uint64) (*model.GameActivityAssets, error) {
	var data *model.GameActivityAssets
	cacheKey := c.GetGameActivityAssetsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameActivityAssetsCache) MultiSet(ctx context.Context, data []*model.GameActivityAssets, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameActivityAssetsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameActivityAssetsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameActivityAssets, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameActivityAssetsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameActivityAssets)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameActivityAssets)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameActivityAssetsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameActivityAssetsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameActivityAssetsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameActivityAssetsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameActivityAssetsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameActivityAssetsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
