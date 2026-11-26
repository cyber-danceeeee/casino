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
	gameNoticesCachePrefixKey = "gameNotices:"
	// GameNoticesExpireTime expire time
	GameNoticesExpireTime = 5 * time.Minute
)

var _ GameNoticesCache = (*gameNoticesCache)(nil)

// GameNoticesCache cache interface
type GameNoticesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameNotices, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameNotices, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameNotices, error)
	MultiSet(ctx context.Context, data []*model.GameNotices, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameNoticesCache define a cache struct
type gameNoticesCache struct {
	cache cache.Cache
}

// NewGameNoticesCache new a cache
func NewGameNoticesCache(cacheType *database.CacheType) GameNoticesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameNotices{}
		})
		return &gameNoticesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameNotices{}
		})
		return &gameNoticesCache{cache: c}
	}

	return nil // no cache
}

// GetGameNoticesCacheKey cache key
func (c *gameNoticesCache) GetGameNoticesCacheKey(id uint64) string {
	return gameNoticesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameNoticesCache) Set(ctx context.Context, id uint64, data *model.GameNotices, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameNoticesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameNoticesCache) Get(ctx context.Context, id uint64) (*model.GameNotices, error) {
	var data *model.GameNotices
	cacheKey := c.GetGameNoticesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameNoticesCache) MultiSet(ctx context.Context, data []*model.GameNotices, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameNoticesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameNoticesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameNotices, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameNoticesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameNotices)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameNotices)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameNoticesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameNoticesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameNoticesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameNoticesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameNoticesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameNoticesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
