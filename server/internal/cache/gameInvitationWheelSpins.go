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
	gameInvitationWheelSpinsCachePrefixKey = "gameInvitationWheelSpins:"
	// GameInvitationWheelSpinsExpireTime expire time
	GameInvitationWheelSpinsExpireTime = 5 * time.Minute
)

var _ GameInvitationWheelSpinsCache = (*gameInvitationWheelSpinsCache)(nil)

// GameInvitationWheelSpinsCache cache interface
type GameInvitationWheelSpinsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameInvitationWheelSpins, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameInvitationWheelSpins, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameInvitationWheelSpins, error)
	MultiSet(ctx context.Context, data []*model.GameInvitationWheelSpins, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameInvitationWheelSpinsCache define a cache struct
type gameInvitationWheelSpinsCache struct {
	cache cache.Cache
}

// NewGameInvitationWheelSpinsCache new a cache
func NewGameInvitationWheelSpinsCache(cacheType *database.CacheType) GameInvitationWheelSpinsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameInvitationWheelSpins{}
		})
		return &gameInvitationWheelSpinsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameInvitationWheelSpins{}
		})
		return &gameInvitationWheelSpinsCache{cache: c}
	}

	return nil // no cache
}

// GetGameInvitationWheelSpinsCacheKey cache key
func (c *gameInvitationWheelSpinsCache) GetGameInvitationWheelSpinsCacheKey(id uint64) string {
	return gameInvitationWheelSpinsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameInvitationWheelSpinsCache) Set(ctx context.Context, id uint64, data *model.GameInvitationWheelSpins, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameInvitationWheelSpinsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameInvitationWheelSpinsCache) Get(ctx context.Context, id uint64) (*model.GameInvitationWheelSpins, error) {
	var data *model.GameInvitationWheelSpins
	cacheKey := c.GetGameInvitationWheelSpinsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameInvitationWheelSpinsCache) MultiSet(ctx context.Context, data []*model.GameInvitationWheelSpins, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameInvitationWheelSpinsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameInvitationWheelSpinsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameInvitationWheelSpins, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameInvitationWheelSpinsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameInvitationWheelSpins)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameInvitationWheelSpins)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameInvitationWheelSpinsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameInvitationWheelSpinsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameInvitationWheelSpinsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameInvitationWheelSpinsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameInvitationWheelSpinsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameInvitationWheelSpinsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
