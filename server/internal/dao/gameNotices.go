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

var _ GameNoticesDao = (*gameNoticesDao)(nil)

// GameNoticesDao defining the dao interface
type GameNoticesDao interface {
	Create(ctx context.Context, table *model.GameNotices) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameNotices) error
	GetByID(ctx context.Context, id uint64) (*model.GameNotices, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameNotices, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameNotices) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameNotices) error
}

type gameNoticesDao struct {
	db    *gorm.DB
	cache cache.GameNoticesCache // if nil, the cache is not used.
	sfg   *singleflight.Group    // if cache is nil, the sfg is not used.
}

// NewGameNoticesDao creating the dao interface
func NewGameNoticesDao(db *gorm.DB, xCache cache.GameNoticesCache) GameNoticesDao {
	if xCache == nil {
		return &gameNoticesDao{db: db}
	}
	return &gameNoticesDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameNoticesDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameNotices, insert the record and the id value is written back to the table
func (d *gameNoticesDao) Create(ctx context.Context, table *model.GameNotices) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameNotices by id
func (d *gameNoticesDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameNotices{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameNotices by id, support partial update
func (d *gameNoticesDao) UpdateByID(ctx context.Context, table *model.GameNotices) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameNoticesDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameNotices) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.NoticeCode != "" {
		update["notice_code"] = table.NoticeCode
	}
	if table.NoticeType != "" {
		update["notice_type"] = table.NoticeType
	}
	if table.Title != "" {
		update["title"] = table.Title
	}
	if table.ContentHTML != "" {
		update["content_html"] = table.ContentHTML
	}
	if table.ImageURL != "" {
		update["image_url"] = table.ImageURL
	}
	if table.Sort != 0 {
		update["sort"] = table.Sort
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.StartAt != nil && table.StartAt.IsZero() == false {
		update["start_at"] = table.StartAt
	}
	if table.EndAt != nil && table.EndAt.IsZero() == false {
		update["end_at"] = table.EndAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameNotices by id
func (d *gameNoticesDao) GetByID(ctx context.Context, id uint64) (*model.GameNotices, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameNotices{}
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
			table := &model.GameNotices{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameNoticesExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameNotices)
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

// GetByColumns get a paginated list of gameNoticess by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameNoticesDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameNotices, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameNoticesColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameNotices{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameNotices{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameNoticesDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameNotices) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameNoticesDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameNotices{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameNoticesDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameNotices) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
