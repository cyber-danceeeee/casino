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
	gameSupportCategoriesCachePrefixKey = "gameSupportCategories:"
	// GameSupportCategoriesExpireTime expire time
	GameSupportCategoriesExpireTime = 5 * time.Minute
)

var _ GameSupportCategoriesCache = (*gameSupportCategoriesCache)(nil)

// GameSupportCategoriesCache cache interface
type GameSupportCategoriesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameSupportCategories, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameSupportCategories, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameSupportCategories, error)
	MultiSet(ctx context.Context, data []*model.GameSupportCategories, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameSupportCategoriesCache define a cache struct
type gameSupportCategoriesCache struct {
	cache cache.Cache
}

// NewGameSupportCategoriesCache new a cache
func NewGameSupportCategoriesCache(cacheType *database.CacheType) GameSupportCategoriesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameSupportCategories{}
		})
		return &gameSupportCategoriesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameSupportCategories{}
		})
		return &gameSupportCategoriesCache{cache: c}
	}

	return nil // no cache
}

// GetGameSupportCategoriesCacheKey cache key
func (c *gameSupportCategoriesCache) GetGameSupportCategoriesCacheKey(id uint64) string {
	return gameSupportCategoriesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameSupportCategoriesCache) Set(ctx context.Context, id uint64, data *model.GameSupportCategories, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameSupportCategoriesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameSupportCategoriesCache) Get(ctx context.Context, id uint64) (*model.GameSupportCategories, error) {
	var data *model.GameSupportCategories
	cacheKey := c.GetGameSupportCategoriesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameSupportCategoriesCache) MultiSet(ctx context.Context, data []*model.GameSupportCategories, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameSupportCategoriesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameSupportCategoriesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameSupportCategories, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameSupportCategoriesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameSupportCategories)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameSupportCategories)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameSupportCategoriesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameSupportCategoriesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameSupportCategoriesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameSupportCategoriesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameSupportCategoriesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameSupportCategoriesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
