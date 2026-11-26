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

var _ GameWithdrawOrdersHandler = (*gameWithdrawOrdersHandler)(nil)

// GameWithdrawOrdersHandler defining the handler interface
type GameWithdrawOrdersHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	CreateWalletWithdraw(c *gin.Context)
}

type gameWithdrawOrdersHandler struct {
	iDao dao.GameWithdrawOrdersDao
}

// NewGameWithdrawOrdersHandler creating the handler interface
func NewGameWithdrawOrdersHandler() GameWithdrawOrdersHandler {
	return &gameWithdrawOrdersHandler{
		iDao: dao.NewGameWithdrawOrdersDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameWithdrawOrdersCache(database.GetCacheType()),
		),
	}
}

// Create a new gameWithdrawOrders
// @Summary Create a new gameWithdrawOrders
// @Description Creates a new gameWithdrawOrders entity using the provided data in the request body.
// @Tags gameWithdrawOrders
// @Accept json
// @Produce json
// @Param data body types.CreateGameWithdrawOrdersRequest true "gameWithdrawOrders information"
// @Success 200 {object} types.CreateGameWithdrawOrdersReply{}
// @Router /api/v1/gameWithdrawOrders [post]
// @Security BearerAuth
func (h *gameWithdrawOrdersHandler) Create(c *gin.Context) {
	form := &types.CreateGameWithdrawOrdersRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameWithdrawOrders := &model.GameWithdrawOrders{}
	err = copier.Copy(gameWithdrawOrders, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameWithdrawOrders)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameWithdrawOrders)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameWithdrawOrders.ID})
}

// DeleteByID delete a gameWithdrawOrders by id
// @Summary Delete a gameWithdrawOrders by id
// @Description Deletes a existing gameWithdrawOrders identified by the given id in the path.
// @Tags gameWithdrawOrders
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameWithdrawOrdersByIDReply{}
// @Router /api/v1/gameWithdrawOrders/{id} [delete]
// @Security BearerAuth
func (h *gameWithdrawOrdersHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameWithdrawOrdersIDFromPath(c)
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

// UpdateByID update a gameWithdrawOrders by id
// @Summary Update a gameWithdrawOrders by id
// @Description Updates the specified gameWithdrawOrders by given id in the path, support partial update.
// @Tags gameWithdrawOrders
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameWithdrawOrdersByIDRequest true "gameWithdrawOrders information"
// @Success 200 {object} types.UpdateGameWithdrawOrdersByIDReply{}
// @Router /api/v1/gameWithdrawOrders/{id} [put]
// @Security BearerAuth
func (h *gameWithdrawOrdersHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameWithdrawOrdersIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameWithdrawOrdersByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameWithdrawOrders := &model.GameWithdrawOrders{}
	err = copier.Copy(gameWithdrawOrders, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameWithdrawOrders)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameWithdrawOrders)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameWithdrawOrders by id
// @Summary Get a gameWithdrawOrders by id
// @Description Gets detailed information of a gameWithdrawOrders specified by the given id in the path.
// @Tags gameWithdrawOrders
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameWithdrawOrdersByIDReply{}
// @Router /api/v1/gameWithdrawOrders/{id} [get]
// @Security BearerAuth
func (h *gameWithdrawOrdersHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameWithdrawOrdersIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameWithdrawOrders, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameWithdrawOrdersObjDetail{}
	err = copier.Copy(data, gameWithdrawOrders)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameWithdrawOrders)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameWithdrawOrders": data})
}

// List get a paginated list of gameWithdrawOrderss by custom conditions
// @Summary Get a paginated list of gameWithdrawOrderss by custom conditions
// @Description Returns a paginated list of gameWithdrawOrders based on query filters, including page number and size.
// @Tags gameWithdrawOrders
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameWithdrawOrderssReply{}
// @Router /api/v1/gameWithdrawOrders/list [post]
// @Security BearerAuth
func (h *gameWithdrawOrdersHandler) List(c *gin.Context) {
	form := &types.ListGameWithdrawOrderssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameWithdrawOrderss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameWithdrawOrderss(gameWithdrawOrderss)
	if err != nil {
		response.Error(c, ecode.ErrListGameWithdrawOrders)
		return
	}

	response.Success(c, gin.H{
		"gameWithdrawOrderss": data,
		"total":               total,
	})
}

func getGameWithdrawOrdersIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameWithdrawOrders(gameWithdrawOrders *model.GameWithdrawOrders) (*types.GameWithdrawOrdersObjDetail, error) {
	data := &types.GameWithdrawOrdersObjDetail{}
	err := copier.Copy(data, gameWithdrawOrders)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameWithdrawOrderss(fromValues []*model.GameWithdrawOrders) ([]*types.GameWithdrawOrdersObjDetail, error) {
	toValues := []*types.GameWithdrawOrdersObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameWithdrawOrders(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
