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
	gameUserSettingsCachePrefixKey = "gameUserSettings:"
	// GameUserSettingsExpireTime expire time
	GameUserSettingsExpireTime = 5 * time.Minute
)

var _ GameUserSettingsCache = (*gameUserSettingsCache)(nil)

// GameUserSettingsCache cache interface
type GameUserSettingsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameUserSettings, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameUserSettings, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserSettings, error)
	MultiSet(ctx context.Context, data []*model.GameUserSettings, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameUserSettingsCache define a cache struct
type gameUserSettingsCache struct {
	cache cache.Cache
}

// NewGameUserSettingsCache new a cache
func NewGameUserSettingsCache(cacheType *database.CacheType) GameUserSettingsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserSettings{}
		})
		return &gameUserSettingsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserSettings{}
		})
		return &gameUserSettingsCache{cache: c}
	}

	return nil // no cache
}

// GetGameUserSettingsCacheKey cache key
func (c *gameUserSettingsCache) GetGameUserSettingsCacheKey(id uint64) string {
	return gameUserSettingsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameUserSettingsCache) Set(ctx context.Context, id uint64, data *model.GameUserSettings, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameUserSettingsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameUserSettingsCache) Get(ctx context.Context, id uint64) (*model.GameUserSettings, error) {
	var data *model.GameUserSettings
	cacheKey := c.GetGameUserSettingsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameUserSettingsCache) MultiSet(ctx context.Context, data []*model.GameUserSettings, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameUserSettingsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameUserSettingsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserSettings, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameUserSettingsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameUserSettings)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameUserSettings)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameUserSettingsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameUserSettingsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserSettingsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameUserSettingsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserSettingsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameUserSettingsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
