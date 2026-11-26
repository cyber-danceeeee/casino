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
	gameWithdrawOrdersCachePrefixKey = "gameWithdrawOrders:"
	// GameWithdrawOrdersExpireTime expire time
	GameWithdrawOrdersExpireTime = 5 * time.Minute
)

var _ GameWithdrawOrdersCache = (*gameWithdrawOrdersCache)(nil)

// GameWithdrawOrdersCache cache interface
type GameWithdrawOrdersCache interface {
	Set(ctx context.Context, id uint64, data *model.GameWithdrawOrders, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameWithdrawOrders, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameWithdrawOrders, error)
	MultiSet(ctx context.Context, data []*model.GameWithdrawOrders, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameWithdrawOrdersCache define a cache struct
type gameWithdrawOrdersCache struct {
	cache cache.Cache
}

// NewGameWithdrawOrdersCache new a cache
func NewGameWithdrawOrdersCache(cacheType *database.CacheType) GameWithdrawOrdersCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameWithdrawOrders{}
		})
		return &gameWithdrawOrdersCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameWithdrawOrders{}
		})
		return &gameWithdrawOrdersCache{cache: c}
	}

	return nil // no cache
}

// GetGameWithdrawOrdersCacheKey cache key
func (c *gameWithdrawOrdersCache) GetGameWithdrawOrdersCacheKey(id uint64) string {
	return gameWithdrawOrdersCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameWithdrawOrdersCache) Set(ctx context.Context, id uint64, data *model.GameWithdrawOrders, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameWithdrawOrdersCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameWithdrawOrdersCache) Get(ctx context.Context, id uint64) (*model.GameWithdrawOrders, error) {
	var data *model.GameWithdrawOrders
	cacheKey := c.GetGameWithdrawOrdersCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameWithdrawOrdersCache) MultiSet(ctx context.Context, data []*model.GameWithdrawOrders, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameWithdrawOrdersCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameWithdrawOrdersCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameWithdrawOrders, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameWithdrawOrdersCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameWithdrawOrders)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameWithdrawOrders)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameWithdrawOrdersCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameWithdrawOrdersCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameWithdrawOrdersCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameWithdrawOrdersCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameWithdrawOrdersCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameWithdrawOrdersCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
