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
	gameInvitesCachePrefixKey = "gameInvites:"
	// GameInvitesExpireTime expire time
	GameInvitesExpireTime = 5 * time.Minute
)

var _ GameInvitesCache = (*gameInvitesCache)(nil)

// GameInvitesCache cache interface
type GameInvitesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameInvites, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameInvites, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameInvites, error)
	MultiSet(ctx context.Context, data []*model.GameInvites, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameInvitesCache define a cache struct
type gameInvitesCache struct {
	cache cache.Cache
}

// NewGameInvitesCache new a cache
func NewGameInvitesCache(cacheType *database.CacheType) GameInvitesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameInvites{}
		})
		return &gameInvitesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameInvites{}
		})
		return &gameInvitesCache{cache: c}
	}

	return nil // no cache
}

// GetGameInvitesCacheKey cache key
func (c *gameInvitesCache) GetGameInvitesCacheKey(id uint64) string {
	return gameInvitesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameInvitesCache) Set(ctx context.Context, id uint64, data *model.GameInvites, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameInvitesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameInvitesCache) Get(ctx context.Context, id uint64) (*model.GameInvites, error) {
	var data *model.GameInvites
	cacheKey := c.GetGameInvitesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameInvitesCache) MultiSet(ctx context.Context, data []*model.GameInvites, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameInvitesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameInvitesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameInvites, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameInvitesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameInvites)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameInvites)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameInvitesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameInvitesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameInvitesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameInvitesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameInvitesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameInvitesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
