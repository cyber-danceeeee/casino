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

var _ GameAgentCommissionRecordsDao = (*gameAgentCommissionRecordsDao)(nil)

// GameAgentCommissionRecordsDao defining the dao interface
type GameAgentCommissionRecordsDao interface {
	Create(ctx context.Context, table *model.GameAgentCommissionRecords) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameAgentCommissionRecords) error
	GetByID(ctx context.Context, id uint64) (*model.GameAgentCommissionRecords, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameAgentCommissionRecords, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameAgentCommissionRecords) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameAgentCommissionRecords) error
}

type gameAgentCommissionRecordsDao struct {
	db    *gorm.DB
	cache cache.GameAgentCommissionRecordsCache // if nil, the cache is not used.
	sfg   *singleflight.Group                   // if cache is nil, the sfg is not used.
}

// NewGameAgentCommissionRecordsDao creating the dao interface
func NewGameAgentCommissionRecordsDao(db *gorm.DB, xCache cache.GameAgentCommissionRecordsCache) GameAgentCommissionRecordsDao {
	if xCache == nil {
		return &gameAgentCommissionRecordsDao{db: db}
	}
	return &gameAgentCommissionRecordsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameAgentCommissionRecordsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameAgentCommissionRecords, insert the record and the id value is written back to the table
func (d *gameAgentCommissionRecordsDao) Create(ctx context.Context, table *model.GameAgentCommissionRecords) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameAgentCommissionRecords by id
func (d *gameAgentCommissionRecordsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameAgentCommissionRecords{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameAgentCommissionRecords by id, support partial update
func (d *gameAgentCommissionRecordsDao) UpdateByID(ctx context.Context, table *model.GameAgentCommissionRecords) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameAgentCommissionRecordsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameAgentCommissionRecords) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.PeriodStart != nil && table.PeriodStart.IsZero() == false {
		update["period_start"] = table.PeriodStart
	}
	if table.PeriodEnd != nil && table.PeriodEnd.IsZero() == false {
		update["period_end"] = table.PeriodEnd
	}
	if table.ReferralsCount != 0 {
		update["referrals_count"] = table.ReferralsCount
	}
	if table.ValidBetAmount != nil && table.ValidBetAmount.IsZero() == false {
		update["valid_bet_amount"] = table.ValidBetAmount
	}
	if table.CommissionAmount != nil && table.CommissionAmount.IsZero() == false {
		update["commission_amount"] = table.CommissionAmount
	}
	if table.BonusAmount != nil && table.BonusAmount.IsZero() == false {
		update["bonus_amount"] = table.BonusAmount
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.SettledAt != nil && table.SettledAt.IsZero() == false {
		update["settled_at"] = table.SettledAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameAgentCommissionRecords by id
func (d *gameAgentCommissionRecordsDao) GetByID(ctx context.Context, id uint64) (*model.GameAgentCommissionRecords, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameAgentCommissionRecords{}
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
			table := &model.GameAgentCommissionRecords{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameAgentCommissionRecordsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameAgentCommissionRecords)
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

// GetByColumns get a paginated list of gameAgentCommissionRecordss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameAgentCommissionRecordsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameAgentCommissionRecords, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameAgentCommissionRecordsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameAgentCommissionRecords{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameAgentCommissionRecords{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameAgentCommissionRecordsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameAgentCommissionRecords) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameAgentCommissionRecordsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameAgentCommissionRecords{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameAgentCommissionRecordsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameAgentCommissionRecords) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
