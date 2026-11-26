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

var _ GameHomeFeatureCardsDao = (*gameHomeFeatureCardsDao)(nil)

// GameHomeFeatureCardsDao defining the dao interface
type GameHomeFeatureCardsDao interface {
	Create(ctx context.Context, table *model.GameHomeFeatureCards) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameHomeFeatureCards) error
	GetByID(ctx context.Context, id uint64) (*model.GameHomeFeatureCards, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameHomeFeatureCards, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameHomeFeatureCards) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameHomeFeatureCards) error
}

type gameHomeFeatureCardsDao struct {
	db    *gorm.DB
	cache cache.GameHomeFeatureCardsCache // if nil, the cache is not used.
	sfg   *singleflight.Group             // if cache is nil, the sfg is not used.
}

// NewGameHomeFeatureCardsDao creating the dao interface
func NewGameHomeFeatureCardsDao(db *gorm.DB, xCache cache.GameHomeFeatureCardsCache) GameHomeFeatureCardsDao {
	if xCache == nil {
		return &gameHomeFeatureCardsDao{db: db}
	}
	return &gameHomeFeatureCardsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameHomeFeatureCardsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameHomeFeatureCards, insert the record and the id value is written back to the table
func (d *gameHomeFeatureCardsDao) Create(ctx context.Context, table *model.GameHomeFeatureCards) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameHomeFeatureCards by id
func (d *gameHomeFeatureCardsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameHomeFeatureCards{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameHomeFeatureCards by id, support partial update
func (d *gameHomeFeatureCardsDao) UpdateByID(ctx context.Context, table *model.GameHomeFeatureCards) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameHomeFeatureCardsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameHomeFeatureCards) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.FeatureCode != "" {
		update["feature_code"] = table.FeatureCode
	}
	if table.Title != "" {
		update["title"] = table.Title
	}
	if table.ImageURL != "" {
		update["image_url"] = table.ImageURL
	}
	if table.BackgroundStyle != "" {
		update["background_style"] = table.BackgroundStyle
	}
	if table.LinkType != "" {
		update["link_type"] = table.LinkType
	}
	if table.LinkURL != "" {
		update["link_url"] = table.LinkURL
	}
	if table.CountdownEndAt != nil && table.CountdownEndAt.IsZero() == false {
		update["countdown_end_at"] = table.CountdownEndAt
	}
	if table.Sort != 0 {
		update["sort"] = table.Sort
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameHomeFeatureCards by id
func (d *gameHomeFeatureCardsDao) GetByID(ctx context.Context, id uint64) (*model.GameHomeFeatureCards, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameHomeFeatureCards{}
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
			table := &model.GameHomeFeatureCards{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameHomeFeatureCardsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameHomeFeatureCards)
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

// GetByColumns get a paginated list of gameHomeFeatureCardss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameHomeFeatureCardsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameHomeFeatureCards, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameHomeFeatureCardsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameHomeFeatureCards{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameHomeFeatureCards{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameHomeFeatureCardsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameHomeFeatureCards) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameHomeFeatureCardsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameHomeFeatureCards{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameHomeFeatureCardsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameHomeFeatureCards) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
