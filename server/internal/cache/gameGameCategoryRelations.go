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
	gameGameCategoryRelationsCachePrefixKey = "gameGameCategoryRelations:"
	// GameGameCategoryRelationsExpireTime expire time
	GameGameCategoryRelationsExpireTime = 5 * time.Minute
)

var _ GameGameCategoryRelationsCache = (*gameGameCategoryRelationsCache)(nil)

// GameGameCategoryRelationsCache cache interface
type GameGameCategoryRelationsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameGameCategoryRelations, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameGameCategoryRelations, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameGameCategoryRelations, error)
	MultiSet(ctx context.Context, data []*model.GameGameCategoryRelations, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameGameCategoryRelationsCache define a cache struct
type gameGameCategoryRelationsCache struct {
	cache cache.Cache
}

// NewGameGameCategoryRelationsCache new a cache
func NewGameGameCategoryRelationsCache(cacheType *database.CacheType) GameGameCategoryRelationsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameGameCategoryRelations{}
		})
		return &gameGameCategoryRelationsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameGameCategoryRelations{}
		})
		return &gameGameCategoryRelationsCache{cache: c}
	}

	return nil // no cache
}

// GetGameGameCategoryRelationsCacheKey cache key
func (c *gameGameCategoryRelationsCache) GetGameGameCategoryRelationsCacheKey(id uint64) string {
	return gameGameCategoryRelationsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameGameCategoryRelationsCache) Set(ctx context.Context, id uint64, data *model.GameGameCategoryRelations, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameGameCategoryRelationsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameGameCategoryRelationsCache) Get(ctx context.Context, id uint64) (*model.GameGameCategoryRelations, error) {
	var data *model.GameGameCategoryRelations
	cacheKey := c.GetGameGameCategoryRelationsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameGameCategoryRelationsCache) MultiSet(ctx context.Context, data []*model.GameGameCategoryRelations, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameGameCategoryRelationsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameGameCategoryRelationsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameGameCategoryRelations, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameGameCategoryRelationsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameGameCategoryRelations)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameGameCategoryRelations)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameGameCategoryRelationsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameGameCategoryRelationsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameGameCategoryRelationsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameGameCategoryRelationsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameGameCategoryRelationsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameGameCategoryRelationsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
