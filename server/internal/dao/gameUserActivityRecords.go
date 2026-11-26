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

var _ GameUserActivityRecordsDao = (*gameUserActivityRecordsDao)(nil)

// GameUserActivityRecordsDao defining the dao interface
type GameUserActivityRecordsDao interface {
	Create(ctx context.Context, table *model.GameUserActivityRecords) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameUserActivityRecords) error
	GetByID(ctx context.Context, id uint64) (*model.GameUserActivityRecords, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUserActivityRecords, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserActivityRecords) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserActivityRecords) error
}

type gameUserActivityRecordsDao struct {
	db    *gorm.DB
	cache cache.GameUserActivityRecordsCache // if nil, the cache is not used.
	sfg   *singleflight.Group                // if cache is nil, the sfg is not used.
}

// NewGameUserActivityRecordsDao creating the dao interface
func NewGameUserActivityRecordsDao(db *gorm.DB, xCache cache.GameUserActivityRecordsCache) GameUserActivityRecordsDao {
	if xCache == nil {
		return &gameUserActivityRecordsDao{db: db}
	}
	return &gameUserActivityRecordsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameUserActivityRecordsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameUserActivityRecords, insert the record and the id value is written back to the table
func (d *gameUserActivityRecordsDao) Create(ctx context.Context, table *model.GameUserActivityRecords) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameUserActivityRecords by id
func (d *gameUserActivityRecordsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUserActivityRecords{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameUserActivityRecords by id, support partial update
func (d *gameUserActivityRecordsDao) UpdateByID(ctx context.Context, table *model.GameUserActivityRecords) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameUserActivityRecordsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameUserActivityRecords) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.ActivityID != 0 {
		update["activity_id"] = table.ActivityID
	}
	if table.ActionType != "" {
		update["action_type"] = table.ActionType
	}
	if table.OrderNo != "" {
		update["order_no"] = table.OrderNo
	}
	if table.RewardCash != nil && table.RewardCash.IsZero() == false {
		update["reward_cash"] = table.RewardCash
	}
	if table.RewardBonus != nil && table.RewardBonus.IsZero() == false {
		update["reward_bonus"] = table.RewardBonus
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.ClaimedAt != nil && table.ClaimedAt.IsZero() == false {
		update["claimed_at"] = table.ClaimedAt
	}
	if table.Remark != "" {
		update["remark"] = table.Remark
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameUserActivityRecords by id
func (d *gameUserActivityRecordsDao) GetByID(ctx context.Context, id uint64) (*model.GameUserActivityRecords, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameUserActivityRecords{}
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
			table := &model.GameUserActivityRecords{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameUserActivityRecordsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameUserActivityRecords)
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

// GetByColumns get a paginated list of gameUserActivityRecordss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameUserActivityRecordsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUserActivityRecords, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameUserActivityRecordsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameUserActivityRecords{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameUserActivityRecords{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameUserActivityRecordsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserActivityRecords) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameUserActivityRecordsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUserActivityRecords{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameUserActivityRecordsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserActivityRecords) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
