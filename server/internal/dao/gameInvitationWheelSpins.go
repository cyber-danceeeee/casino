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

var _ GameInvitationWheelSpinsDao = (*gameInvitationWheelSpinsDao)(nil)

// GameInvitationWheelSpinsDao defining the dao interface
type GameInvitationWheelSpinsDao interface {
	Create(ctx context.Context, table *model.GameInvitationWheelSpins) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameInvitationWheelSpins) error
	GetByID(ctx context.Context, id uint64) (*model.GameInvitationWheelSpins, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameInvitationWheelSpins, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvitationWheelSpins) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvitationWheelSpins) error
}

type gameInvitationWheelSpinsDao struct {
	db    *gorm.DB
	cache cache.GameInvitationWheelSpinsCache // if nil, the cache is not used.
	sfg   *singleflight.Group                 // if cache is nil, the sfg is not used.
}

// NewGameInvitationWheelSpinsDao creating the dao interface
func NewGameInvitationWheelSpinsDao(db *gorm.DB, xCache cache.GameInvitationWheelSpinsCache) GameInvitationWheelSpinsDao {
	if xCache == nil {
		return &gameInvitationWheelSpinsDao{db: db}
	}
	return &gameInvitationWheelSpinsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameInvitationWheelSpinsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameInvitationWheelSpins, insert the record and the id value is written back to the table
func (d *gameInvitationWheelSpinsDao) Create(ctx context.Context, table *model.GameInvitationWheelSpins) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameInvitationWheelSpins by id
func (d *gameInvitationWheelSpinsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameInvitationWheelSpins{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameInvitationWheelSpins by id, support partial update
func (d *gameInvitationWheelSpinsDao) UpdateByID(ctx context.Context, table *model.GameInvitationWheelSpins) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameInvitationWheelSpinsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameInvitationWheelSpins) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.SpinNo != "" {
		update["spin_no"] = table.SpinNo
	}
	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.RoundID != 0 {
		update["round_id"] = table.RoundID
	}
	if table.PrizeID != 0 {
		update["prize_id"] = table.PrizeID
	}
	if table.RewardType != "" {
		update["reward_type"] = table.RewardType
	}
	if table.RewardAmount != nil && table.RewardAmount.IsZero() == false {
		update["reward_amount"] = table.RewardAmount
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.SpunAt != nil && table.SpunAt.IsZero() == false {
		update["spun_at"] = table.SpunAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameInvitationWheelSpins by id
func (d *gameInvitationWheelSpinsDao) GetByID(ctx context.Context, id uint64) (*model.GameInvitationWheelSpins, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameInvitationWheelSpins{}
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
			table := &model.GameInvitationWheelSpins{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameInvitationWheelSpinsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameInvitationWheelSpins)
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

// GetByColumns get a paginated list of gameInvitationWheelSpinss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameInvitationWheelSpinsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameInvitationWheelSpins, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameInvitationWheelSpinsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameInvitationWheelSpins{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameInvitationWheelSpins{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameInvitationWheelSpinsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvitationWheelSpins) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameInvitationWheelSpinsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameInvitationWheelSpins{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameInvitationWheelSpinsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvitationWheelSpins) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
