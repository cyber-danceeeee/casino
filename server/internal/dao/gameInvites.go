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

var _ GameInvitesDao = (*gameInvitesDao)(nil)

// GameInvitesDao defining the dao interface
type GameInvitesDao interface {
	Create(ctx context.Context, table *model.GameInvites) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameInvites) error
	GetByID(ctx context.Context, id uint64) (*model.GameInvites, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameInvites, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvites) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvites) error
}

type gameInvitesDao struct {
	db    *gorm.DB
	cache cache.GameInvitesCache // if nil, the cache is not used.
	sfg   *singleflight.Group    // if cache is nil, the sfg is not used.
}

// NewGameInvitesDao creating the dao interface
func NewGameInvitesDao(db *gorm.DB, xCache cache.GameInvitesCache) GameInvitesDao {
	if xCache == nil {
		return &gameInvitesDao{db: db}
	}
	return &gameInvitesDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameInvitesDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameInvites, insert the record and the id value is written back to the table
func (d *gameInvitesDao) Create(ctx context.Context, table *model.GameInvites) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameInvites by id
func (d *gameInvitesDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameInvites{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameInvites by id, support partial update
func (d *gameInvitesDao) UpdateByID(ctx context.Context, table *model.GameInvites) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameInvitesDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameInvites) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.InviterUserID != 0 {
		update["inviter_user_id"] = table.InviterUserID
	}
	if table.InviteeUserID != 0 {
		update["invitee_user_id"] = table.InviteeUserID
	}
	if table.InviteCode != "" {
		update["invite_code"] = table.InviteCode
	}
	if table.InviteePhone != "" {
		update["invitee_phone"] = table.InviteePhone
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.RegisteredAt != nil && table.RegisteredAt.IsZero() == false {
		update["registered_at"] = table.RegisteredAt
	}
	if table.FirstDepositAt != nil && table.FirstDepositAt.IsZero() == false {
		update["first_deposit_at"] = table.FirstDepositAt
	}
	if table.ValidAt != nil && table.ValidAt.IsZero() == false {
		update["valid_at"] = table.ValidAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameInvites by id
func (d *gameInvitesDao) GetByID(ctx context.Context, id uint64) (*model.GameInvites, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameInvites{}
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
			table := &model.GameInvites{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameInvitesExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameInvites)
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

// GetByColumns get a paginated list of gameInvitess by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameInvitesDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameInvites, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameInvitesColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameInvites{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameInvites{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameInvitesDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvites) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameInvitesDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameInvites{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameInvitesDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameInvites) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
