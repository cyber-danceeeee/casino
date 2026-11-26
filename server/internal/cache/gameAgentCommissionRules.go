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
	gameAgentCommissionRulesCachePrefixKey = "gameAgentCommissionRules:"
	// GameAgentCommissionRulesExpireTime expire time
	GameAgentCommissionRulesExpireTime = 5 * time.Minute
)

var _ GameAgentCommissionRulesCache = (*gameAgentCommissionRulesCache)(nil)

// GameAgentCommissionRulesCache cache interface
type GameAgentCommissionRulesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameAgentCommissionRules, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameAgentCommissionRules, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameAgentCommissionRules, error)
	MultiSet(ctx context.Context, data []*model.GameAgentCommissionRules, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameAgentCommissionRulesCache define a cache struct
type gameAgentCommissionRulesCache struct {
	cache cache.Cache
}

// NewGameAgentCommissionRulesCache new a cache
func NewGameAgentCommissionRulesCache(cacheType *database.CacheType) GameAgentCommissionRulesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameAgentCommissionRules{}
		})
		return &gameAgentCommissionRulesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameAgentCommissionRules{}
		})
		return &gameAgentCommissionRulesCache{cache: c}
	}

	return nil // no cache
}

// GetGameAgentCommissionRulesCacheKey cache key
func (c *gameAgentCommissionRulesCache) GetGameAgentCommissionRulesCacheKey(id uint64) string {
	return gameAgentCommissionRulesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameAgentCommissionRulesCache) Set(ctx context.Context, id uint64, data *model.GameAgentCommissionRules, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameAgentCommissionRulesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameAgentCommissionRulesCache) Get(ctx context.Context, id uint64) (*model.GameAgentCommissionRules, error) {
	var data *model.GameAgentCommissionRules
	cacheKey := c.GetGameAgentCommissionRulesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameAgentCommissionRulesCache) MultiSet(ctx context.Context, data []*model.GameAgentCommissionRules, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameAgentCommissionRulesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameAgentCommissionRulesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameAgentCommissionRules, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameAgentCommissionRulesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameAgentCommissionRules)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameAgentCommissionRules)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameAgentCommissionRulesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameAgentCommissionRulesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameAgentCommissionRulesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameAgentCommissionRulesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameAgentCommissionRulesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameAgentCommissionRulesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
