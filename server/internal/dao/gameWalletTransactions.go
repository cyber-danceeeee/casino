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

var _ GameWalletTransactionsDao = (*gameWalletTransactionsDao)(nil)

// GameWalletTransactionsDao defining the dao interface
type GameWalletTransactionsDao interface {
	Create(ctx context.Context, table *model.GameWalletTransactions) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.GameWalletTransactions) error
	GetByID(ctx context.Context, id uint64) (*model.GameWalletTransactions, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameWalletTransactions, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameWalletTransactions) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameWalletTransactions) error
}

type gameWalletTransactionsDao struct {
	db    *gorm.DB
	cache cache.GameWalletTransactionsCache // if nil, the cache is not used.
	sfg   *singleflight.Group               // if cache is nil, the sfg is not used.
}

// NewGameWalletTransactionsDao creating the dao interface
func NewGameWalletTransactionsDao(db *gorm.DB, xCache cache.GameWalletTransactionsCache) GameWalletTransactionsDao {
	if xCache == nil {
		return &gameWalletTransactionsDao{db: db}
	}
	return &gameWalletTransactionsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *gameWalletTransactionsDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new gameWalletTransactions, insert the record and the id value is written back to the table
func (d *gameWalletTransactionsDao) Create(ctx context.Context, table *model.GameWalletTransactions) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a gameWalletTransactions by id
func (d *gameWalletTransactionsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.GameWalletTransactions{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a gameWalletTransactions by id, support partial update
func (d *gameWalletTransactionsDao) UpdateByID(ctx context.Context, table *model.GameWalletTransactions) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *gameWalletTransactionsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.GameWalletTransactions) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.TxnNo != "" {
		update["txn_no"] = table.TxnNo
	}
	if table.UserID != 0 {
		update["user_id"] = table.UserID
	}
	if table.WalletType != "" {
		update["wallet_type"] = table.WalletType
	}
	if table.Direction != "" {
		update["direction"] = table.Direction
	}
	if table.Category != "" {
		update["category"] = table.Category
	}
	if table.SourceType != "" {
		update["source_type"] = table.SourceType
	}
	if table.SourceID != 0 {
		update["source_id"] = table.SourceID
	}
	if table.OrderNo != "" {
		update["order_no"] = table.OrderNo
	}
	if table.Amount != nil && table.Amount.IsZero() == false {
		update["amount"] = table.Amount
	}
	if table.BalanceBefore != nil && table.BalanceBefore.IsZero() == false {
		update["balance_before"] = table.BalanceBefore
	}
	if table.BalanceAfter != nil && table.BalanceAfter.IsZero() == false {
		update["balance_after"] = table.BalanceAfter
	}
	if table.Remark != "" {
		update["remark"] = table.Remark
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a gameWalletTransactions by id
func (d *gameWalletTransactionsDao) GetByID(ctx context.Context, id uint64) (*model.GameWalletTransactions, error) {
	// no cache
	if d.cache == nil {
		record := &model.GameWalletTransactions{}
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
			table := &model.GameWalletTransactions{}
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
			if err = d.cache.Set(ctx, id, table, cache.GameWalletTransactionsExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.GameWalletTransactions)
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

// GetByColumns get a paginated list of gameWalletTransactionss by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *gameWalletTransactionsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.GameWalletTransactions, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.GameWalletTransactionsColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.GameWalletTransactions{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.GameWalletTransactions{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *gameWalletTransactionsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.GameWalletTransactions) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *gameWalletTransactionsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.GameWalletTransactions{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *gameWalletTransactionsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.GameWalletTransactions) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
