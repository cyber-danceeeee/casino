package hotranking

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-dev-frame/sponge/pkg/goredis"

	"casino/internal/content"
)

const keyPrefix = "casino:"

// Store keeps the rolling HOT ranking in a H5 content-scoped Redis sorted set.
// The game ID is the sorted-set member and the score is the number of successful
// game launches during the current business day.
type Store struct {
	client *goredis.Client
}

func New(client *goredis.Client) *Store {
	return &Store{client: client}
}

func (s *Store) RecordLaunch(ctx context.Context, scope content.Scope, gameID uint64) error {
	if s == nil || s.client == nil || !scope.SiteScoped() || scope.SiteID == 0 || gameID == 0 {
		return nil
	}

	now := time.Now()
	key := Key(scope, now)
	if err := s.client.ZIncrBy(ctx, key, 1, strconv.FormatUint(gameID, 10)).Err(); err != nil {
		return err
	}

	// Keep the current day's ranking until the next day plus a small recovery
	// window so workers can still read it around the day boundary.
	expiresIn := time.Until(nextDay(now)).Round(time.Second) + time.Hour
	if expiresIn <= 0 {
		expiresIn = 25 * time.Hour
	}
	return s.client.Expire(ctx, key, expiresIn).Err()
}

func (s *Store) TopGameIDs(ctx context.Context, scope content.Scope, now time.Time, limit int) ([]uint64, error) {
	if s == nil || s.client == nil || !scope.SiteScoped() || scope.SiteID == 0 || limit <= 0 {
		return nil, nil
	}
	values, err := s.client.ZRevRange(ctx, Key(scope, now), 0, int64(limit)-1).Result()
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(values))
	for _, value := range values {
		id, parseErr := strconv.ParseUint(value, 10, 64)
		if parseErr == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func Key(scope content.Scope, now time.Time) string {
	return fmt.Sprintf("%ssite:%d:site:%d:hot:games:%s", keyPrefix, scope.SiteID, scope.SiteID, now.Format("20060102"))
}

func nextDay(now time.Time) time.Time {
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return start.Add(24 * time.Hour)
}
