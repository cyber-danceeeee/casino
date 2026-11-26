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
	gameUserDevicesCachePrefixKey = "gameUserDevices:"
	// GameUserDevicesExpireTime expire time
	GameUserDevicesExpireTime = 5 * time.Minute
)

var _ GameUserDevicesCache = (*gameUserDevicesCache)(nil)

// GameUserDevicesCache cache interface
type GameUserDevicesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameUserDevices, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameUserDevices, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserDevices, error)
	MultiSet(ctx context.Context, data []*model.GameUserDevices, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameUserDevicesCache define a cache struct
type gameUserDevicesCache struct {
	cache cache.Cache
}

// NewGameUserDevicesCache new a cache
func NewGameUserDevicesCache(cacheType *database.CacheType) GameUserDevicesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserDevices{}
		})
		return &gameUserDevicesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserDevices{}
		})
		return &gameUserDevicesCache{cache: c}
	}

	return nil // no cache
}

// GetGameUserDevicesCacheKey cache key
func (c *gameUserDevicesCache) GetGameUserDevicesCacheKey(id uint64) string {
	return gameUserDevicesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameUserDevicesCache) Set(ctx context.Context, id uint64, data *model.GameUserDevices, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameUserDevicesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameUserDevicesCache) Get(ctx context.Context, id uint64) (*model.GameUserDevices, error) {
	var data *model.GameUserDevices
	cacheKey := c.GetGameUserDevicesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameUserDevicesCache) MultiSet(ctx context.Context, data []*model.GameUserDevices, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameUserDevicesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameUserDevicesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserDevices, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameUserDevicesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameUserDevices)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameUserDevices)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameUserDevicesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameUserDevicesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserDevicesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameUserDevicesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserDevicesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameUserDevicesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
