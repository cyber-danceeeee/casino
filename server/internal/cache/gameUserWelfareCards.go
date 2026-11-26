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
	gameUserWelfareCardsCachePrefixKey = "gameUserWelfareCards:"
	// GameUserWelfareCardsExpireTime expire time
	GameUserWelfareCardsExpireTime = 5 * time.Minute
)

var _ GameUserWelfareCardsCache = (*gameUserWelfareCardsCache)(nil)

// GameUserWelfareCardsCache cache interface
type GameUserWelfareCardsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameUserWelfareCards, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameUserWelfareCards, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserWelfareCards, error)
	MultiSet(ctx context.Context, data []*model.GameUserWelfareCards, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameUserWelfareCardsCache define a cache struct
type gameUserWelfareCardsCache struct {
	cache cache.Cache
}

// NewGameUserWelfareCardsCache new a cache
func NewGameUserWelfareCardsCache(cacheType *database.CacheType) GameUserWelfareCardsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserWelfareCards{}
		})
		return &gameUserWelfareCardsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserWelfareCards{}
		})
		return &gameUserWelfareCardsCache{cache: c}
	}

	return nil // no cache
}

// GetGameUserWelfareCardsCacheKey cache key
func (c *gameUserWelfareCardsCache) GetGameUserWelfareCardsCacheKey(id uint64) string {
	return gameUserWelfareCardsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameUserWelfareCardsCache) Set(ctx context.Context, id uint64, data *model.GameUserWelfareCards, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameUserWelfareCardsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameUserWelfareCardsCache) Get(ctx context.Context, id uint64) (*model.GameUserWelfareCards, error) {
	var data *model.GameUserWelfareCards
	cacheKey := c.GetGameUserWelfareCardsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameUserWelfareCardsCache) MultiSet(ctx context.Context, data []*model.GameUserWelfareCards, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameUserWelfareCardsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameUserWelfareCardsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserWelfareCards, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameUserWelfareCardsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameUserWelfareCards)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameUserWelfareCards)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameUserWelfareCardsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameUserWelfareCardsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserWelfareCardsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameUserWelfareCardsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserWelfareCardsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameUserWelfareCardsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
