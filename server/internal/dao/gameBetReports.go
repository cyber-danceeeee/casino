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

var _ GameBetReportsDao = (*gameBetReportsDao)(nil)

// GameBetReportsDao defining the dao interface
type GameBetReportsDao interface {
	Create(ctx context.Context, table *model.GameBetReports) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameBetReports) error
	GetByID(ctx context.Context, id uint64) (*model.GameBetReports, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameBetReports, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBetReports) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBetReports) error
}

type gameBetReportsDao struct {
	db    *gorm.DB
	cache cache.GameBetReportsCache // if nil, the cache is not used.
	sfg   *singleflight.Group       // if cache is nil, the sfg is not used.
}

// NewGameBetReportsDao creating the dao interface
func NewGameBetReportsDao(db *gorm.DB, xCache cache.GameBetReportsCache) GameBetReportsDao {
	if xCache == nil {
		return &gameBetReportsDao{db: db}
	}
	return &gameBetReportsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameBetReportsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameBetReports, insert the record and the id value is written back to the table
func (d *gameBetReportsDao) Create(ctx context.Context, table *model.GameBetReports) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameBetReports by id
func (d *gameBetReportsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameBetReports{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameBetReports by id, support partial update
func (d *gameBetReportsDao) UpdateByID(ctx context.Context, table *model.GameBetReports) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameBetReportsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameBetReports) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.ReportDate != nil && table.ReportDate.IsZero() == false {
		update["report_date"] = table.ReportDate
	}
	if table.GameType != "" {
		update["game_type"] = table.GameType
	}
	if table.BetCount != 0 {
		update["bet_count"] = table.BetCount
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

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameBetReports by id
func (d *gameBetReportsDao) GetByID(ctx context.Context, id uint64) (*model.GameBetReports, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameBetReports{}
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
			table := &model.GameBetReports{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameBetReportsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameBetReports)
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

// GetByColumns get a paginated list of gameBetReportss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameBetReportsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameBetReports, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameBetReportsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameBetReports{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameBetReports{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameBetReportsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBetReports) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameBetReportsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameBetReports{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameBetReportsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBetReports) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
