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
	gameInvitationWheelPrizesCachePrefixKey = "gameInvitationWheelPrizes:"
	// GameInvitationWheelPrizesExpireTime expire time
	GameInvitationWheelPrizesExpireTime = 5 * time.Minute
)

var _ GameInvitationWheelPrizesCache = (*gameInvitationWheelPrizesCache)(nil)

// GameInvitationWheelPrizesCache cache interface
type GameInvitationWheelPrizesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameInvitationWheelPrizes, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameInvitationWheelPrizes, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameInvitationWheelPrizes, error)
	MultiSet(ctx context.Context, data []*model.GameInvitationWheelPrizes, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameInvitationWheelPrizesCache define a cache struct
type gameInvitationWheelPrizesCache struct {
	cache cache.Cache
}

// NewGameInvitationWheelPrizesCache new a cache
func NewGameInvitationWheelPrizesCache(cacheType *database.CacheType) GameInvitationWheelPrizesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameInvitationWheelPrizes{}
		})
		return &gameInvitationWheelPrizesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameInvitationWheelPrizes{}
		})
		return &gameInvitationWheelPrizesCache{cache: c}
	}

	return nil // no cache
}

// GetGameInvitationWheelPrizesCacheKey cache key
func (c *gameInvitationWheelPrizesCache) GetGameInvitationWheelPrizesCacheKey(id uint64) string {
	return gameInvitationWheelPrizesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameInvitationWheelPrizesCache) Set(ctx context.Context, id uint64, data *model.GameInvitationWheelPrizes, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameInvitationWheelPrizesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameInvitationWheelPrizesCache) Get(ctx context.Context, id uint64) (*model.GameInvitationWheelPrizes, error) {
	var data *model.GameInvitationWheelPrizes
	cacheKey := c.GetGameInvitationWheelPrizesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameInvitationWheelPrizesCache) MultiSet(ctx context.Context, data []*model.GameInvitationWheelPrizes, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameInvitationWheelPrizesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameInvitationWheelPrizesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameInvitationWheelPrizes, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameInvitationWheelPrizesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameInvitationWheelPrizes)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameInvitationWheelPrizes)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameInvitationWheelPrizesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameInvitationWheelPrizesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameInvitationWheelPrizesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameInvitationWheelPrizesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameInvitationWheelPrizesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameInvitationWheelPrizesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
