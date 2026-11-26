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
	gameSmsVerificationCodesCachePrefixKey = "gameSmsVerificationCodes:"
	// GameSmsVerificationCodesExpireTime expire time
	GameSmsVerificationCodesExpireTime = 5 * time.Minute
)

var _ GameSmsVerificationCodesCache = (*gameSmsVerificationCodesCache)(nil)

// GameSmsVerificationCodesCache cache interface
type GameSmsVerificationCodesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameSmsVerificationCodes, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameSmsVerificationCodes, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameSmsVerificationCodes, error)
	MultiSet(ctx context.Context, data []*model.GameSmsVerificationCodes, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameSmsVerificationCodesCache define a cache struct
type gameSmsVerificationCodesCache struct {
	cache cache.Cache
}

// NewGameSmsVerificationCodesCache new a cache
func NewGameSmsVerificationCodesCache(cacheType *database.CacheType) GameSmsVerificationCodesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameSmsVerificationCodes{}
		})
		return &gameSmsVerificationCodesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameSmsVerificationCodes{}
		})
		return &gameSmsVerificationCodesCache{cache: c}
	}

	return nil // no cache
}

// GetGameSmsVerificationCodesCacheKey cache key
func (c *gameSmsVerificationCodesCache) GetGameSmsVerificationCodesCacheKey(id uint64) string {
	return gameSmsVerificationCodesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameSmsVerificationCodesCache) Set(ctx context.Context, id uint64, data *model.GameSmsVerificationCodes, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameSmsVerificationCodesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameSmsVerificationCodesCache) Get(ctx context.Context, id uint64) (*model.GameSmsVerificationCodes, error) {
	var data *model.GameSmsVerificationCodes
	cacheKey := c.GetGameSmsVerificationCodesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameSmsVerificationCodesCache) MultiSet(ctx context.Context, data []*model.GameSmsVerificationCodes, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameSmsVerificationCodesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameSmsVerificationCodesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameSmsVerificationCodes, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameSmsVerificationCodesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameSmsVerificationCodes)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameSmsVerificationCodes)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameSmsVerificationCodesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameSmsVerificationCodesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameSmsVerificationCodesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameSmsVerificationCodesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameSmsVerificationCodesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameSmsVerificationCodesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
