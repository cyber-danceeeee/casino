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
	gameInvitationChestConfigsCachePrefixKey = "gameInvitationChestConfigs:"
	// GameInvitationChestConfigsExpireTime expire time
	GameInvitationChestConfigsExpireTime = 5 * time.Minute
)

var _ GameInvitationChestConfigsCache = (*gameInvitationChestConfigsCache)(nil)

// GameInvitationChestConfigsCache cache interface
type GameInvitationChestConfigsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameInvitationChestConfigs, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameInvitationChestConfigs, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameInvitationChestConfigs, error)
	MultiSet(ctx context.Context, data []*model.GameInvitationChestConfigs, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameInvitationChestConfigsCache define a cache struct
type gameInvitationChestConfigsCache struct {
	cache cache.Cache
}

// NewGameInvitationChestConfigsCache new a cache
func NewGameInvitationChestConfigsCache(cacheType *database.CacheType) GameInvitationChestConfigsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameInvitationChestConfigs{}
		})
		return &gameInvitationChestConfigsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameInvitationChestConfigs{}
		})
		return &gameInvitationChestConfigsCache{cache: c}
	}

	return nil // no cache
}

// GetGameInvitationChestConfigsCacheKey cache key
func (c *gameInvitationChestConfigsCache) GetGameInvitationChestConfigsCacheKey(id uint64) string {
	return gameInvitationChestConfigsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameInvitationChestConfigsCache) Set(ctx context.Context, id uint64, data *model.GameInvitationChestConfigs, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameInvitationChestConfigsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameInvitationChestConfigsCache) Get(ctx context.Context, id uint64) (*model.GameInvitationChestConfigs, error) {
	var data *model.GameInvitationChestConfigs
	cacheKey := c.GetGameInvitationChestConfigsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameInvitationChestConfigsCache) MultiSet(ctx context.Context, data []*model.GameInvitationChestConfigs, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameInvitationChestConfigsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameInvitationChestConfigsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameInvitationChestConfigs, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameInvitationChestConfigsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameInvitationChestConfigs)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameInvitationChestConfigs)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameInvitationChestConfigsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameInvitationChestConfigsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameInvitationChestConfigsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameInvitationChestConfigsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameInvitationChestConfigsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameInvitationChestConfigsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
