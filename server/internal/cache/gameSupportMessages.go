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
	gameSupportMessagesCachePrefixKey = "gameSupportMessages:"
	// GameSupportMessagesExpireTime expire time
	GameSupportMessagesExpireTime = 5 * time.Minute
)

var _ GameSupportMessagesCache = (*gameSupportMessagesCache)(nil)

// GameSupportMessagesCache cache interface
type GameSupportMessagesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameSupportMessages, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameSupportMessages, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameSupportMessages, error)
	MultiSet(ctx context.Context, data []*model.GameSupportMessages, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameSupportMessagesCache define a cache struct
type gameSupportMessagesCache struct {
	cache cache.Cache
}

// NewGameSupportMessagesCache new a cache
func NewGameSupportMessagesCache(cacheType *database.CacheType) GameSupportMessagesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameSupportMessages{}
		})
		return &gameSupportMessagesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameSupportMessages{}
		})
		return &gameSupportMessagesCache{cache: c}
	}

	return nil // no cache
}

// GetGameSupportMessagesCacheKey cache key
func (c *gameSupportMessagesCache) GetGameSupportMessagesCacheKey(id uint64) string {
	return gameSupportMessagesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameSupportMessagesCache) Set(ctx context.Context, id uint64, data *model.GameSupportMessages, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameSupportMessagesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameSupportMessagesCache) Get(ctx context.Context, id uint64) (*model.GameSupportMessages, error) {
	var data *model.GameSupportMessages
	cacheKey := c.GetGameSupportMessagesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameSupportMessagesCache) MultiSet(ctx context.Context, data []*model.GameSupportMessages, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameSupportMessagesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameSupportMessagesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameSupportMessages, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameSupportMessagesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameSupportMessages)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameSupportMessages)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameSupportMessagesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameSupportMessagesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameSupportMessagesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameSupportMessagesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameSupportMessagesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameSupportMessagesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
