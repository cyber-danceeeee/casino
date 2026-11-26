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
	gameCmsPagesCachePrefixKey = "gameCmsPages:"
	// GameCmsPagesExpireTime expire time
	GameCmsPagesExpireTime = 5 * time.Minute
)

var _ GameCmsPagesCache = (*gameCmsPagesCache)(nil)

// GameCmsPagesCache cache interface
type GameCmsPagesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameCmsPages, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameCmsPages, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameCmsPages, error)
	MultiSet(ctx context.Context, data []*model.GameCmsPages, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameCmsPagesCache define a cache struct
type gameCmsPagesCache struct {
	cache cache.Cache
}

// NewGameCmsPagesCache new a cache
func NewGameCmsPagesCache(cacheType *database.CacheType) GameCmsPagesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameCmsPages{}
		})
		return &gameCmsPagesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameCmsPages{}
		})
		return &gameCmsPagesCache{cache: c}
	}

	return nil // no cache
}

// GetGameCmsPagesCacheKey cache key
func (c *gameCmsPagesCache) GetGameCmsPagesCacheKey(id uint64) string {
	return gameCmsPagesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameCmsPagesCache) Set(ctx context.Context, id uint64, data *model.GameCmsPages, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameCmsPagesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameCmsPagesCache) Get(ctx context.Context, id uint64) (*model.GameCmsPages, error) {
	var data *model.GameCmsPages
	cacheKey := c.GetGameCmsPagesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameCmsPagesCache) MultiSet(ctx context.Context, data []*model.GameCmsPages, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameCmsPagesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameCmsPagesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameCmsPages, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameCmsPagesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameCmsPages)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameCmsPages)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameCmsPagesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameCmsPagesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameCmsPagesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameCmsPagesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameCmsPagesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameCmsPagesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
