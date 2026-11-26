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
	gameGamesCachePrefixKey = "gameGames:"
	// GameGamesExpireTime expire time
	GameGamesExpireTime = 5 * time.Minute
)

var _ GameGamesCache = (*gameGamesCache)(nil)

// GameGamesCache cache interface
type GameGamesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameGames, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameGames, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameGames, error)
	MultiSet(ctx context.Context, data []*model.GameGames, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameGamesCache define a cache struct
type gameGamesCache struct {
	cache cache.Cache
}

// NewGameGamesCache new a cache
func NewGameGamesCache(cacheType *database.CacheType) GameGamesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameGames{}
		})
		return &gameGamesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameGames{}
		})
		return &gameGamesCache{cache: c}
	}

	return nil // no cache
}

// GetGameGamesCacheKey cache key
func (c *gameGamesCache) GetGameGamesCacheKey(id uint64) string {
	return gameGamesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameGamesCache) Set(ctx context.Context, id uint64, data *model.GameGames, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameGamesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameGamesCache) Get(ctx context.Context, id uint64) (*model.GameGames, error) {
	var data *model.GameGames
	cacheKey := c.GetGameGamesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameGamesCache) MultiSet(ctx context.Context, data []*model.GameGames, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameGamesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameGamesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameGames, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameGamesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameGames)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameGames)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameGamesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameGamesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameGamesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameGamesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameGamesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameGamesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
