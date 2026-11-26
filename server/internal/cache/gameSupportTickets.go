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
	gameSupportTicketsCachePrefixKey = "gameSupportTickets:"
	// GameSupportTicketsExpireTime expire time
	GameSupportTicketsExpireTime = 5 * time.Minute
)

var _ GameSupportTicketsCache = (*gameSupportTicketsCache)(nil)

// GameSupportTicketsCache cache interface
type GameSupportTicketsCache interface {
	Set(ctx context.Context, id uint64, data *model.GameSupportTickets, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameSupportTickets, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameSupportTickets, error)
	MultiSet(ctx context.Context, data []*model.GameSupportTickets, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameSupportTicketsCache define a cache struct
type gameSupportTicketsCache struct {
	cache cache.Cache
}

// NewGameSupportTicketsCache new a cache
func NewGameSupportTicketsCache(cacheType *database.CacheType) GameSupportTicketsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameSupportTickets{}
		})
		return &gameSupportTicketsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameSupportTickets{}
		})
		return &gameSupportTicketsCache{cache: c}
	}

	return nil // no cache
}

// GetGameSupportTicketsCacheKey cache key
func (c *gameSupportTicketsCache) GetGameSupportTicketsCacheKey(id uint64) string {
	return gameSupportTicketsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameSupportTicketsCache) Set(ctx context.Context, id uint64, data *model.GameSupportTickets, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameSupportTicketsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameSupportTicketsCache) Get(ctx context.Context, id uint64) (*model.GameSupportTickets, error) {
	var data *model.GameSupportTickets
	cacheKey := c.GetGameSupportTicketsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameSupportTicketsCache) MultiSet(ctx context.Context, data []*model.GameSupportTickets, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameSupportTicketsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameSupportTicketsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameSupportTickets, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameSupportTicketsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameSupportTickets)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameSupportTickets)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameSupportTicketsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameSupportTicketsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameSupportTicketsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameSupportTicketsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameSupportTicketsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameSupportTicketsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
