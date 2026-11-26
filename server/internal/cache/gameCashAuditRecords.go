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
	gameCashAuditRecordsCachePrefixKey = "gameCashAuditRecords:"
	// GameCashAuditRecordsExpireTime expire time
	GameCashAuditRecordsExpireTime = 5 * time.Minute
)

var _ GameCashAuditRecordsCache = (*gameCashAuditRecordsCache)(nil)

// GameCashAuditRecordsCache cache interface
type GameCashAuditRecordsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameCashAuditRecords, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameCashAuditRecords, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameCashAuditRecords, error)
	MultiSet(ctx context.Context, data []*model.GameCashAuditRecords, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameCashAuditRecordsCache define a cache struct
type gameCashAuditRecordsCache struct {
	cache cache.Cache
}

// NewGameCashAuditRecordsCache new a cache
func NewGameCashAuditRecordsCache(cacheType *database.CacheType) GameCashAuditRecordsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameCashAuditRecords{}
		})
		return &gameCashAuditRecordsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameCashAuditRecords{}
		})
		return &gameCashAuditRecordsCache{cache: c}
	}

	return nil // no cache
}

// GetGameCashAuditRecordsCacheKey cache key
func (c *gameCashAuditRecordsCache) GetGameCashAuditRecordsCacheKey(id uint64) string {
	return gameCashAuditRecordsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameCashAuditRecordsCache) Set(ctx context.Context, id uint64, data *model.GameCashAuditRecords, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameCashAuditRecordsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameCashAuditRecordsCache) Get(ctx context.Context, id uint64) (*model.GameCashAuditRecords, error) {
	var data *model.GameCashAuditRecords
	cacheKey := c.GetGameCashAuditRecordsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameCashAuditRecordsCache) MultiSet(ctx context.Context, data []*model.GameCashAuditRecords, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameCashAuditRecordsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameCashAuditRecordsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameCashAuditRecords, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameCashAuditRecordsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameCashAuditRecords)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameCashAuditRecords)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameCashAuditRecordsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameCashAuditRecordsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameCashAuditRecordsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameCashAuditRecordsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameCashAuditRecordsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameCashAuditRecordsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
