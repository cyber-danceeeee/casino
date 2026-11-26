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
	gameBannersCachePrefixKey = "gameBanners:"
	// GameBannersExpireTime expire time
	GameBannersExpireTime = 5 * time.Minute
)

var _ GameBannersCache = (*gameBannersCache)(nil)

// GameBannersCache cache interface
type GameBannersCache interface {
	Set(ctx context.Context, id uint64, data *model.GameBanners, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameBanners, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameBanners, error)
	MultiSet(ctx context.Context, data []*model.GameBanners, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameBannersCache define a cache struct
type gameBannersCache struct {
	cache cache.Cache
}

// NewGameBannersCache new a cache
func NewGameBannersCache(cacheType *database.CacheType) GameBannersCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameBanners{}
		})
		return &gameBannersCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameBanners{}
		})
		return &gameBannersCache{cache: c}
	}

	return nil // no cache
}

// GetGameBannersCacheKey cache key
func (c *gameBannersCache) GetGameBannersCacheKey(id uint64) string {
	return gameBannersCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameBannersCache) Set(ctx context.Context, id uint64, data *model.GameBanners, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameBannersCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameBannersCache) Get(ctx context.Context, id uint64) (*model.GameBanners, error) {
	var data *model.GameBanners
	cacheKey := c.GetGameBannersCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameBannersCache) MultiSet(ctx context.Context, data []*model.GameBanners, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameBannersCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameBannersCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameBanners, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameBannersCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameBanners)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameBanners)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameBannersCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameBannersCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameBannersCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameBannersCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameBannersCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameBannersCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
