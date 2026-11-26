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

var _ GameRankingPeriodsDao = (*gameRankingPeriodsDao)(nil)

// GameRankingPeriodsDao defining the dao interface
type GameRankingPeriodsDao interface {
	Create(ctx context.Context, table *model.GameRankingPeriods) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameRankingPeriods) error
	GetByID(ctx context.Context, id uint64) (*model.GameRankingPeriods, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameRankingPeriods, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameRankingPeriods) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameRankingPeriods) error
}

type gameRankingPeriodsDao struct {
	db    *gorm.DB
	cache cache.GameRankingPeriodsCache // if nil, the cache is not used.
	sfg   *singleflight.Group           // if cache is nil, the sfg is not used.
}

// NewGameRankingPeriodsDao creating the dao interface
func NewGameRankingPeriodsDao(db *gorm.DB, xCache cache.GameRankingPeriodsCache) GameRankingPeriodsDao {
	if xCache == nil {
		return &gameRankingPeriodsDao{db: db}
	}
	return &gameRankingPeriodsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameRankingPeriodsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameRankingPeriods, insert the record and the id value is written back to the table
func (d *gameRankingPeriodsDao) Create(ctx context.Context, table *model.GameRankingPeriods) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameRankingPeriods by id
func (d *gameRankingPeriodsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameRankingPeriods{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameRankingPeriods by id, support partial update
func (d *gameRankingPeriodsDao) UpdateByID(ctx context.Context, table *model.GameRankingPeriods) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameRankingPeriodsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameRankingPeriods) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.RankingType != "" {
		update["ranking_type"] = table.RankingType
	}
	if table.Title != "" {
		update["title"] = table.Title
	}
	if table.RankingDate != nil && table.RankingDate.IsZero() == false {
		update["ranking_date"] = table.RankingDate
	}
	if table.PeriodStart != nil && table.PeriodStart.IsZero() == false {
		update["period_start"] = table.PeriodStart
	}
	if table.PeriodEnd != nil && table.PeriodEnd.IsZero() == false {
		update["period_end"] = table.PeriodEnd
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameRankingPeriods by id
func (d *gameRankingPeriodsDao) GetByID(ctx context.Context, id uint64) (*model.GameRankingPeriods, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameRankingPeriods{}
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
			table := &model.GameRankingPeriods{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameRankingPeriodsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameRankingPeriods)
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

// GetByColumns get a paginated list of gameRankingPeriodss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameRankingPeriodsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameRankingPeriods, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameRankingPeriodsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameRankingPeriods{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameRankingPeriods{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameRankingPeriodsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameRankingPeriods) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameRankingPeriodsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameRankingPeriods{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameRankingPeriodsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameRankingPeriods) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
