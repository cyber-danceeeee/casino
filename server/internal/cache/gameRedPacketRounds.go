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
	gameRedPacketRoundsCachePrefixKey = "gameRedPacketRounds:"
	// GameRedPacketRoundsExpireTime expire time
	GameRedPacketRoundsExpireTime = 5 * time.Minute
)

var _ GameRedPacketRoundsCache = (*gameRedPacketRoundsCache)(nil)

// GameRedPacketRoundsCache cache interface
type GameRedPacketRoundsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameRedPacketRounds, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameRedPacketRounds, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameRedPacketRounds, error)
	MultiSet(ctx context.Context, data []*model.GameRedPacketRounds, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameRedPacketRoundsCache define a cache struct
type gameRedPacketRoundsCache struct {
	cache cache.Cache
}

// NewGameRedPacketRoundsCache new a cache
func NewGameRedPacketRoundsCache(cacheType *database.CacheType) GameRedPacketRoundsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameRedPacketRounds{}
		})
		return &gameRedPacketRoundsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameRedPacketRounds{}
		})
		return &gameRedPacketRoundsCache{cache: c}
	}

	return nil // no cache
}

// GetGameRedPacketRoundsCacheKey cache key
func (c *gameRedPacketRoundsCache) GetGameRedPacketRoundsCacheKey(id uint64) string {
	return gameRedPacketRoundsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameRedPacketRoundsCache) Set(ctx context.Context, id uint64, data *model.GameRedPacketRounds, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameRedPacketRoundsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameRedPacketRoundsCache) Get(ctx context.Context, id uint64) (*model.GameRedPacketRounds, error) {
	var data *model.GameRedPacketRounds
	cacheKey := c.GetGameRedPacketRoundsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameRedPacketRoundsCache) MultiSet(ctx context.Context, data []*model.GameRedPacketRounds, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameRedPacketRoundsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameRedPacketRoundsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameRedPacketRounds, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameRedPacketRoundsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameRedPacketRounds)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameRedPacketRounds)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameRedPacketRoundsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameRedPacketRoundsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameRedPacketRoundsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameRedPacketRoundsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameRedPacketRoundsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameRedPacketRoundsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
