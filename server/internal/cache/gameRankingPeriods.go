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
	gameRankingPeriodsCachePrefixKey = "gameRankingPeriods:"
	// GameRankingPeriodsExpireTime expire time
	GameRankingPeriodsExpireTime = 5 * time.Minute
)

var _ GameRankingPeriodsCache = (*gameRankingPeriodsCache)(nil)

// GameRankingPeriodsCache cache interface
type GameRankingPeriodsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameRankingPeriods, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameRankingPeriods, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameRankingPeriods, error)
	MultiSet(ctx context.Context, data []*model.GameRankingPeriods, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameRankingPeriodsCache define a cache struct
type gameRankingPeriodsCache struct {
	cache cache.Cache
}

// NewGameRankingPeriodsCache new a cache
func NewGameRankingPeriodsCache(cacheType *database.CacheType) GameRankingPeriodsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameRankingPeriods{}
		})
		return &gameRankingPeriodsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameRankingPeriods{}
		})
		return &gameRankingPeriodsCache{cache: c}
	}

	return nil // no cache
}

// GetGameRankingPeriodsCacheKey cache key
func (c *gameRankingPeriodsCache) GetGameRankingPeriodsCacheKey(id uint64) string {
	return gameRankingPeriodsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameRankingPeriodsCache) Set(ctx context.Context, id uint64, data *model.GameRankingPeriods, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameRankingPeriodsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameRankingPeriodsCache) Get(ctx context.Context, id uint64) (*model.GameRankingPeriods, error) {
	var data *model.GameRankingPeriods
	cacheKey := c.GetGameRankingPeriodsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameRankingPeriodsCache) MultiSet(ctx context.Context, data []*model.GameRankingPeriods, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameRankingPeriodsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameRankingPeriodsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameRankingPeriods, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameRankingPeriodsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameRankingPeriods)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameRankingPeriods)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameRankingPeriodsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameRankingPeriodsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameRankingPeriodsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameRankingPeriodsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameRankingPeriodsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameRankingPeriodsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
