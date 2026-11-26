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

var _ GameCashAuditRecordsDao = (*gameCashAuditRecordsDao)(nil)

// GameCashAuditRecordsDao defining the dao interface
type GameCashAuditRecordsDao interface {
	Create(ctx context.Context, table *model.GameCashAuditRecords) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameCashAuditRecords) error
	GetByID(ctx context.Context, id uint64) (*model.GameCashAuditRecords, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameCashAuditRecords, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameCashAuditRecords) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameCashAuditRecords) error
}

type gameCashAuditRecordsDao struct {
	db    *gorm.DB
	cache cache.GameCashAuditRecordsCache // if nil, the cache is not used.
	sfg   *singleflight.Group             // if cache is nil, the sfg is not used.
}

// NewGameCashAuditRecordsDao creating the dao interface
func NewGameCashAuditRecordsDao(db *gorm.DB, xCache cache.GameCashAuditRecordsCache) GameCashAuditRecordsDao {
	if xCache == nil {
		return &gameCashAuditRecordsDao{db: db}
	}
	return &gameCashAuditRecordsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameCashAuditRecordsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameCashAuditRecords, insert the record and the id value is written back to the table
func (d *gameCashAuditRecordsDao) Create(ctx context.Context, table *model.GameCashAuditRecords) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameCashAuditRecords by id
func (d *gameCashAuditRecordsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameCashAuditRecords{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameCashAuditRecords by id, support partial update
func (d *gameCashAuditRecordsDao) UpdateByID(ctx context.Context, table *model.GameCashAuditRecords) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameCashAuditRecordsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameCashAuditRecords) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.SourceType != "" {
		update["source_type"] = table.SourceType
	}
	if table.SourceID != 0 {
		update["source_id"] = table.SourceID
	}
	if table.Amount != nil && table.Amount.IsZero() == false {
		update["amount"] = table.Amount
	}
	if table.RequirementMultiplier != nil && table.RequirementMultiplier.IsZero() == false {
		update["requirement_multiplier"] = table.RequirementMultiplier
	}
	if table.RequiredValidBet != nil && table.RequiredValidBet.IsZero() == false {
		update["required_valid_bet"] = table.RequiredValidBet
	}
	if table.CompletedValidBet != nil && table.CompletedValidBet.IsZero() == false {
		update["completed_valid_bet"] = table.CompletedValidBet
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.StartedAt != nil && table.StartedAt.IsZero() == false {
		update["started_at"] = table.StartedAt
	}
	if table.CompletedAt != nil && table.CompletedAt.IsZero() == false {
		update["completed_at"] = table.CompletedAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameCashAuditRecords by id
func (d *gameCashAuditRecordsDao) GetByID(ctx context.Context, id uint64) (*model.GameCashAuditRecords, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameCashAuditRecords{}
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
			table := &model.GameCashAuditRecords{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameCashAuditRecordsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameCashAuditRecords)
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

// GetByColumns get a paginated list of gameCashAuditRecordss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameCashAuditRecordsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameCashAuditRecords, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameCashAuditRecordsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameCashAuditRecords{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameCashAuditRecords{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameCashAuditRecordsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameCashAuditRecords) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameCashAuditRecordsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameCashAuditRecords{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameCashAuditRecordsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameCashAuditRecords) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
