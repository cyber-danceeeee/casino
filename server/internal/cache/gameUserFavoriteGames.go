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
	gameUserFavoriteGamesCachePrefixKey = "gameUserFavoriteGames:"
	// GameUserFavoriteGamesExpireTime expire time
	GameUserFavoriteGamesExpireTime = 5 * time.Minute
)

var _ GameUserFavoriteGamesCache = (*gameUserFavoriteGamesCache)(nil)

// GameUserFavoriteGamesCache cache interface
type GameUserFavoriteGamesCache interface {
	Set(ctx context.Context, id uint64, data *model.GameUserFavoriteGames, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.GameUserFavoriteGames, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserFavoriteGames, error)
	MultiSet(ctx context.Context, data []*model.GameUserFavoriteGames, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// gameUserFavoriteGamesCache define a cache struct
type gameUserFavoriteGamesCache struct {
	cache cache.Cache
}

// NewGameUserFavoriteGamesCache new a cache
func NewGameUserFavoriteGamesCache(cacheType *database.CacheType) GameUserFavoriteGamesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserFavoriteGames{}
		})
		return &gameUserFavoriteGamesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.GameUserFavoriteGames{}
		})
		return &gameUserFavoriteGamesCache{cache: c}
	}

	return nil // no cache
}

// GetGameUserFavoriteGamesCacheKey cache key
func (c *gameUserFavoriteGamesCache) GetGameUserFavoriteGamesCacheKey(id uint64) string {
	return gameUserFavoriteGamesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *gameUserFavoriteGamesCache) Set(ctx context.Context, id uint64, data *model.GameUserFavoriteGames, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetGameUserFavoriteGamesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *gameUserFavoriteGamesCache) Get(ctx context.Context, id uint64) (*model.GameUserFavoriteGames, error) {
	var data *model.GameUserFavoriteGames
	cacheKey := c.GetGameUserFavoriteGamesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *gameUserFavoriteGamesCache) MultiSet(ctx context.Context, data []*model.GameUserFavoriteGames, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetGameUserFavoriteGamesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *gameUserFavoriteGamesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.GameUserFavoriteGames, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetGameUserFavoriteGamesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.GameUserFavoriteGames)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.GameUserFavoriteGames)
	for _, id := range ids {
		val, ok := itemMap[c.GetGameUserFavoriteGamesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *gameUserFavoriteGamesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserFavoriteGamesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *gameUserFavoriteGamesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetGameUserFavoriteGamesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *gameUserFavoriteGamesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
