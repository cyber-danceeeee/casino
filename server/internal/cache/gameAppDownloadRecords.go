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
	gameAppDownloadRecordsCachePrefixKey = "gameAppDownloadRecords:"
	// GameAppDownloadRecordsExpireTime expire time
	GameAppDownloadRecordsExpireTime = 5 * time.Minute
)

var _ GameAppDownloadRecordsCache = (*gameAppDownloadRecordsCache)(nil)

// GameAppDownloadRecordsCache cache interface
type GameAppDownloadRecordsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameAppDownloadRecords, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameAppDownloadRecords, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameAppDownloadRecords, error)
	MultiSet(ctx context.Context, data []*model.GameAppDownloadRecords, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameAppDownloadRecordsCache define a cache struct
type gameAppDownloadRecordsCache struct {
	cache cache.Cache
}

// NewGameAppDownloadRecordsCache new a cache
func NewGameAppDownloadRecordsCache(cacheType *database.CacheType) GameAppDownloadRecordsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameAppDownloadRecords{}
		})
		return &gameAppDownloadRecordsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameAppDownloadRecords{}
		})
		return &gameAppDownloadRecordsCache{cache: c}
	}

	return nil // no cache
}

// GetGameAppDownloadRecordsCacheKey cache key
func (c *gameAppDownloadRecordsCache) GetGameAppDownloadRecordsCacheKey(id uint64) string {
	return gameAppDownloadRecordsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameAppDownloadRecordsCache) Set(ctx context.Context, id uint64, data *model.GameAppDownloadRecords, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameAppDownloadRecordsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameAppDownloadRecordsCache) Get(ctx context.Context, id uint64) (*model.GameAppDownloadRecords, error) {
	var data *model.GameAppDownloadRecords
	cacheKey := c.GetGameAppDownloadRecordsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameAppDownloadRecordsCache) MultiSet(ctx context.Context, data []*model.GameAppDownloadRecords, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameAppDownloadRecordsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameAppDownloadRecordsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameAppDownloadRecords, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameAppDownloadRecordsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameAppDownloadRecords)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameAppDownloadRecords)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameAppDownloadRecordsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameAppDownloadRecordsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameAppDownloadRecordsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameAppDownloadRecordsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameAppDownloadRecordsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameAppDownloadRecordsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
