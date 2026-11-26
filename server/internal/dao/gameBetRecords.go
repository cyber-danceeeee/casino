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

var _ GameBetRecordsDao = (*gameBetRecordsDao)(nil)

// GameBetRecordsDao defining the dao interface
type GameBetRecordsDao interface {
	Create(ctx context.Context, table *model.GameBetRecords) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameBetRecords) error
	GetByID(ctx context.Context, id uint64) (*model.GameBetRecords, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameBetRecords, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBetRecords) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBetRecords) error
}

type gameBetRecordsDao struct {
	db    *gorm.DB
	cache cache.GameBetRecordsCache // if nil, the cache is not used.
	sfg   *singleflight.Group       // if cache is nil, the sfg is not used.
}

// NewGameBetRecordsDao creating the dao interface
func NewGameBetRecordsDao(db *gorm.DB, xCache cache.GameBetRecordsCache) GameBetRecordsDao {
	if xCache == nil {
		return &gameBetRecordsDao{db: db}
	}
	return &gameBetRecordsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameBetRecordsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameBetRecords, insert the record and the id value is written back to the table
func (d *gameBetRecordsDao) Create(ctx context.Context, table *model.GameBetRecords) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameBetRecords by id
func (d *gameBetRecordsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameBetRecords{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameBetRecords by id, support partial update
func (d *gameBetRecordsDao) UpdateByID(ctx context.Context, table *model.GameBetRecords) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameBetRecordsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameBetRecords) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.BetNo != "" {
		update["bet_no"] = table.BetNo
	}
	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.GameID != 0 {
		update["game_id"] = table.GameID
	}
	if table.ProviderID != 0 {
		update["provider_id"] = table.ProviderID
	}
	if table.GameType != "" {
		update["game_type"] = table.GameType
	}
	if table.BetAmount != nil && table.BetAmount.IsZero() == false {
		update["bet_amount"] = table.BetAmount
	}
	if table.ValidBetAmount != nil && table.ValidBetAmount.IsZero() == false {
		update["valid_bet_amount"] = table.ValidBetAmount
	}
	if table.PayoutAmount != nil && table.PayoutAmount.IsZero() == false {
		update["payout_amount"] = table.PayoutAmount
	}
	if table.ProfitAmount != nil && table.ProfitAmount.IsZero() == false {
		update["profit_amount"] = table.ProfitAmount
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.BetAt != nil && table.BetAt.IsZero() == false {
		update["bet_at"] = table.BetAt
	}
	if table.SettledAt != nil && table.SettledAt.IsZero() == false {
		update["settled_at"] = table.SettledAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameBetRecords by id
func (d *gameBetRecordsDao) GetByID(ctx context.Context, id uint64) (*model.GameBetRecords, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameBetRecords{}
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
			table := &model.GameBetRecords{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameBetRecordsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameBetRecords)
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

// GetByColumns get a paginated list of gameBetRecordss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameBetRecordsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameBetRecords, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameBetRecordsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameBetRecords{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameBetRecords{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameBetRecordsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBetRecords) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameBetRecordsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameBetRecords{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameBetRecordsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBetRecords) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
