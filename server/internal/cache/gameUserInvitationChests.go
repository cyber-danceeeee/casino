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
	gameUserInvitationChestsCachePrefixKey = "gameUserInvitationChests:"
	// GameUserInvitationChestsExpireTime expire time
	GameUserInvitationChestsExpireTime = 5 * time.Minute
)

var _ GameUserInvitationChestsCache = (*gameUserInvitationChestsCache)(nil)

// GameUserInvitationChestsCache cache interface
type GameUserInvitationChestsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameUserInvitationChests, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameUserInvitationChests, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserInvitationChests, error)
	MultiSet(ctx context.Context, data []*model.GameUserInvitationChests, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameUserInvitationChestsCache define a cache struct
type gameUserInvitationChestsCache struct {
	cache cache.Cache
}

// NewGameUserInvitationChestsCache new a cache
func NewGameUserInvitationChestsCache(cacheType *database.CacheType) GameUserInvitationChestsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserInvitationChests{}
		})
		return &gameUserInvitationChestsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserInvitationChests{}
		})
		return &gameUserInvitationChestsCache{cache: c}
	}

	return nil // no cache
}

// GetGameUserInvitationChestsCacheKey cache key
func (c *gameUserInvitationChestsCache) GetGameUserInvitationChestsCacheKey(id uint64) string {
	return gameUserInvitationChestsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameUserInvitationChestsCache) Set(ctx context.Context, id uint64, data *model.GameUserInvitationChests, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameUserInvitationChestsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameUserInvitationChestsCache) Get(ctx context.Context, id uint64) (*model.GameUserInvitationChests, error) {
	var data *model.GameUserInvitationChests
	cacheKey := c.GetGameUserInvitationChestsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameUserInvitationChestsCache) MultiSet(ctx context.Context, data []*model.GameUserInvitationChests, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameUserInvitationChestsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameUserInvitationChestsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserInvitationChests, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameUserInvitationChestsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameUserInvitationChests)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameUserInvitationChests)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameUserInvitationChestsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameUserInvitationChestsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserInvitationChestsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameUserInvitationChestsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserInvitationChestsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameUserInvitationChestsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
