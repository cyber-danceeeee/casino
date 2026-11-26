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
	"casino/internal/depositconfig"
	"casino/internal/ecode"
	"casino/internal/model"
	"casino/internal/types"
)

var _ GameDepositOrdersHandler = (*gameDepositOrdersHandler)(nil)

// GameDepositOrdersHandler defining the handler interface
type GameDepositOrdersHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	WalletDepositConfig(c *gin.Context)
	CreateWalletDeposit(c *gin.Context)
}

type gameDepositOrdersHandler struct {
	iDao          dao.GameDepositOrdersDao
	configService *depositconfig.Service
}

// NewGameDepositOrdersHandler creating the handler interface
func NewGameDepositOrdersHandler() GameDepositOrdersHandler {
	return &gameDepositOrdersHandler{
		iDao: dao.NewGameDepositOrdersDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameDepositOrdersCache(database.GetCacheType()),
		),
		configService: depositconfig.New(database.GetDB()),
	}
}

// Create a new gameDepositOrders
// @Summary Create a new gameDepositOrders
// @Description Creates a new gameDepositOrders entity using the provided data in the request body.
// @Tags gameDepositOrders
// @Accept json
// @Produce json
// @Param data body types.CreateGameDepositOrdersRequest true "gameDepositOrders information"
// @Success 200 {object} types.CreateGameDepositOrdersReply{}
// @Router /api/v1/gameDepositOrders [post]
// @Security BearerAuth
func (h *gameDepositOrdersHandler) Create(c *gin.Context) {
	form := &types.CreateGameDepositOrdersRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameDepositOrders := &model.GameDepositOrders{}
	err = copier.Copy(gameDepositOrders, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameDepositOrders)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameDepositOrders)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameDepositOrders.ID})
}

// DeleteByID delete a gameDepositOrders by id
// @Summary Delete a gameDepositOrders by id
// @Description Deletes a existing gameDepositOrders identified by the given id in the path.
// @Tags gameDepositOrders
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameDepositOrdersByIDReply{}
// @Router /api/v1/gameDepositOrders/{id} [delete]
// @Security BearerAuth
func (h *gameDepositOrdersHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameDepositOrdersIDFromPath(c)
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

// UpdateByID update a gameDepositOrders by id
// @Summary Update a gameDepositOrders by id
// @Description Updates the specified gameDepositOrders by given id in the path, support partial update.
// @Tags gameDepositOrders
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameDepositOrdersByIDRequest true "gameDepositOrders information"
// @Success 200 {object} types.UpdateGameDepositOrdersByIDReply{}
// @Router /api/v1/gameDepositOrders/{id} [put]
// @Security BearerAuth
func (h *gameDepositOrdersHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameDepositOrdersIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameDepositOrdersByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameDepositOrders := &model.GameDepositOrders{}
	err = copier.Copy(gameDepositOrders, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameDepositOrders)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameDepositOrders)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameDepositOrders by id
// @Summary Get a gameDepositOrders by id
// @Description Gets detailed information of a gameDepositOrders specified by the given id in the path.
// @Tags gameDepositOrders
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameDepositOrdersByIDReply{}
// @Router /api/v1/gameDepositOrders/{id} [get]
// @Security BearerAuth
func (h *gameDepositOrdersHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameDepositOrdersIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameDepositOrders, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameDepositOrdersObjDetail{}
	err = copier.Copy(data, gameDepositOrders)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameDepositOrders)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameDepositOrders": data})
}

// List get a paginated list of gameDepositOrderss by custom conditions
// @Summary Get a paginated list of gameDepositOrderss by custom conditions
// @Description Returns a paginated list of gameDepositOrders based on query filters, including page number and size.
// @Tags gameDepositOrders
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameDepositOrderssReply{}
// @Router /api/v1/gameDepositOrders/list [post]
// @Security BearerAuth
func (h *gameDepositOrdersHandler) List(c *gin.Context) {
	form := &types.ListGameDepositOrderssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameDepositOrderss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameDepositOrderss(gameDepositOrderss)
	if err != nil {
		response.Error(c, ecode.ErrListGameDepositOrders)
		return
	}

	response.Success(c, gin.H{
		"gameDepositOrderss": data,
		"total":              total,
	})
}

func getGameDepositOrdersIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameDepositOrders(gameDepositOrders *model.GameDepositOrders) (*types.GameDepositOrdersObjDetail, error) {
	data := &types.GameDepositOrdersObjDetail{}
	err := copier.Copy(data, gameDepositOrders)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameDepositOrderss(fromValues []*model.GameDepositOrders) ([]*types.GameDepositOrdersObjDetail, error) {
	toValues := []*types.GameDepositOrdersObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameDepositOrders(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
