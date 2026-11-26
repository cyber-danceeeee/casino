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

var _ GameUserSessionsDao = (*gameUserSessionsDao)(nil)

// GameUserSessionsDao defining the dao interface
type GameUserSessionsDao interface {
	Create(ctx context.Context, table *model.GameUserSessions) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameUserSessions) error
	GetByID(ctx context.Context, id uint64) (*model.GameUserSessions, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUserSessions, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserSessions) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserSessions) error
}

type gameUserSessionsDao struct {
	db    *gorm.DB
	cache cache.GameUserSessionsCache // if nil, the cache is not used.
	sfg   *singleflight.Group         // if cache is nil, the sfg is not used.
}

// NewGameUserSessionsDao creating the dao interface
func NewGameUserSessionsDao(db *gorm.DB, xCache cache.GameUserSessionsCache) GameUserSessionsDao {
	if xCache == nil {
		return &gameUserSessionsDao{db: db}
	}
	return &gameUserSessionsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameUserSessionsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameUserSessions, insert the record and the id value is written back to the table
func (d *gameUserSessionsDao) Create(ctx context.Context, table *model.GameUserSessions) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameUserSessions by id
func (d *gameUserSessionsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUserSessions{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameUserSessions by id, support partial update
func (d *gameUserSessionsDao) UpdateByID(ctx context.Context, table *model.GameUserSessions) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameUserSessionsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameUserSessions) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.TokenHash != "" {
		update["token_hash"] = table.TokenHash
	}
	if table.DeviceID != "" {
		update["device_id"] = table.DeviceID
	}
	if table.ClientIP != "" {
		update["client_ip"] = table.ClientIP
	}
	if table.UserAgent != "" {
		update["user_agent"] = table.UserAgent
	}
	if table.ExpiredAt != nil && table.ExpiredAt.IsZero() == false {
		update["expired_at"] = table.ExpiredAt
	}
	if table.RevokedAt != nil && table.RevokedAt.IsZero() == false {
		update["revoked_at"] = table.RevokedAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameUserSessions by id
func (d *gameUserSessionsDao) GetByID(ctx context.Context, id uint64) (*model.GameUserSessions, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameUserSessions{}
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
			table := &model.GameUserSessions{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameUserSessionsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameUserSessions)
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

// GetByColumns get a paginated list of gameUserSessionss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameUserSessionsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameUserSessions, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameUserSessionsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameUserSessions{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameUserSessions{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameUserSessionsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserSessions) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameUserSessionsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameUserSessions{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameUserSessionsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameUserSessions) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
