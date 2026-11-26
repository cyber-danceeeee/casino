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
	gameDepositOrdersCachePrefixKey = "gameDepositOrders:"
	// GameDepositOrdersExpireTime expire time
	GameDepositOrdersExpireTime = 5 * time.Minute
)

var _ GameDepositOrdersCache = (*gameDepositOrdersCache)(nil)

// GameDepositOrdersCache cache interface
type GameDepositOrdersCache interface {
	Set(ctx context.Context, id uint64, data *model.GameDepositOrders, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameDepositOrders, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameDepositOrders, error)
	MultiSet(ctx context.Context, data []*model.GameDepositOrders, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameDepositOrdersCache define a cache struct
type gameDepositOrdersCache struct {
	cache cache.Cache
}

// NewGameDepositOrdersCache new a cache
func NewGameDepositOrdersCache(cacheType *database.CacheType) GameDepositOrdersCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameDepositOrders{}
		})
		return &gameDepositOrdersCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameDepositOrders{}
		})
		return &gameDepositOrdersCache{cache: c}
	}

	return nil // no cache
}

// GetGameDepositOrdersCacheKey cache key
func (c *gameDepositOrdersCache) GetGameDepositOrdersCacheKey(id uint64) string {
	return gameDepositOrdersCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameDepositOrdersCache) Set(ctx context.Context, id uint64, data *model.GameDepositOrders, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameDepositOrdersCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameDepositOrdersCache) Get(ctx context.Context, id uint64) (*model.GameDepositOrders, error) {
	var data *model.GameDepositOrders
	cacheKey := c.GetGameDepositOrdersCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameDepositOrdersCache) MultiSet(ctx context.Context, data []*model.GameDepositOrders, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameDepositOrdersCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameDepositOrdersCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameDepositOrders, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameDepositOrdersCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameDepositOrders)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameDepositOrders)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameDepositOrdersCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameDepositOrdersCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameDepositOrdersCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameDepositOrdersCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameDepositOrdersCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameDepositOrdersCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
