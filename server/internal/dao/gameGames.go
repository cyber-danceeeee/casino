package dao

import (
	"context"
	"errors"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"casino/internal/cache"
	"casino/internal/database"
	"casino/internal/model"
)

var _ GameGamesDao = (*gameGamesDao)(nil)

// GameGamesDao defining the dao interface
type GameGamesDao interface {
	Create(ctx context.Context, table *model.GameGames) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameGames) error
	GetByID(ctx context.Context, id uint64) (*model.GameGames, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameGames, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameGames) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameGames) error
}

type gameGamesDao struct {
	db    *gorm.DB
	cache cache.GameGamesCache // if nil, the cache is not used.
	sfg   *singleflight.Group  // if cache is nil, the sfg is not used.
}

// NewGameGamesDao creating the dao interface
func NewGameGamesDao(db *gorm.DB, xCache cache.GameGamesCache) GameGamesDao {
	if xCache == nil {
		return &gameGamesDao{db: db}
	}
	return &gameGamesDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameGamesDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameGames, insert the record and the id value is written back to the table
func (d *gameGamesDao) Create(ctx context.Context, table *model.GameGames) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameGames by id
func (d *gameGamesDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameGames{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameGames by id, support partial update
func (d *gameGamesDao) UpdateByID(ctx context.Context, table *model.GameGames) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameGamesDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameGames) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.GameCode != "" {
		update["game_code"] = table.GameCode
	}
	if table.ProviderID != 0 {
		update["provider_id"] = table.ProviderID
	}
	if table.ProviderGameCode != "" {
		update["provider_game_code"] = table.ProviderGameCode
	}
	if table.Name != "" {
		update["name"] = table.Name
	}
	if table.GameType != "" {
		update["game_type"] = table.GameType
	}
	if table.CoverURL != "" {
		update["cover_url"] = table.CoverURL
	}
	if table.Badge != "" {
		update["badge"] = table.Badge
	}
	if table.PlayerCount != 0 {
		update["player_count"] = table.PlayerCount
	}
	if table.HotScore != 0 {
		update["hot_score"] = table.HotScore
	}
	if table.LaunchURL != "" {
		update["launch_url"] = table.LaunchURL
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.Sort != 0 {
		update["sort"] = table.Sort
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameGames by id
func (d *gameGamesDao) GetByID(ctx context.Context, id uint64) (*model.GameGames, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameGames{}
		err := d.db.WithContext(ctx).Where("id = ?", id).First(record).Error
		return record, err
	}

	// get from cache
	record, err := d.cache.Get(ctx, id)
	if err == nil {
		return record, nil
	}

	// get from database
	if errors.Is(err, database.ErrCacheNotFound) {
		// for the same id, prevent high concurrent simultaneous access to database
		val, err, _ := d.sfg.Do(utils.Uint64ToStr(id), func() (interface{}, error) { //nolint
			table := &model.GameGames{}
			err = d.db.WithContext(ctx).Where("id = ?", id).First(table).Error
			if err != nil {
				if errors.Is(err, database.ErrRecordNotFound) {
					// set placeholder cache to prevent cache penetration, default expiration time 10 minutes
					if err = d.cache.SetPlaceholder(ctx, id); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("id", id))
					}
					return nil, database.ErrRecordNotFound
				}
				return nil, err
			}
			// set cache
			if err = d.cache.Set(ctx, id, table, cache.GameGamesExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameGames)
		if !ok {
			return nil, database.ErrRecordNotFound
		}
		return table, nil
	}

	if d.cache.IsPlaceholderErr(err) {
		return nil, database.ErrRecordNotFound
	}

	return nil, err
}

// GetByColumns get a paginated list of gameGamess by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameGamesDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameGames, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameGamesColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameGames{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameGames{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameGamesDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameGames) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameGamesDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameGames{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameGamesDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameGames) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
