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
	gameBonusRecordsCachePrefixKey = "gameBonusRecords:"
	// GameBonusRecordsExpireTime expire time
	GameBonusRecordsExpireTime = 5 * time.Minute
)

var _ GameBonusRecordsCache = (*gameBonusRecordsCache)(nil)

// GameBonusRecordsCache cache interface
type GameBonusRecordsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameBonusRecords, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameBonusRecords, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameBonusRecords, error)
	MultiSet(ctx context.Context, data []*model.GameBonusRecords, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameBonusRecordsCache define a cache struct
type gameBonusRecordsCache struct {
	cache cache.Cache
}

// NewGameBonusRecordsCache new a cache
func NewGameBonusRecordsCache(cacheType *database.CacheType) GameBonusRecordsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameBonusRecords{}
		})
		return &gameBonusRecordsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameBonusRecords{}
		})
		return &gameBonusRecordsCache{cache: c}
	}

	return nil // no cache
}

// GetGameBonusRecordsCacheKey cache key
func (c *gameBonusRecordsCache) GetGameBonusRecordsCacheKey(id uint64) string {
	return gameBonusRecordsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameBonusRecordsCache) Set(ctx context.Context, id uint64, data *model.GameBonusRecords, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameBonusRecordsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameBonusRecordsCache) Get(ctx context.Context, id uint64) (*model.GameBonusRecords, error) {
	var data *model.GameBonusRecords
	cacheKey := c.GetGameBonusRecordsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameBonusRecordsCache) MultiSet(ctx context.Context, data []*model.GameBonusRecords, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameBonusRecordsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameBonusRecordsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameBonusRecords, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameBonusRecordsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameBonusRecords)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameBonusRecords)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameBonusRecordsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameBonusRecordsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameBonusRecordsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameBonusRecordsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameBonusRecordsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameBonusRecordsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
