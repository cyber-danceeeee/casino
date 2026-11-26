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
	gameUserSessionsCachePrefixKey = "gameUserSessions:"
	// GameUserSessionsExpireTime expire time
	GameUserSessionsExpireTime = 5 * time.Minute
)

var _ GameUserSessionsCache = (*gameUserSessionsCache)(nil)

// GameUserSessionsCache cache interface
type GameUserSessionsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameUserSessions, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameUserSessions, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserSessions, error)
	MultiSet(ctx context.Context, data []*model.GameUserSessions, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameUserSessionsCache define a cache struct
type gameUserSessionsCache struct {
	cache cache.Cache
}

// NewGameUserSessionsCache new a cache
func NewGameUserSessionsCache(cacheType *database.CacheType) GameUserSessionsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserSessions{}
		})
		return &gameUserSessionsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserSessions{}
		})
		return &gameUserSessionsCache{cache: c}
	}

	return nil // no cache
}

// GetGameUserSessionsCacheKey cache key
func (c *gameUserSessionsCache) GetGameUserSessionsCacheKey(id uint64) string {
	return gameUserSessionsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameUserSessionsCache) Set(ctx context.Context, id uint64, data *model.GameUserSessions, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameUserSessionsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameUserSessionsCache) Get(ctx context.Context, id uint64) (*model.GameUserSessions, error) {
	var data *model.GameUserSessions
	cacheKey := c.GetGameUserSessionsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameUserSessionsCache) MultiSet(ctx context.Context, data []*model.GameUserSessions, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameUserSessionsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameUserSessionsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserSessions, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameUserSessionsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameUserSessions)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameUserSessions)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameUserSessionsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameUserSessionsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserSessionsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameUserSessionsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserSessionsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameUserSessionsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
