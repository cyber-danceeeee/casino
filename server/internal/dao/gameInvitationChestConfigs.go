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

var _ GameInvitationChestConfigsDao = (*gameInvitationChestConfigsDao)(nil)

// GameInvitationChestConfigsDao defining the dao interface
type GameInvitationChestConfigsDao interface {
	Create(ctx context.Context, table *model.GameInvitationChestConfigs) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameInvitationChestConfigs) error
	GetByID(ctx context.Context, id uint64) (*model.GameInvitationChestConfigs, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameInvitationChestConfigs, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvitationChestConfigs) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvitationChestConfigs) error
}

type gameInvitationChestConfigsDao struct {
	db    *gorm.DB
	cache cache.GameInvitationChestConfigsCache // if nil, the cache is not used.
	sfg   *singleflight.Group                   // if cache is nil, the sfg is not used.
}

// NewGameInvitationChestConfigsDao creating the dao interface
func NewGameInvitationChestConfigsDao(db *gorm.DB, xCache cache.GameInvitationChestConfigsCache) GameInvitationChestConfigsDao {
	if xCache == nil {
		return &gameInvitationChestConfigsDao{db: db}
	}
	return &gameInvitationChestConfigsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameInvitationChestConfigsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameInvitationChestConfigs, insert the record and the id value is written back to the table
func (d *gameInvitationChestConfigsDao) Create(ctx context.Context, table *model.GameInvitationChestConfigs) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameInvitationChestConfigs by id
func (d *gameInvitationChestConfigsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameInvitationChestConfigs{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameInvitationChestConfigs by id, support partial update
func (d *gameInvitationChestConfigsDao) UpdateByID(ctx context.Context, table *model.GameInvitationChestConfigs) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameInvitationChestConfigsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameInvitationChestConfigs) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.ChestCode != "" {
		update["chest_code"] = table.ChestCode
	}
	if table.Title != "" {
		update["title"] = table.Title
	}
	if table.RequiredInvites != 0 {
		update["required_invites"] = table.RequiredInvites
	}
	if table.RewardCash != nil && table.RewardCash.IsZero() == false {
		update["reward_cash"] = table.RewardCash
	}
	if table.RewardBonus != nil && table.RewardBonus.IsZero() == false {
		update["reward_bonus"] = table.RewardBonus
	}
	if table.IconURL != "" {
		update["icon_url"] = table.IconURL
	}
	if table.Sort != 0 {
		update["sort"] = table.Sort
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameInvitationChestConfigs by id
func (d *gameInvitationChestConfigsDao) GetByID(ctx context.Context, id uint64) (*model.GameInvitationChestConfigs, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameInvitationChestConfigs{}
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
			table := &model.GameInvitationChestConfigs{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameInvitationChestConfigsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameInvitationChestConfigs)
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

// GetByColumns get a paginated list of gameInvitationChestConfigss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameInvitationChestConfigsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameInvitationChestConfigs, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameInvitationChestConfigsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameInvitationChestConfigs{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameInvitationChestConfigs{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameInvitationChestConfigsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvitationChestConfigs) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameInvitationChestConfigsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameInvitationChestConfigs{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameInvitationChestConfigsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvitationChestConfigs) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
