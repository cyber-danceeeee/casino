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

var _ GameSupportTicketsDao = (*gameSupportTicketsDao)(nil)

// GameSupportTicketsDao defining the dao interface
type GameSupportTicketsDao interface {
	Create(ctx context.Context, table *model.GameSupportTickets) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameSupportTickets) error
	GetByID(ctx context.Context, id uint64) (*model.GameSupportTickets, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameSupportTickets, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSupportTickets) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSupportTickets) error
}

type gameSupportTicketsDao struct {
	db    *gorm.DB
	cache cache.GameSupportTicketsCache // if nil, the cache is not used.
	sfg   *singleflight.Group           // if cache is nil, the sfg is not used.
}

// NewGameSupportTicketsDao creating the dao interface
func NewGameSupportTicketsDao(db *gorm.DB, xCache cache.GameSupportTicketsCache) GameSupportTicketsDao {
	if xCache == nil {
		return &gameSupportTicketsDao{db: db}
	}
	return &gameSupportTicketsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameSupportTicketsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameSupportTickets, insert the record and the id value is written back to the table
func (d *gameSupportTicketsDao) Create(ctx context.Context, table *model.GameSupportTickets) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameSupportTickets by id
func (d *gameSupportTicketsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameSupportTickets{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameSupportTickets by id, support partial update
func (d *gameSupportTicketsDao) UpdateByID(ctx context.Context, table *model.GameSupportTickets) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameSupportTicketsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameSupportTickets) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.TicketNo != "" {
		update["ticket_no"] = table.TicketNo
	}
	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.CategoryCode != "" {
		update["category_code"] = table.CategoryCode
	}
	if table.Subject != "" {
		update["subject"] = table.Subject
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.LastMessageAt != nil && table.LastMessageAt.IsZero() == false {
		update["last_message_at"] = table.LastMessageAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameSupportTickets by id
func (d *gameSupportTicketsDao) GetByID(ctx context.Context, id uint64) (*model.GameSupportTickets, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameSupportTickets{}
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
			table := &model.GameSupportTickets{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameSupportTicketsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameSupportTickets)
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

// GetByColumns get a paginated list of gameSupportTicketss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameSupportTicketsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameSupportTickets, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameSupportTicketsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameSupportTickets{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameSupportTickets{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameSupportTicketsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSupportTickets) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameSupportTicketsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameSupportTickets{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameSupportTicketsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameSupportTickets) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
