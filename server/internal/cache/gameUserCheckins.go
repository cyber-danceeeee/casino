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
	gameUserCheckinsCachePrefixKey = "gameUserCheckins:"
	// GameUserCheckinsExpireTime expire time
	GameUserCheckinsExpireTime = 5 * time.Minute
)

var _ GameUserCheckinsCache = (*gameUserCheckinsCache)(nil)

// GameUserCheckinsCache cache interface
type GameUserCheckinsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameUserCheckins, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameUserCheckins, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserCheckins, error)
	MultiSet(ctx context.Context, data []*model.GameUserCheckins, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameUserCheckinsCache define a cache struct
type gameUserCheckinsCache struct {
	cache cache.Cache
}

// NewGameUserCheckinsCache new a cache
func NewGameUserCheckinsCache(cacheType *database.CacheType) GameUserCheckinsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserCheckins{}
		})
		return &gameUserCheckinsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserCheckins{}
		})
		return &gameUserCheckinsCache{cache: c}
	}

	return nil // no cache
}

// GetGameUserCheckinsCacheKey cache key
func (c *gameUserCheckinsCache) GetGameUserCheckinsCacheKey(id uint64) string {
	return gameUserCheckinsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameUserCheckinsCache) Set(ctx context.Context, id uint64, data *model.GameUserCheckins, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameUserCheckinsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameUserCheckinsCache) Get(ctx context.Context, id uint64) (*model.GameUserCheckins, error) {
	var data *model.GameUserCheckins
	cacheKey := c.GetGameUserCheckinsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameUserCheckinsCache) MultiSet(ctx context.Context, data []*model.GameUserCheckins, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameUserCheckinsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameUserCheckinsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserCheckins, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameUserCheckinsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameUserCheckins)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameUserCheckins)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameUserCheckinsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameUserCheckinsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserCheckinsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameUserCheckinsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserCheckinsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameUserCheckinsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
