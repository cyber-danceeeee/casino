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

var _ GameAppVersionsDao = (*gameAppVersionsDao)(nil)

// GameAppVersionsDao defining the dao interface
type GameAppVersionsDao interface {
	Create(ctx context.Context, table *model.GameAppVersions) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameAppVersions) error
	GetByID(ctx context.Context, id uint64) (*model.GameAppVersions, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameAppVersions, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameAppVersions) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameAppVersions) error
}

type gameAppVersionsDao struct {
	db    *gorm.DB
	cache cache.GameAppVersionsCache // if nil, the cache is not used.
	sfg   *singleflight.Group        // if cache is nil, the sfg is not used.
}

// NewGameAppVersionsDao creating the dao interface
func NewGameAppVersionsDao(db *gorm.DB, xCache cache.GameAppVersionsCache) GameAppVersionsDao {
	if xCache == nil {
		return &gameAppVersionsDao{db: db}
	}
	return &gameAppVersionsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameAppVersionsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameAppVersions, insert the record and the id value is written back to the table
func (d *gameAppVersionsDao) Create(ctx context.Context, table *model.GameAppVersions) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameAppVersions by id
func (d *gameAppVersionsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameAppVersions{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameAppVersions by id, support partial update
func (d *gameAppVersionsDao) UpdateByID(ctx context.Context, table *model.GameAppVersions) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameAppVersionsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameAppVersions) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.Platform != "" {
		update["platform"] = table.Platform
	}
	if table.VersionName != "" {
		update["version_name"] = table.VersionName
	}
	if table.VersionCode != 0 {
		update["version_code"] = table.VersionCode
	}
	if table.DownloadURL != "" {
		update["download_url"] = table.DownloadURL
	}
	if table.ForceUpdate != 0 {
		update["force_update"] = table.ForceUpdate
	}
	if table.ReleaseNote != "" {
		update["release_note"] = table.ReleaseNote
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameAppVersions by id
func (d *gameAppVersionsDao) GetByID(ctx context.Context, id uint64) (*model.GameAppVersions, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameAppVersions{}
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
			table := &model.GameAppVersions{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameAppVersionsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameAppVersions)
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

// GetByColumns get a paginated list of gameAppVersionss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameAppVersionsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameAppVersions, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameAppVersionsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameAppVersions{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameAppVersions{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameAppVersionsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameAppVersions) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameAppVersionsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameAppVersions{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameAppVersionsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameAppVersions) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
