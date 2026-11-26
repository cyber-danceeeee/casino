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
	gameBetReportsCachePrefixKey = "gameBetReports:"
	// GameBetReportsExpireTime expire time
	GameBetReportsExpireTime = 5 * time.Minute
)

var _ GameBetReportsCache = (*gameBetReportsCache)(nil)

// GameBetReportsCache cache interface
type GameBetReportsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameBetReports, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameBetReports, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameBetReports, error)
	MultiSet(ctx context.Context, data []*model.GameBetReports, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameBetReportsCache define a cache struct
type gameBetReportsCache struct {
	cache cache.Cache
}

// NewGameBetReportsCache new a cache
func NewGameBetReportsCache(cacheType *database.CacheType) GameBetReportsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameBetReports{}
		})
		return &gameBetReportsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameBetReports{}
		})
		return &gameBetReportsCache{cache: c}
	}

	return nil // no cache
}

// GetGameBetReportsCacheKey cache key
func (c *gameBetReportsCache) GetGameBetReportsCacheKey(id uint64) string {
	return gameBetReportsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameBetReportsCache) Set(ctx context.Context, id uint64, data *model.GameBetReports, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameBetReportsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameBetReportsCache) Get(ctx context.Context, id uint64) (*model.GameBetReports, error) {
	var data *model.GameBetReports
	cacheKey := c.GetGameBetReportsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameBetReportsCache) MultiSet(ctx context.Context, data []*model.GameBetReports, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameBetReportsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameBetReportsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameBetReports, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameBetReportsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameBetReports)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameBetReports)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameBetReportsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameBetReportsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameBetReportsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameBetReportsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameBetReportsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameBetReportsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
