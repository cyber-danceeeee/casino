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

var _ GameSmsVerificationCodesDao = (*gameSmsVerificationCodesDao)(nil)

// GameSmsVerificationCodesDao defining the dao interface
type GameSmsVerificationCodesDao interface {
	Create(ctx context.Context, table *model.GameSmsVerificationCodes) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameSmsVerificationCodes) error
	GetByID(ctx context.Context, id uint64) (*model.GameSmsVerificationCodes, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameSmsVerificationCodes, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSmsVerificationCodes) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSmsVerificationCodes) error
}

type gameSmsVerificationCodesDao struct {
	db    *gorm.DB
	cache cache.GameSmsVerificationCodesCache // if nil, the cache is not used.
	sfg   *singleflight.Group                 // if cache is nil, the sfg is not used.
}

// NewGameSmsVerificationCodesDao creating the dao interface
func NewGameSmsVerificationCodesDao(db *gorm.DB, xCache cache.GameSmsVerificationCodesCache) GameSmsVerificationCodesDao {
	if xCache == nil {
		return &gameSmsVerificationCodesDao{db: db}
	}
	return &gameSmsVerificationCodesDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameSmsVerificationCodesDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameSmsVerificationCodes, insert the record and the id value is written back to the table
func (d *gameSmsVerificationCodesDao) Create(ctx context.Context, table *model.GameSmsVerificationCodes) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameSmsVerificationCodes by id
func (d *gameSmsVerificationCodesDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameSmsVerificationCodes{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameSmsVerificationCodes by id, support partial update
func (d *gameSmsVerificationCodesDao) UpdateByID(ctx context.Context, table *model.GameSmsVerificationCodes) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameSmsVerificationCodesDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameSmsVerificationCodes) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.PhoneCountryCode != "" {
		update["phone_country_code"] = table.PhoneCountryCode
	}
	if table.Phone != "" {
		update["phone"] = table.Phone
	}
	if table.CodeHash != "" {
		update["code_hash"] = table.CodeHash
	}
	if table.Scene != "" {
		update["scene"] = table.Scene
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.ExpiredAt != nil && table.ExpiredAt.IsZero() == false {
		update["expired_at"] = table.ExpiredAt
	}
	if table.UsedAt != nil && table.UsedAt.IsZero() == false {
		update["used_at"] = table.UsedAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameSmsVerificationCodes by id
func (d *gameSmsVerificationCodesDao) GetByID(ctx context.Context, id uint64) (*model.GameSmsVerificationCodes, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameSmsVerificationCodes{}
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
			table := &model.GameSmsVerificationCodes{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameSmsVerificationCodesExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameSmsVerificationCodes)
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

// GetByColumns get a paginated list of gameSmsVerificationCodess by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameSmsVerificationCodesDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameSmsVerificationCodes, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameSmsVerificationCodesColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameSmsVerificationCodes{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameSmsVerificationCodes{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameSmsVerificationCodesDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSmsVerificationCodes) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameSmsVerificationCodesDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameSmsVerificationCodes{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameSmsVerificationCodesDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSmsVerificationCodes) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
