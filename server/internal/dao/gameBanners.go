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

var _ GameBannersDao = (*gameBannersDao)(nil)

// GameBannersDao defining the dao interface
type GameBannersDao interface {
	Create(ctx context.Context, table *model.GameBanners) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameBanners) error
	GetByID(ctx context.Context, id uint64) (*model.GameBanners, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameBanners, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBanners) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBanners) error
}

type gameBannersDao struct {
	db    *gorm.DB
	cache cache.GameBannersCache // if nil, the cache is not used.
	sfg   *singleflight.Group    // if cache is nil, the sfg is not used.
}

// NewGameBannersDao creating the dao interface
func NewGameBannersDao(db *gorm.DB, xCache cache.GameBannersCache) GameBannersDao {
	if xCache == nil {
		return &gameBannersDao{db: db}
	}
	return &gameBannersDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameBannersDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameBanners, insert the record and the id value is written back to the table
func (d *gameBannersDao) Create(ctx context.Context, table *model.GameBanners) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameBanners by id
func (d *gameBannersDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameBanners{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameBanners by id, support partial update
func (d *gameBannersDao) UpdateByID(ctx context.Context, table *model.GameBanners) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameBannersDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameBanners) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.BannerCode != "" {
		update["banner_code"] = table.BannerCode
	}
	if table.Position != "" {
		update["position"] = table.Position
	}
	if table.Title != "" {
		update["title"] = table.Title
	}
	if table.ImageURL != "" {
		update["image_url"] = table.ImageURL
	}
	if table.LinkType != "" {
		update["link_type"] = table.LinkType
	}
	if table.LinkURL != "" {
		update["link_url"] = table.LinkURL
	}
	if table.ActivitySlug != "" {
		update["activity_slug"] = table.ActivitySlug
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

// GetByID get a gameBanners by id
func (d *gameBannersDao) GetByID(ctx context.Context, id uint64) (*model.GameBanners, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameBanners{}
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
			table := &model.GameBanners{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameBannersExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameBanners)
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

// GetByColumns get a paginated list of gameBannerss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameBannersDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameBanners, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameBannersColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameBanners{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameBanners{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameBannersDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBanners) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameBannersDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameBanners{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameBannersDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameBanners) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
