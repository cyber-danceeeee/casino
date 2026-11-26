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

var _ GameUserVipLogsDao = (*gameUserVipLogsDao)(nil)

// GameUserVipLogsDao defining the dao interface
type GameUserVipLogsDao interface {
	Create(ctx context.Context, table *model.GameUserVipLogs) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameUserVipLogs) error
	GetByID(ctx context.Context, id uint64) (*model.GameUserVipLogs, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUserVipLogs, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserVipLogs) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserVipLogs) error
}

type gameUserVipLogsDao struct {
	db    *gorm.DB
	cache cache.GameUserVipLogsCache // if nil, the cache is not used.
	sfg   *singleflight.Group        // if cache is nil, the sfg is not used.
}

// NewGameUserVipLogsDao creating the dao interface
func NewGameUserVipLogsDao(db *gorm.DB, xCache cache.GameUserVipLogsCache) GameUserVipLogsDao {
	if xCache == nil {
		return &gameUserVipLogsDao{db: db}
	}
	return &gameUserVipLogsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameUserVipLogsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameUserVipLogs, insert the record and the id value is written back to the table
func (d *gameUserVipLogsDao) Create(ctx context.Context, table *model.GameUserVipLogs) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameUserVipLogs by id
func (d *gameUserVipLogsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUserVipLogs{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameUserVipLogs by id, support partial update
func (d *gameUserVipLogsDao) UpdateByID(ctx context.Context, table *model.GameUserVipLogs) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameUserVipLogsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameUserVipLogs) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.OldLevelNo != 0 {
		update["old_level_no"] = table.OldLevelNo
	}
	if table.NewLevelNo != 0 {
		update["new_level_no"] = table.NewLevelNo
	}
	if table.Reason != "" {
		update["reason"] = table.Reason
	}
	if table.DepositAmount != nil && table.DepositAmount.IsZero() == false {
		update["deposit_amount"] = table.DepositAmount
	}
	if table.ValidBetAmount != nil && table.ValidBetAmount.IsZero() == false {
		update["valid_bet_amount"] = table.ValidBetAmount
	}
	if table.RewardAmount != nil && table.RewardAmount.IsZero() == false {
		update["reward_amount"] = table.RewardAmount
	}
	if table.ChangedAt != nil && table.ChangedAt.IsZero() == false {
		update["changed_at"] = table.ChangedAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameUserVipLogs by id
func (d *gameUserVipLogsDao) GetByID(ctx context.Context, id uint64) (*model.GameUserVipLogs, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameUserVipLogs{}
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
			table := &model.GameUserVipLogs{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameUserVipLogsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameUserVipLogs)
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

// GetByColumns get a paginated list of gameUserVipLogss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameUserVipLogsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUserVipLogs, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameUserVipLogsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameUserVipLogs{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameUserVipLogs{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameUserVipLogsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserVipLogs) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameUserVipLogsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUserVipLogs{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameUserVipLogsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserVipLogs) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
