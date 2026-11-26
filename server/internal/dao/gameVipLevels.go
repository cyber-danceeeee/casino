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

var _ GameVipLevelsDao = (*gameVipLevelsDao)(nil)

// GameVipLevelsDao defining the dao interface
type GameVipLevelsDao interface {
	Create(ctx context.Context, table *model.GameVipLevels) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameVipLevels) error
	GetByID(ctx context.Context, id uint64) (*model.GameVipLevels, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameVipLevels, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameVipLevels) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameVipLevels) error
}

type gameVipLevelsDao struct {
	db    *gorm.DB
	cache cache.GameVipLevelsCache // if nil, the cache is not used.
	sfg   *singleflight.Group      // if cache is nil, the sfg is not used.
}

// NewGameVipLevelsDao creating the dao interface
func NewGameVipLevelsDao(db *gorm.DB, xCache cache.GameVipLevelsCache) GameVipLevelsDao {
	if xCache == nil {
		return &gameVipLevelsDao{db: db}
	}
	return &gameVipLevelsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameVipLevelsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameVipLevels, insert the record and the id value is written back to the table
func (d *gameVipLevelsDao) Create(ctx context.Context, table *model.GameVipLevels) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameVipLevels by id
func (d *gameVipLevelsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameVipLevels{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameVipLevels by id, support partial update
func (d *gameVipLevelsDao) UpdateByID(ctx context.Context, table *model.GameVipLevels) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameVipLevelsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameVipLevels) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.LevelNo != 0 {
		update["level_no"] = table.LevelNo
	}
	if table.Name != "" {
		update["name"] = table.Name
	}
	if table.IconURL != "" {
		update["icon_url"] = table.IconURL
	}
	if table.LevelUpReward != nil && table.LevelUpReward.IsZero() == false {
		update["level_up_reward"] = table.LevelUpReward
	}
	if table.WeeklyReward != nil && table.WeeklyReward.IsZero() == false {
		update["weekly_reward"] = table.WeeklyReward
	}
	if table.MonthlyReward != nil && table.MonthlyReward.IsZero() == false {
		update["monthly_reward"] = table.MonthlyReward
	}
	if table.DailyWithdrawCount != 0 {
		update["daily_withdraw_count"] = table.DailyWithdrawCount
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameVipLevels by id
func (d *gameVipLevelsDao) GetByID(ctx context.Context, id uint64) (*model.GameVipLevels, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameVipLevels{}
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
			table := &model.GameVipLevels{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameVipLevelsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameVipLevels)
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

// GetByColumns get a paginated list of gameVipLevelss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameVipLevelsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameVipLevels, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameVipLevelsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameVipLevels{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameVipLevels{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameVipLevelsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameVipLevels) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameVipLevelsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameVipLevels{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameVipLevelsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameVipLevels) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
