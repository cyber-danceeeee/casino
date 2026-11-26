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
	gameWalletTransactionsCachePrefixKey = "gameWalletTransactions:"
	// GameWalletTransactionsExpireTime expire time
	GameWalletTransactionsExpireTime = 5 * time.Minute
)

var _ GameWalletTransactionsCache = (*gameWalletTransactionsCache)(nil)

// GameWalletTransactionsCache cache interface
type GameWalletTransactionsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameWalletTransactions, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameWalletTransactions, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameWalletTransactions, error)
	MultiSet(ctx context.Context, data []*model.GameWalletTransactions, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameWalletTransactionsCache define a cache struct
type gameWalletTransactionsCache struct {
	cache cache.Cache
}

// NewGameWalletTransactionsCache new a cache
func NewGameWalletTransactionsCache(cacheType *database.CacheType) GameWalletTransactionsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameWalletTransactions{}
		})
		return &gameWalletTransactionsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameWalletTransactions{}
		})
		return &gameWalletTransactionsCache{cache: c}
	}

	return nil // no cache
}

// GetGameWalletTransactionsCacheKey cache key
func (c *gameWalletTransactionsCache) GetGameWalletTransactionsCacheKey(id uint64) string {
	return gameWalletTransactionsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameWalletTransactionsCache) Set(ctx context.Context, id uint64, data *model.GameWalletTransactions, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameWalletTransactionsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameWalletTransactionsCache) Get(ctx context.Context, id uint64) (*model.GameWalletTransactions, error) {
	var data *model.GameWalletTransactions
	cacheKey := c.GetGameWalletTransactionsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameWalletTransactionsCache) MultiSet(ctx context.Context, data []*model.GameWalletTransactions, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameWalletTransactionsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameWalletTransactionsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameWalletTransactions, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameWalletTransactionsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameWalletTransactions)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameWalletTransactions)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameWalletTransactionsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameWalletTransactionsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameWalletTransactionsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameWalletTransactionsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameWalletTransactionsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameWalletTransactionsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
