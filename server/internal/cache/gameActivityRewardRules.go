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
	gameActivityRewardRulesCachePrefixKey = "gameActivityRewardRules:"
	// GameActivityRewardRulesExpireTime expire time
	GameActivityRewardRulesExpireTime = 5 * time.Minute
)

var _ GameActivityRewardRulesCache = (*gameActivityRewardRulesCache)(nil)

// GameActivityRewardRulesCache cache interface
type GameActivityRewardRulesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameActivityRewardRules, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameActivityRewardRules, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameActivityRewardRules, error)
	MultiSet(ctx context.Context, data []*model.GameActivityRewardRules, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameActivityRewardRulesCache define a cache struct
type gameActivityRewardRulesCache struct {
	cache cache.Cache
}

// NewGameActivityRewardRulesCache new a cache
func NewGameActivityRewardRulesCache(cacheType *database.CacheType) GameActivityRewardRulesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameActivityRewardRules{}
		})
		return &gameActivityRewardRulesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameActivityRewardRules{}
		})
		return &gameActivityRewardRulesCache{cache: c}
	}

	return nil // no cache
}

// GetGameActivityRewardRulesCacheKey cache key
func (c *gameActivityRewardRulesCache) GetGameActivityRewardRulesCacheKey(id uint64) string {
	return gameActivityRewardRulesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameActivityRewardRulesCache) Set(ctx context.Context, id uint64, data *model.GameActivityRewardRules, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameActivityRewardRulesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameActivityRewardRulesCache) Get(ctx context.Context, id uint64) (*model.GameActivityRewardRules, error) {
	var data *model.GameActivityRewardRules
	cacheKey := c.GetGameActivityRewardRulesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameActivityRewardRulesCache) MultiSet(ctx context.Context, data []*model.GameActivityRewardRules, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameActivityRewardRulesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameActivityRewardRulesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameActivityRewardRules, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameActivityRewardRulesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameActivityRewardRules)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameActivityRewardRules)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameActivityRewardRulesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameActivityRewardRulesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameActivityRewardRulesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameActivityRewardRulesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameActivityRewardRulesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameActivityRewardRulesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
