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

var _ GameUsersDao = (*gameUsersDao)(nil)

// GameUsersDao defining the dao interface
type GameUsersDao interface {
	Create(ctx context.Context, table *model.GameUsers) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameUsers) error
	GetByID(ctx context.Context, id uint64) (*model.GameUsers, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUsers, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUsers) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUsers) error
}

type gameUsersDao struct {
	db    *gorm.DB
	cache cache.GameUsersCache // if nil, the cache is not used.
	sfg   *singleflight.Group  // if cache is nil, the sfg is not used.
}

// NewGameUsersDao creating the dao interface
func NewGameUsersDao(db *gorm.DB, xCache cache.GameUsersCache) GameUsersDao {
	if xCache == nil {
		return &gameUsersDao{db: db}
	}
	return &gameUsersDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameUsersDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameUsers, insert the record and the id value is written back to the table
func (d *gameUsersDao) Create(ctx context.Context, table *model.GameUsers) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameUsers by id
func (d *gameUsersDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUsers{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameUsers by id, support partial update
func (d *gameUsersDao) UpdateByID(ctx context.Context, table *model.GameUsers) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameUsersDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameUsers) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.UserNo != "" {
		update["user_no"] = table.UserNo
	}
	if table.PhoneCountryCode != "" {
		update["phone_country_code"] = table.PhoneCountryCode
	}
	if table.Phone != "" {
		update["phone"] = table.Phone
	}
	if table.PasswordHash != "" {
		update["password_hash"] = table.PasswordHash
	}
	if table.Nickname != "" {
		update["nickname"] = table.Nickname
	}
	if table.AvatarURL != "" {
		update["avatar_url"] = table.AvatarURL
	}
	if table.InviteCode != "" {
		update["invite_code"] = table.InviteCode
	}
	if table.ReferrerUserID != 0 {
		update["referrer_user_id"] = table.ReferrerUserID
	}
	if table.VipLevel != 0 {
		update["vip_level"] = table.VipLevel
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.RegisteredIP != "" {
		update["registered_ip"] = table.RegisteredIP
	}
	if table.LastLoginAt != nil && table.LastLoginAt.IsZero() == false {
		update["last_login_at"] = table.LastLoginAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameUsers by id
func (d *gameUsersDao) GetByID(ctx context.Context, id uint64) (*model.GameUsers, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameUsers{}
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
			table := &model.GameUsers{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameUsersExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameUsers)
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

// GetByColumns get a paginated list of gameUserss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameUsersDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUsers, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameUsersColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameUsers{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameUsers{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameUsersDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUsers) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameUsersDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUsers{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameUsersDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUsers) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
