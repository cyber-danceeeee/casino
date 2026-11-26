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
	gameVipRequirementsCachePrefixKey = "gameVipRequirements:"
	// GameVipRequirementsExpireTime expire time
	GameVipRequirementsExpireTime = 5 * time.Minute
)

var _ GameVipRequirementsCache = (*gameVipRequirementsCache)(nil)

// GameVipRequirementsCache cache interface
type GameVipRequirementsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameVipRequirements, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameVipRequirements, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameVipRequirements, error)
	MultiSet(ctx context.Context, data []*model.GameVipRequirements, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameVipRequirementsCache define a cache struct
type gameVipRequirementsCache struct {
	cache cache.Cache
}

// NewGameVipRequirementsCache new a cache
func NewGameVipRequirementsCache(cacheType *database.CacheType) GameVipRequirementsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameVipRequirements{}
		})
		return &gameVipRequirementsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameVipRequirements{}
		})
		return &gameVipRequirementsCache{cache: c}
	}

	return nil // no cache
}

// GetGameVipRequirementsCacheKey cache key
func (c *gameVipRequirementsCache) GetGameVipRequirementsCacheKey(id uint64) string {
	return gameVipRequirementsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameVipRequirementsCache) Set(ctx context.Context, id uint64, data *model.GameVipRequirements, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameVipRequirementsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameVipRequirementsCache) Get(ctx context.Context, id uint64) (*model.GameVipRequirements, error) {
	var data *model.GameVipRequirements
	cacheKey := c.GetGameVipRequirementsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameVipRequirementsCache) MultiSet(ctx context.Context, data []*model.GameVipRequirements, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameVipRequirementsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameVipRequirementsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameVipRequirements, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameVipRequirementsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameVipRequirements)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameVipRequirements)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameVipRequirementsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameVipRequirementsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameVipRequirementsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameVipRequirementsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameVipRequirementsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameVipRequirementsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
