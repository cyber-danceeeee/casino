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

var _ GameActivityEventsDao = (*gameActivityEventsDao)(nil)

// GameActivityEventsDao defining the dao interface
type GameActivityEventsDao interface {
	Create(ctx context.Context, table *model.GameActivityEvents) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameActivityEvents) error
	GetByID(ctx context.Context, id uint64) (*model.GameActivityEvents, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameActivityEvents, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameActivityEvents) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameActivityEvents) error
}

type gameActivityEventsDao struct {
	db    *gorm.DB
	cache cache.GameActivityEventsCache // if nil, the cache is not used.
	sfg   *singleflight.Group           // if cache is nil, the sfg is not used.
}

// NewGameActivityEventsDao creating the dao interface
func NewGameActivityEventsDao(db *gorm.DB, xCache cache.GameActivityEventsCache) GameActivityEventsDao {
	if xCache == nil {
		return &gameActivityEventsDao{db: db}
	}
	return &gameActivityEventsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameActivityEventsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameActivityEvents, insert the record and the id value is written back to the table
func (d *gameActivityEventsDao) Create(ctx context.Context, table *model.GameActivityEvents) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameActivityEvents by id
func (d *gameActivityEventsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameActivityEvents{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameActivityEvents by id, support partial update
func (d *gameActivityEventsDao) UpdateByID(ctx context.Context, table *model.GameActivityEvents) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameActivityEventsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameActivityEvents) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.ActivityCode != "" {
		update["activity_code"] = table.ActivityCode
	}
	if table.Slug != "" {
		update["slug"] = table.Slug
	}
	if table.Kind != "" {
		update["kind"] = table.Kind
	}
	if table.Title != "" {
		update["title"] = table.Title
	}
	if table.Subtitle != "" {
		update["subtitle"] = table.Subtitle
	}
	if table.RewardText != "" {
		update["reward_text"] = table.RewardText
	}
	if table.MaxRewardText != "" {
		update["max_reward_text"] = table.MaxRewardText
	}
	if table.Note != "" {
		update["note"] = table.Note
	}
	if table.CtaText != "" {
		update["cta_text"] = table.CtaText
	}
	if table.HeroURL != "" {
		update["hero_url"] = table.HeroURL
	}
	if table.CardImageURL != "" {
		update["card_image_url"] = table.CardImageURL
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.Sort != 0 {
		update["sort"] = table.Sort
	}
	if table.StartAt != nil && table.StartAt.IsZero() == false {
		update["start_at"] = table.StartAt
	}
	if table.EndAt != nil && table.EndAt.IsZero() == false {
		update["end_at"] = table.EndAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameActivityEvents by id
func (d *gameActivityEventsDao) GetByID(ctx context.Context, id uint64) (*model.GameActivityEvents, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameActivityEvents{}
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
			table := &model.GameActivityEvents{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameActivityEventsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameActivityEvents)
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

// GetByColumns get a paginated list of gameActivityEventss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameActivityEventsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameActivityEvents, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameActivityEventsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameActivityEvents{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameActivityEvents{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameActivityEventsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameActivityEvents) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameActivityEventsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameActivityEvents{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameActivityEventsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameActivityEvents) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
