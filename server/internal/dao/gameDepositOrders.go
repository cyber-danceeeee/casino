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

var _ GameDepositOrdersDao = (*gameDepositOrdersDao)(nil)

// GameDepositOrdersDao defining the dao interface
type GameDepositOrdersDao interface {
	Create(ctx context.Context, table *model.GameDepositOrders) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameDepositOrders) error
	GetByID(ctx context.Context, id uint64) (*model.GameDepositOrders, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameDepositOrders, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameDepositOrders) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameDepositOrders) error
}

type gameDepositOrdersDao struct {
	db    *gorm.DB
	cache cache.GameDepositOrdersCache // if nil, the cache is not used.
	sfg   *singleflight.Group          // if cache is nil, the sfg is not used.
}

// NewGameDepositOrdersDao creating the dao interface
func NewGameDepositOrdersDao(db *gorm.DB, xCache cache.GameDepositOrdersCache) GameDepositOrdersDao {
	if xCache == nil {
		return &gameDepositOrdersDao{db: db}
	}
	return &gameDepositOrdersDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameDepositOrdersDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameDepositOrders, insert the record and the id value is written back to the table
func (d *gameDepositOrdersDao) Create(ctx context.Context, table *model.GameDepositOrders) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameDepositOrders by id
func (d *gameDepositOrdersDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameDepositOrders{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameDepositOrders by id, support partial update
func (d *gameDepositOrdersDao) UpdateByID(ctx context.Context, table *model.GameDepositOrders) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameDepositOrdersDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameDepositOrders) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.OrderNo != "" {
		update["order_no"] = table.OrderNo
	}
	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.Amount != nil && table.Amount.IsZero() == false {
		update["amount"] = table.Amount
	}
	if table.PaidAmount != nil && table.PaidAmount.IsZero() == false {
		update["paid_amount"] = table.PaidAmount
	}
	if table.Currency != "" {
		update["currency"] = table.Currency
	}
	if table.MethodCode != "" {
		update["method_code"] = table.MethodCode
	}
	if table.ChannelCode != "" {
		update["channel_code"] = table.ChannelCode
	}
	if table.VoucherID != 0 {
		update["voucher_id"] = table.VoucherID
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.PaidAt != nil && table.PaidAt.IsZero() == false {
		update["paid_at"] = table.PaidAt
	}
	if table.ExpiredAt != nil && table.ExpiredAt.IsZero() == false {
		update["expired_at"] = table.ExpiredAt
	}
	if table.FailReason != "" {
		update["fail_reason"] = table.FailReason
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameDepositOrders by id
func (d *gameDepositOrdersDao) GetByID(ctx context.Context, id uint64) (*model.GameDepositOrders, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameDepositOrders{}
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
			table := &model.GameDepositOrders{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameDepositOrdersExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameDepositOrders)
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

// GetByColumns get a paginated list of gameDepositOrderss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameDepositOrdersDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameDepositOrders, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameDepositOrdersColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameDepositOrders{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameDepositOrders{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameDepositOrdersDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameDepositOrders) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameDepositOrdersDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameDepositOrders{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameDepositOrdersDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameDepositOrders) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
