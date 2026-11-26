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
	gameInvitationWheelRoundsCachePrefixKey = "gameInvitationWheelRounds:"
	// GameInvitationWheelRoundsExpireTime expire time
	GameInvitationWheelRoundsExpireTime = 5 * time.Minute
)

var _ GameInvitationWheelRoundsCache = (*gameInvitationWheelRoundsCache)(nil)

// GameInvitationWheelRoundsCache cache interface
type GameInvitationWheelRoundsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameInvitationWheelRounds, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameInvitationWheelRounds, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameInvitationWheelRounds, error)
	MultiSet(ctx context.Context, data []*model.GameInvitationWheelRounds, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameInvitationWheelRoundsCache define a cache struct
type gameInvitationWheelRoundsCache struct {
	cache cache.Cache
}

// NewGameInvitationWheelRoundsCache new a cache
func NewGameInvitationWheelRoundsCache(cacheType *database.CacheType) GameInvitationWheelRoundsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameInvitationWheelRounds{}
		})
		return &gameInvitationWheelRoundsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameInvitationWheelRounds{}
		})
		return &gameInvitationWheelRoundsCache{cache: c}
	}

	return nil // no cache
}

// GetGameInvitationWheelRoundsCacheKey cache key
func (c *gameInvitationWheelRoundsCache) GetGameInvitationWheelRoundsCacheKey(id uint64) string {
	return gameInvitationWheelRoundsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameInvitationWheelRoundsCache) Set(ctx context.Context, id uint64, data *model.GameInvitationWheelRounds, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameInvitationWheelRoundsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameInvitationWheelRoundsCache) Get(ctx context.Context, id uint64) (*model.GameInvitationWheelRounds, error) {
	var data *model.GameInvitationWheelRounds
	cacheKey := c.GetGameInvitationWheelRoundsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameInvitationWheelRoundsCache) MultiSet(ctx context.Context, data []*model.GameInvitationWheelRounds, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameInvitationWheelRoundsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameInvitationWheelRoundsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameInvitationWheelRounds, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameInvitationWheelRoundsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameInvitationWheelRounds)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameInvitationWheelRounds)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameInvitationWheelRoundsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameInvitationWheelRoundsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameInvitationWheelRoundsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameInvitationWheelRoundsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameInvitationWheelRoundsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameInvitationWheelRoundsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
