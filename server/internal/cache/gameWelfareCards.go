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
	gameWelfareCardsCachePrefixKey = "gameWelfareCards:"
	// GameWelfareCardsExpireTime expire time
	GameWelfareCardsExpireTime = 5 * time.Minute
)

var _ GameWelfareCardsCache = (*gameWelfareCardsCache)(nil)

// GameWelfareCardsCache cache interface
type GameWelfareCardsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameWelfareCards, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameWelfareCards, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameWelfareCards, error)
	MultiSet(ctx context.Context, data []*model.GameWelfareCards, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameWelfareCardsCache define a cache struct
type gameWelfareCardsCache struct {
	cache cache.Cache
}

// NewGameWelfareCardsCache new a cache
func NewGameWelfareCardsCache(cacheType *database.CacheType) GameWelfareCardsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameWelfareCards{}
		})
		return &gameWelfareCardsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameWelfareCards{}
		})
		return &gameWelfareCardsCache{cache: c}
	}

	return nil // no cache
}

// GetGameWelfareCardsCacheKey cache key
func (c *gameWelfareCardsCache) GetGameWelfareCardsCacheKey(id uint64) string {
	return gameWelfareCardsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameWelfareCardsCache) Set(ctx context.Context, id uint64, data *model.GameWelfareCards, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameWelfareCardsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameWelfareCardsCache) Get(ctx context.Context, id uint64) (*model.GameWelfareCards, error) {
	var data *model.GameWelfareCards
	cacheKey := c.GetGameWelfareCardsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameWelfareCardsCache) MultiSet(ctx context.Context, data []*model.GameWelfareCards, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameWelfareCardsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameWelfareCardsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameWelfareCards, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameWelfareCardsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameWelfareCards)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameWelfareCards)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameWelfareCardsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameWelfareCardsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameWelfareCardsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameWelfareCardsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameWelfareCardsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameWelfareCardsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
