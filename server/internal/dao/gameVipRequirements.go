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

var _ GameVipRequirementsDao = (*gameVipRequirementsDao)(nil)

// GameVipRequirementsDao defining the dao interface
type GameVipRequirementsDao interface {
	Create(ctx context.Context, table *model.GameVipRequirements) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameVipRequirements) error
	GetByID(ctx context.Context, id uint64) (*model.GameVipRequirements, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameVipRequirements, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameVipRequirements) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameVipRequirements) error
}

type gameVipRequirementsDao struct {
	db    *gorm.DB
	cache cache.GameVipRequirementsCache // if nil, the cache is not used.
	sfg   *singleflight.Group            // if cache is nil, the sfg is not used.
}

// NewGameVipRequirementsDao creating the dao interface
func NewGameVipRequirementsDao(db *gorm.DB, xCache cache.GameVipRequirementsCache) GameVipRequirementsDao {
	if xCache == nil {
		return &gameVipRequirementsDao{db: db}
	}
	return &gameVipRequirementsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameVipRequirementsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameVipRequirements, insert the record and the id value is written back to the table
func (d *gameVipRequirementsDao) Create(ctx context.Context, table *model.GameVipRequirements) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameVipRequirements by id
func (d *gameVipRequirementsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameVipRequirements{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameVipRequirements by id, support partial update
func (d *gameVipRequirementsDao) UpdateByID(ctx context.Context, table *model.GameVipRequirements) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameVipRequirementsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameVipRequirements) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.VipLevelID != 0 {
		update["vip_level_id"] = table.VipLevelID
	}
	if table.RequirementType != "" {
		update["requirement_type"] = table.RequirementType
	}
	if table.TargetLevelNo != 0 {
		update["target_level_no"] = table.TargetLevelNo
	}
	if table.PeriodDays != 0 {
		update["period_days"] = table.PeriodDays
	}
	if table.DepositAmount != nil && table.DepositAmount.IsZero() == false {
		update["deposit_amount"] = table.DepositAmount
	}
	if table.ValidBetAmount != nil && table.ValidBetAmount.IsZero() == false {
		update["valid_bet_amount"] = table.ValidBetAmount
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameVipRequirements by id
func (d *gameVipRequirementsDao) GetByID(ctx context.Context, id uint64) (*model.GameVipRequirements, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameVipRequirements{}
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
			table := &model.GameVipRequirements{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameVipRequirementsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameVipRequirements)
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

// GetByColumns get a paginated list of gameVipRequirementss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameVipRequirementsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameVipRequirements, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameVipRequirementsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameVipRequirements{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameVipRequirements{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameVipRequirementsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameVipRequirements) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameVipRequirementsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameVipRequirements{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameVipRequirementsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameVipRequirements) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
