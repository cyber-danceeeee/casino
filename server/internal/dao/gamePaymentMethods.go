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

var _ GamePaymentMethodsDao = (*gamePaymentMethodsDao)(nil)

// GamePaymentMethodsDao defining the dao interface
type GamePaymentMethodsDao interface {
	Create(ctx context.Context, table *model.GamePaymentMethods) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GamePaymentMethods) error
	GetByID(ctx context.Context, id uint64) (*model.GamePaymentMethods, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GamePaymentMethods, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GamePaymentMethods) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GamePaymentMethods) error
}

type gamePaymentMethodsDao struct {
	db    *gorm.DB
	cache cache.GamePaymentMethodsCache // if nil, the cache is not used.
	sfg   *singleflight.Group           // if cache is nil, the sfg is not used.
}

// NewGamePaymentMethodsDao creating the dao interface
func NewGamePaymentMethodsDao(db *gorm.DB, xCache cache.GamePaymentMethodsCache) GamePaymentMethodsDao {
	if xCache == nil {
		return &gamePaymentMethodsDao{db: db}
	}
	return &gamePaymentMethodsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gamePaymentMethodsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gamePaymentMethods, insert the record and the id value is written back to the table
func (d *gamePaymentMethodsDao) Create(ctx context.Context, table *model.GamePaymentMethods) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gamePaymentMethods by id
func (d *gamePaymentMethodsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GamePaymentMethods{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gamePaymentMethods by id, support partial update
func (d *gamePaymentMethodsDao) UpdateByID(ctx context.Context, table *model.GamePaymentMethods) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gamePaymentMethodsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GamePaymentMethods) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.MethodCode != "" {
		update["method_code"] = table.MethodCode
	}
	if table.Name != "" {
		update["name"] = table.Name
	}
	if table.MethodType != "" {
		update["method_type"] = table.MethodType
	}
	if table.IconURL != "" {
		update["icon_url"] = table.IconURL
	}
	if table.MinAmount != nil && table.MinAmount.IsZero() == false {
		update["min_amount"] = table.MinAmount
	}
	if table.MaxAmount != nil && table.MaxAmount.IsZero() == false {
		update["max_amount"] = table.MaxAmount
	}
	if table.Sort != 0 {
		update["sort"] = table.Sort
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gamePaymentMethods by id
func (d *gamePaymentMethodsDao) GetByID(ctx context.Context, id uint64) (*model.GamePaymentMethods, error) {
	// no cache
	if d.cache == nil {
		record := &model.GamePaymentMethods{}
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
			table := &model.GamePaymentMethods{}
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
			if err = d.cache.Set(ctx, id, table, cache.GamePaymentMethodsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GamePaymentMethods)
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

// GetByColumns get a paginated list of gamePaymentMethodss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gamePaymentMethodsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GamePaymentMethods, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GamePaymentMethodsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GamePaymentMethods{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GamePaymentMethods{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gamePaymentMethodsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GamePaymentMethods) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gamePaymentMethodsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GamePaymentMethods{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gamePaymentMethodsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GamePaymentMethods) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
