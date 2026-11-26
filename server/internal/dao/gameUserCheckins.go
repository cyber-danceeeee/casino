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

var _ GameUserCheckinsDao = (*gameUserCheckinsDao)(nil)

// GameUserCheckinsDao defining the dao interface
type GameUserCheckinsDao interface {
	Create(ctx context.Context, table *model.GameUserCheckins) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameUserCheckins) error
	GetByID(ctx context.Context, id uint64) (*model.GameUserCheckins, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUserCheckins, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserCheckins) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserCheckins) error
}

type gameUserCheckinsDao struct {
	db    *gorm.DB
	cache cache.GameUserCheckinsCache // if nil, the cache is not used.
	sfg   *singleflight.Group         // if cache is nil, the sfg is not used.
}

// NewGameUserCheckinsDao creating the dao interface
func NewGameUserCheckinsDao(db *gorm.DB, xCache cache.GameUserCheckinsCache) GameUserCheckinsDao {
	if xCache == nil {
		return &gameUserCheckinsDao{db: db}
	}
	return &gameUserCheckinsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameUserCheckinsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameUserCheckins, insert the record and the id value is written back to the table
func (d *gameUserCheckinsDao) Create(ctx context.Context, table *model.GameUserCheckins) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameUserCheckins by id
func (d *gameUserCheckinsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUserCheckins{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameUserCheckins by id, support partial update
func (d *gameUserCheckinsDao) UpdateByID(ctx context.Context, table *model.GameUserCheckins) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameUserCheckinsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameUserCheckins) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.ActivityID != 0 {
		update["activity_id"] = table.ActivityID
	}
	if table.CheckinDate != nil && table.CheckinDate.IsZero() == false {
		update["checkin_date"] = table.CheckinDate
	}
	if table.DayNo != 0 {
		update["day_no"] = table.DayNo
	}
	if table.StreakCount != 0 {
		update["streak_count"] = table.StreakCount
	}
	if table.RewardAmount != nil && table.RewardAmount.IsZero() == false {
		update["reward_amount"] = table.RewardAmount
	}
	if table.RewardWallet != "" {
		update["reward_wallet"] = table.RewardWallet
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.ClaimedAt != nil && table.ClaimedAt.IsZero() == false {
		update["claimed_at"] = table.ClaimedAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameUserCheckins by id
func (d *gameUserCheckinsDao) GetByID(ctx context.Context, id uint64) (*model.GameUserCheckins, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameUserCheckins{}
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
			table := &model.GameUserCheckins{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameUserCheckinsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameUserCheckins)
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

// GetByColumns get a paginated list of gameUserCheckinss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameUserCheckinsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUserCheckins, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameUserCheckinsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameUserCheckins{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameUserCheckins{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameUserCheckinsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserCheckins) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameUserCheckinsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUserCheckins{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameUserCheckinsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserCheckins) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
