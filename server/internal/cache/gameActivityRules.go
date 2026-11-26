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
	gameActivityRulesCachePrefixKey = "gameActivityRules:"
	// GameActivityRulesExpireTime expire time
	GameActivityRulesExpireTime = 5 * time.Minute
)

var _ GameActivityRulesCache = (*gameActivityRulesCache)(nil)

// GameActivityRulesCache cache interface
type GameActivityRulesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameActivityRules, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameActivityRules, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameActivityRules, error)
	MultiSet(ctx context.Context, data []*model.GameActivityRules, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameActivityRulesCache define a cache struct
type gameActivityRulesCache struct {
	cache cache.Cache
}

// NewGameActivityRulesCache new a cache
func NewGameActivityRulesCache(cacheType *database.CacheType) GameActivityRulesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameActivityRules{}
		})
		return &gameActivityRulesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameActivityRules{}
		})
		return &gameActivityRulesCache{cache: c}
	}

	return nil // no cache
}

// GetGameActivityRulesCacheKey cache key
func (c *gameActivityRulesCache) GetGameActivityRulesCacheKey(id uint64) string {
	return gameActivityRulesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameActivityRulesCache) Set(ctx context.Context, id uint64, data *model.GameActivityRules, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameActivityRulesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameActivityRulesCache) Get(ctx context.Context, id uint64) (*model.GameActivityRules, error) {
	var data *model.GameActivityRules
	cacheKey := c.GetGameActivityRulesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameActivityRulesCache) MultiSet(ctx context.Context, data []*model.GameActivityRules, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameActivityRulesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameActivityRulesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameActivityRules, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameActivityRulesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameActivityRules)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameActivityRules)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameActivityRulesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameActivityRulesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameActivityRulesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameActivityRulesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameActivityRulesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameActivityRulesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
