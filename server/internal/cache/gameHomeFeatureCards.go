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
	gameHomeFeatureCardsCachePrefixKey = "gameHomeFeatureCards:"
	// GameHomeFeatureCardsExpireTime expire time
	GameHomeFeatureCardsExpireTime = 5 * time.Minute
)

var _ GameHomeFeatureCardsCache = (*gameHomeFeatureCardsCache)(nil)

// GameHomeFeatureCardsCache cache interface
type GameHomeFeatureCardsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameHomeFeatureCards, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameHomeFeatureCards, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameHomeFeatureCards, error)
	MultiSet(ctx context.Context, data []*model.GameHomeFeatureCards, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameHomeFeatureCardsCache define a cache struct
type gameHomeFeatureCardsCache struct {
	cache cache.Cache
}

// NewGameHomeFeatureCardsCache new a cache
func NewGameHomeFeatureCardsCache(cacheType *database.CacheType) GameHomeFeatureCardsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameHomeFeatureCards{}
		})
		return &gameHomeFeatureCardsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameHomeFeatureCards{}
		})
		return &gameHomeFeatureCardsCache{cache: c}
	}

	return nil // no cache
}

// GetGameHomeFeatureCardsCacheKey cache key
func (c *gameHomeFeatureCardsCache) GetGameHomeFeatureCardsCacheKey(id uint64) string {
	return gameHomeFeatureCardsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameHomeFeatureCardsCache) Set(ctx context.Context, id uint64, data *model.GameHomeFeatureCards, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameHomeFeatureCardsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameHomeFeatureCardsCache) Get(ctx context.Context, id uint64) (*model.GameHomeFeatureCards, error) {
	var data *model.GameHomeFeatureCards
	cacheKey := c.GetGameHomeFeatureCardsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameHomeFeatureCardsCache) MultiSet(ctx context.Context, data []*model.GameHomeFeatureCards, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameHomeFeatureCardsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameHomeFeatureCardsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameHomeFeatureCards, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameHomeFeatureCardsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameHomeFeatureCards)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameHomeFeatureCards)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameHomeFeatureCardsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameHomeFeatureCardsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameHomeFeatureCardsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameHomeFeatureCardsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameHomeFeatureCardsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameHomeFeatureCardsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
