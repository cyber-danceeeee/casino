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

var _ GameSupportCategoriesDao = (*gameSupportCategoriesDao)(nil)

// GameSupportCategoriesDao defining the dao interface
type GameSupportCategoriesDao interface {
	Create(ctx context.Context, table *model.GameSupportCategories) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameSupportCategories) error
	GetByID(ctx context.Context, id uint64) (*model.GameSupportCategories, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameSupportCategories, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSupportCategories) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSupportCategories) error
}

type gameSupportCategoriesDao struct {
	db    *gorm.DB
	cache cache.GameSupportCategoriesCache // if nil, the cache is not used.
	sfg   *singleflight.Group              // if cache is nil, the sfg is not used.
}

// NewGameSupportCategoriesDao creating the dao interface
func NewGameSupportCategoriesDao(db *gorm.DB, xCache cache.GameSupportCategoriesCache) GameSupportCategoriesDao {
	if xCache == nil {
		return &gameSupportCategoriesDao{db: db}
	}
	return &gameSupportCategoriesDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameSupportCategoriesDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameSupportCategories, insert the record and the id value is written back to the table
func (d *gameSupportCategoriesDao) Create(ctx context.Context, table *model.GameSupportCategories) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameSupportCategories by id
func (d *gameSupportCategoriesDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameSupportCategories{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameSupportCategories by id, support partial update
func (d *gameSupportCategoriesDao) UpdateByID(ctx context.Context, table *model.GameSupportCategories) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameSupportCategoriesDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameSupportCategories) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.CategoryCode != "" {
		update["category_code"] = table.CategoryCode
	}
	if table.Name != "" {
		update["name"] = table.Name
	}
	if table.Sort != 0 {
		update["sort"] = table.Sort
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameSupportCategories by id
func (d *gameSupportCategoriesDao) GetByID(ctx context.Context, id uint64) (*model.GameSupportCategories, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameSupportCategories{}
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
			table := &model.GameSupportCategories{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameSupportCategoriesExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameSupportCategories)
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

// GetByColumns get a paginated list of gameSupportCategoriess by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameSupportCategoriesDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameSupportCategories, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameSupportCategoriesColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameSupportCategories{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameSupportCategories{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameSupportCategoriesDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSupportCategories) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameSupportCategoriesDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameSupportCategories{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameSupportCategoriesDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSupportCategories) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
