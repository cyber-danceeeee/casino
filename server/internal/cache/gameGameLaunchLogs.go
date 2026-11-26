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
	gameGameLaunchLogsCachePrefixKey = "gameGameLaunchLogs:"
	// GameGameLaunchLogsExpireTime expire time
	GameGameLaunchLogsExpireTime = 5 * time.Minute
)

var _ GameGameLaunchLogsCache = (*gameGameLaunchLogsCache)(nil)

// GameGameLaunchLogsCache cache interface
type GameGameLaunchLogsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameGameLaunchLogs, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameGameLaunchLogs, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameGameLaunchLogs, error)
	MultiSet(ctx context.Context, data []*model.GameGameLaunchLogs, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameGameLaunchLogsCache define a cache struct
type gameGameLaunchLogsCache struct {
	cache cache.Cache
}

// NewGameGameLaunchLogsCache new a cache
func NewGameGameLaunchLogsCache(cacheType *database.CacheType) GameGameLaunchLogsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameGameLaunchLogs{}
		})
		return &gameGameLaunchLogsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameGameLaunchLogs{}
		})
		return &gameGameLaunchLogsCache{cache: c}
	}

	return nil // no cache
}

// GetGameGameLaunchLogsCacheKey cache key
func (c *gameGameLaunchLogsCache) GetGameGameLaunchLogsCacheKey(id uint64) string {
	return gameGameLaunchLogsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameGameLaunchLogsCache) Set(ctx context.Context, id uint64, data *model.GameGameLaunchLogs, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameGameLaunchLogsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameGameLaunchLogsCache) Get(ctx context.Context, id uint64) (*model.GameGameLaunchLogs, error) {
	var data *model.GameGameLaunchLogs
	cacheKey := c.GetGameGameLaunchLogsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameGameLaunchLogsCache) MultiSet(ctx context.Context, data []*model.GameGameLaunchLogs, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameGameLaunchLogsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameGameLaunchLogsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameGameLaunchLogs, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameGameLaunchLogsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameGameLaunchLogs)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameGameLaunchLogs)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameGameLaunchLogsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameGameLaunchLogsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameGameLaunchLogsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameGameLaunchLogsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameGameLaunchLogsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameGameLaunchLogsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
