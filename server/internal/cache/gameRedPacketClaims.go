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
	gameRedPacketClaimsCachePrefixKey = "gameRedPacketClaims:"
	// GameRedPacketClaimsExpireTime expire time
	GameRedPacketClaimsExpireTime = 5 * time.Minute
)

var _ GameRedPacketClaimsCache = (*gameRedPacketClaimsCache)(nil)

// GameRedPacketClaimsCache cache interface
type GameRedPacketClaimsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameRedPacketClaims, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameRedPacketClaims, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameRedPacketClaims, error)
	MultiSet(ctx context.Context, data []*model.GameRedPacketClaims, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameRedPacketClaimsCache define a cache struct
type gameRedPacketClaimsCache struct {
	cache cache.Cache
}

// NewGameRedPacketClaimsCache new a cache
func NewGameRedPacketClaimsCache(cacheType *database.CacheType) GameRedPacketClaimsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameRedPacketClaims{}
		})
		return &gameRedPacketClaimsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameRedPacketClaims{}
		})
		return &gameRedPacketClaimsCache{cache: c}
	}

	return nil // no cache
}

// GetGameRedPacketClaimsCacheKey cache key
func (c *gameRedPacketClaimsCache) GetGameRedPacketClaimsCacheKey(id uint64) string {
	return gameRedPacketClaimsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameRedPacketClaimsCache) Set(ctx context.Context, id uint64, data *model.GameRedPacketClaims, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameRedPacketClaimsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameRedPacketClaimsCache) Get(ctx context.Context, id uint64) (*model.GameRedPacketClaims, error) {
	var data *model.GameRedPacketClaims
	cacheKey := c.GetGameRedPacketClaimsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameRedPacketClaimsCache) MultiSet(ctx context.Context, data []*model.GameRedPacketClaims, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameRedPacketClaimsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameRedPacketClaimsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameRedPacketClaims, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameRedPacketClaimsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameRedPacketClaims)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameRedPacketClaims)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameRedPacketClaimsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameRedPacketClaimsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameRedPacketClaimsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameRedPacketClaimsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameRedPacketClaimsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameRedPacketClaimsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
