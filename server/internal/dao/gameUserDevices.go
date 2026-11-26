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

var _ GameUserDevicesDao = (*gameUserDevicesDao)(nil)

// GameUserDevicesDao defining the dao interface
type GameUserDevicesDao interface {
	Create(ctx context.Context, table *model.GameUserDevices) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameUserDevices) error
	GetByID(ctx context.Context, id uint64) (*model.GameUserDevices, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUserDevices, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserDevices) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserDevices) error
}

type gameUserDevicesDao struct {
	db    *gorm.DB
	cache cache.GameUserDevicesCache // if nil, the cache is not used.
	sfg   *singleflight.Group        // if cache is nil, the sfg is not used.
}

// NewGameUserDevicesDao creating the dao interface
func NewGameUserDevicesDao(db *gorm.DB, xCache cache.GameUserDevicesCache) GameUserDevicesDao {
	if xCache == nil {
		return &gameUserDevicesDao{db: db}
	}
	return &gameUserDevicesDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameUserDevicesDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameUserDevices, insert the record and the id value is written back to the table
func (d *gameUserDevicesDao) Create(ctx context.Context, table *model.GameUserDevices) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameUserDevices by id
func (d *gameUserDevicesDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUserDevices{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameUserDevices by id, support partial update
func (d *gameUserDevicesDao) UpdateByID(ctx context.Context, table *model.GameUserDevices) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameUserDevicesDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameUserDevices) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.DeviceID != "" {
		update["device_id"] = table.DeviceID
	}
	if table.Platform != "" {
		update["platform"] = table.Platform
	}
	if table.AppVersion != "" {
		update["app_version"] = table.AppVersion
	}
	if table.ClientIP != "" {
		update["client_ip"] = table.ClientIP
	}
	if table.UserAgent != "" {
		update["user_agent"] = table.UserAgent
	}
	if table.FirstSeenAt != nil && table.FirstSeenAt.IsZero() == false {
		update["first_seen_at"] = table.FirstSeenAt
	}
	if table.LastSeenAt != nil && table.LastSeenAt.IsZero() == false {
		update["last_seen_at"] = table.LastSeenAt
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameUserDevices by id
func (d *gameUserDevicesDao) GetByID(ctx context.Context, id uint64) (*model.GameUserDevices, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameUserDevices{}
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
			table := &model.GameUserDevices{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameUserDevicesExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameUserDevices)
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

// GetByColumns get a paginated list of gameUserDevicess by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameUserDevicesDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUserDevices, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameUserDevicesColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameUserDevices{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameUserDevices{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameUserDevicesDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserDevices) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameUserDevicesDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUserDevices{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameUserDevicesDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserDevices) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
