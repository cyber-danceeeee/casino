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
	gameUserWalletsCachePrefixKey = "gameUserWallets:"
	// GameUserWalletsExpireTime expire time
	GameUserWalletsExpireTime = 5 * time.Minute
)

var _ GameUserWalletsCache = (*gameUserWalletsCache)(nil)

// GameUserWalletsCache cache interface
type GameUserWalletsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameUserWallets, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameUserWallets, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserWallets, error)
	MultiSet(ctx context.Context, data []*model.GameUserWallets, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameUserWalletsCache define a cache struct
type gameUserWalletsCache struct {
	cache cache.Cache
}

// NewGameUserWalletsCache new a cache
func NewGameUserWalletsCache(cacheType *database.CacheType) GameUserWalletsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserWallets{}
		})
		return &gameUserWalletsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserWallets{}
		})
		return &gameUserWalletsCache{cache: c}
	}

	return nil // no cache
}

// GetGameUserWalletsCacheKey cache key
func (c *gameUserWalletsCache) GetGameUserWalletsCacheKey(id uint64) string {
	return gameUserWalletsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameUserWalletsCache) Set(ctx context.Context, id uint64, data *model.GameUserWallets, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameUserWalletsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameUserWalletsCache) Get(ctx context.Context, id uint64) (*model.GameUserWallets, error) {
	var data *model.GameUserWallets
	cacheKey := c.GetGameUserWalletsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameUserWalletsCache) MultiSet(ctx context.Context, data []*model.GameUserWallets, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameUserWalletsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameUserWalletsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserWallets, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameUserWalletsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameUserWallets)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameUserWallets)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameUserWalletsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameUserWalletsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserWalletsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameUserWalletsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserWalletsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameUserWalletsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
