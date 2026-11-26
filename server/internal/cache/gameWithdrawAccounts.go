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
	gameWithdrawAccountsCachePrefixKey = "gameWithdrawAccounts:"
	// GameWithdrawAccountsExpireTime expire time
	GameWithdrawAccountsExpireTime = 5 * time.Minute
)

var _ GameWithdrawAccountsCache = (*gameWithdrawAccountsCache)(nil)

// GameWithdrawAccountsCache cache interface
type GameWithdrawAccountsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameWithdrawAccounts, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameWithdrawAccounts, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameWithdrawAccounts, error)
	MultiSet(ctx context.Context, data []*model.GameWithdrawAccounts, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameWithdrawAccountsCache define a cache struct
type gameWithdrawAccountsCache struct {
	cache cache.Cache
}

// NewGameWithdrawAccountsCache new a cache
func NewGameWithdrawAccountsCache(cacheType *database.CacheType) GameWithdrawAccountsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameWithdrawAccounts{}
		})
		return &gameWithdrawAccountsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameWithdrawAccounts{}
		})
		return &gameWithdrawAccountsCache{cache: c}
	}

	return nil // no cache
}

// GetGameWithdrawAccountsCacheKey cache key
func (c *gameWithdrawAccountsCache) GetGameWithdrawAccountsCacheKey(id uint64) string {
	return gameWithdrawAccountsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameWithdrawAccountsCache) Set(ctx context.Context, id uint64, data *model.GameWithdrawAccounts, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameWithdrawAccountsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameWithdrawAccountsCache) Get(ctx context.Context, id uint64) (*model.GameWithdrawAccounts, error) {
	var data *model.GameWithdrawAccounts
	cacheKey := c.GetGameWithdrawAccountsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameWithdrawAccountsCache) MultiSet(ctx context.Context, data []*model.GameWithdrawAccounts, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameWithdrawAccountsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameWithdrawAccountsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameWithdrawAccounts, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameWithdrawAccountsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameWithdrawAccounts)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameWithdrawAccounts)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameWithdrawAccountsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameWithdrawAccountsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameWithdrawAccountsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameWithdrawAccountsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameWithdrawAccountsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameWithdrawAccountsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
