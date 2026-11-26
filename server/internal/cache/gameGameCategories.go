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
	gameGameCategoriesCachePrefixKey = "gameGameCategories:"
	// GameGameCategoriesExpireTime expire time
	GameGameCategoriesExpireTime = 5 * time.Minute
)

var _ GameGameCategoriesCache = (*gameGameCategoriesCache)(nil)

// GameGameCategoriesCache cache interface
type GameGameCategoriesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameGameCategories, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameGameCategories, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameGameCategories, error)
	MultiSet(ctx context.Context, data []*model.GameGameCategories, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameGameCategoriesCache define a cache struct
type gameGameCategoriesCache struct {
	cache cache.Cache
}

// NewGameGameCategoriesCache new a cache
func NewGameGameCategoriesCache(cacheType *database.CacheType) GameGameCategoriesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameGameCategories{}
		})
		return &gameGameCategoriesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameGameCategories{}
		})
		return &gameGameCategoriesCache{cache: c}
	}

	return nil // no cache
}

// GetGameGameCategoriesCacheKey cache key
func (c *gameGameCategoriesCache) GetGameGameCategoriesCacheKey(id uint64) string {
	return gameGameCategoriesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameGameCategoriesCache) Set(ctx context.Context, id uint64, data *model.GameGameCategories, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameGameCategoriesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameGameCategoriesCache) Get(ctx context.Context, id uint64) (*model.GameGameCategories, error) {
	var data *model.GameGameCategories
	cacheKey := c.GetGameGameCategoriesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameGameCategoriesCache) MultiSet(ctx context.Context, data []*model.GameGameCategories, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameGameCategoriesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameGameCategoriesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameGameCategories, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameGameCategoriesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameGameCategories)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameGameCategories)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameGameCategoriesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameGameCategoriesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameGameCategoriesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameGameCategoriesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameGameCategoriesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameGameCategoriesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
