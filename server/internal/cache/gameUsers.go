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
	gameUsersCachePrefixKey = "gameUsers:"
	// GameUsersExpireTime expire time
	GameUsersExpireTime = 5 * time.Minute
)

var _ GameUsersCache = (*gameUsersCache)(nil)

// GameUsersCache cache interface
type GameUsersCache interface {
	Set(ctx context.Context, id uint64, data *model.GameUsers, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameUsers, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUsers, error)
	MultiSet(ctx context.Context, data []*model.GameUsers, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameUsersCache define a cache struct
type gameUsersCache struct {
	cache cache.Cache
}

// NewGameUsersCache new a cache
func NewGameUsersCache(cacheType *database.CacheType) GameUsersCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUsers{}
		})
		return &gameUsersCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUsers{}
		})
		return &gameUsersCache{cache: c}
	}

	return nil // no cache
}

// GetGameUsersCacheKey cache key
func (c *gameUsersCache) GetGameUsersCacheKey(id uint64) string {
	return gameUsersCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameUsersCache) Set(ctx context.Context, id uint64, data *model.GameUsers, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameUsersCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameUsersCache) Get(ctx context.Context, id uint64) (*model.GameUsers, error) {
	var data *model.GameUsers
	cacheKey := c.GetGameUsersCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameUsersCache) MultiSet(ctx context.Context, data []*model.GameUsers, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameUsersCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameUsersCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUsers, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameUsersCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameUsers)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameUsers)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameUsersCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameUsersCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUsersCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameUsersCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUsersCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameUsersCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
