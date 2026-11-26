package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/go-dev-frame/sponge/pkg/copier"
	"github.com/go-dev-frame/sponge/pkg/gin/middleware"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"casino/internal/cache"
	"casino/internal/dao"
	"casino/internal/database"
	"casino/internal/ecode"
	"casino/internal/model"
	"casino/internal/types"
)

var _ GameWalletTransactionsHandler = (*gameWalletTransactionsHandler)(nil)

// GameWalletTransactionsHandler defining the handler interface
type GameWalletTransactionsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	WalletRecords(c *gin.Context)
}

type gameWalletTransactionsHandler struct {
	iDao dao.GameWalletTransactionsDao
}

// NewGameWalletTransactionsHandler creating the handler interface
func NewGameWalletTransactionsHandler() GameWalletTransactionsHandler {
	return &gameWalletTransactionsHandler{
		iDao: dao.NewGameWalletTransactionsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameWalletTransactionsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameWalletTransactions
// @Summary Create a new gameWalletTransactions
// @Description Creates a new gameWalletTransactions entity using the provided data in the request body.
// @Tags gameWalletTransactions
// @Accept json
// @Produce json
// @Param data body types.CreateGameWalletTransactionsRequest true "gameWalletTransactions information"
// @Success 200 {object} types.CreateGameWalletTransactionsReply{}
// @Router /api/v1/gameWalletTransactions [post]
// @Security BearerAuth
func (h *gameWalletTransactionsHandler) Create(c *gin.Context) {
	form := &types.CreateGameWalletTransactionsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameWalletTransactions := &model.GameWalletTransactions{}
	err = copier.Copy(gameWalletTransactions, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameWalletTransactions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameWalletTransactions)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameWalletTransactions.ID})
}

// DeleteByID delete a gameWalletTransactions by id
// @Summary Delete a gameWalletTransactions by id
// @Description Deletes a existing gameWalletTransactions identified by the given id in the path.
// @Tags gameWalletTransactions
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameWalletTransactionsByIDReply{}
// @Router /api/v1/gameWalletTransactions/{id} [delete]
// @Security BearerAuth
func (h *gameWalletTransactionsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameWalletTransactionsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	err := h.iDao.DeleteByID(ctx, id)
	if err != nil {
		logger.Error("DeleteByID error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// UpdateByID update a gameWalletTransactions by id
// @Summary Update a gameWalletTransactions by id
// @Description Updates the specified gameWalletTransactions by given id in the path, support partial update.
// @Tags gameWalletTransactions
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameWalletTransactionsByIDRequest true "gameWalletTransactions information"
// @Success 200 {object} types.UpdateGameWalletTransactionsByIDReply{}
// @Router /api/v1/gameWalletTransactions/{id} [put]
// @Security BearerAuth
func (h *gameWalletTransactionsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameWalletTransactionsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameWalletTransactionsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameWalletTransactions := &model.GameWalletTransactions{}
	err = copier.Copy(gameWalletTransactions, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameWalletTransactions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameWalletTransactions)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameWalletTransactions by id
// @Summary Get a gameWalletTransactions by id
// @Description Gets detailed information of a gameWalletTransactions specified by the given id in the path.
// @Tags gameWalletTransactions
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameWalletTransactionsByIDReply{}
// @Router /api/v1/gameWalletTransactions/{id} [get]
// @Security BearerAuth
func (h *gameWalletTransactionsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameWalletTransactionsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameWalletTransactions, err := h.iDao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			logger.Warn("GetByID not found", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
			response.Error(c, ecode.NotFound)
		} else {
			logger.Error("GetByID error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
			response.Output(c, ecode.InternalServerError.ToHTTPCode())
		}
		return
	}

	data := &types.GameWalletTransactionsObjDetail{}
	err = copier.Copy(data, gameWalletTransactions)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameWalletTransactions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameWalletTransactions": data})
}

// List get a paginated list of gameWalletTransactionss by custom conditions
// @Summary Get a paginated list of gameWalletTransactionss by custom conditions
// @Description Returns a paginated list of gameWalletTransactions based on query filters, including page number and size.
// @Tags gameWalletTransactions
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameWalletTransactionssReply{}
// @Router /api/v1/gameWalletTransactions/list [post]
// @Security BearerAuth
func (h *gameWalletTransactionsHandler) List(c *gin.Context) {
	form := &types.ListGameWalletTransactionssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameWalletTransactionss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameWalletTransactionss(gameWalletTransactionss)
	if err != nil {
		response.Error(c, ecode.ErrListGameWalletTransactions)
		return
	}

	response.Success(c, gin.H{
		"gameWalletTransactionss": data,
		"total":                   total,
	})
}

func getGameWalletTransactionsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameWalletTransactions(gameWalletTransactions *model.GameWalletTransactions) (*types.GameWalletTransactionsObjDetail, error) {
	data := &types.GameWalletTransactionsObjDetail{}
	err := copier.Copy(data, gameWalletTransactions)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameWalletTransactionss(fromValues []*model.GameWalletTransactions) ([]*types.GameWalletTransactionsObjDetail, error) {
	toValues := []*types.GameWalletTransactionsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameWalletTransactions(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
