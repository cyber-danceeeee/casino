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

var _ GameWithdrawOrdersDao = (*gameWithdrawOrdersDao)(nil)

// GameWithdrawOrdersDao defining the dao interface
type GameWithdrawOrdersDao interface {
	Create(ctx context.Context, table *model.GameWithdrawOrders) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameWithdrawOrders) error
	GetByID(ctx context.Context, id uint64) (*model.GameWithdrawOrders, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameWithdrawOrders, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameWithdrawOrders) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameWithdrawOrders) error
}

type gameWithdrawOrdersDao struct {
	db    *gorm.DB
	cache cache.GameWithdrawOrdersCache // if nil, the cache is not used.
	sfg   *singleflight.Group           // if cache is nil, the sfg is not used.
}

// NewGameWithdrawOrdersDao creating the dao interface
func NewGameWithdrawOrdersDao(db *gorm.DB, xCache cache.GameWithdrawOrdersCache) GameWithdrawOrdersDao {
	if xCache == nil {
		return &gameWithdrawOrdersDao{db: db}
	}
	return &gameWithdrawOrdersDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameWithdrawOrdersDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameWithdrawOrders, insert the record and the id value is written back to the table
func (d *gameWithdrawOrdersDao) Create(ctx context.Context, table *model.GameWithdrawOrders) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameWithdrawOrders by id
func (d *gameWithdrawOrdersDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameWithdrawOrders{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameWithdrawOrders by id, support partial update
func (d *gameWithdrawOrdersDao) UpdateByID(ctx context.Context, table *model.GameWithdrawOrders) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameWithdrawOrdersDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameWithdrawOrders) error {
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
	if table.WithdrawAccountID != 0 {
		update["withdraw_account_id"] = table.WithdrawAccountID
	}
	if table.Amount != nil && table.Amount.IsZero() == false {
		update["amount"] = table.Amount
	}
	if table.FeeAmount != nil && table.FeeAmount.IsZero() == false {
		update["fee_amount"] = table.FeeAmount
	}
	if table.ActualAmount != nil && table.ActualAmount.IsZero() == false {
		update["actual_amount"] = table.ActualAmount
	}
	if table.Currency != "" {
		update["currency"] = table.Currency
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	if table.AuditStatus != 0 {
		update["audit_status"] = table.AuditStatus
	}
	if table.ReviewRemark != "" {
		update["review_remark"] = table.ReviewRemark
	}
	if table.SubmittedAt != nil && table.SubmittedAt.IsZero() == false {
		update["submitted_at"] = table.SubmittedAt
	}
	if table.CompletedAt != nil && table.CompletedAt.IsZero() == false {
		update["completed_at"] = table.CompletedAt
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameWithdrawOrders by id
func (d *gameWithdrawOrdersDao) GetByID(ctx context.Context, id uint64) (*model.GameWithdrawOrders, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameWithdrawOrders{}
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
			table := &model.GameWithdrawOrders{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameWithdrawOrdersExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameWithdrawOrders)
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

// GetByColumns get a paginated list of gameWithdrawOrderss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameWithdrawOrdersDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameWithdrawOrders, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameWithdrawOrdersColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameWithdrawOrders{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameWithdrawOrders{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameWithdrawOrdersDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameWithdrawOrders) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameWithdrawOrdersDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameWithdrawOrders{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameWithdrawOrdersDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameWithdrawOrders) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
