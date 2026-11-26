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

var _ GamePaymentMethodsHandler = (*gamePaymentMethodsHandler)(nil)

// GamePaymentMethodsHandler defining the handler interface
type GamePaymentMethodsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	WalletMethods(c *gin.Context)
}

type gamePaymentMethodsHandler struct {
	iDao dao.GamePaymentMethodsDao
}

// NewGamePaymentMethodsHandler creating the handler interface
func NewGamePaymentMethodsHandler() GamePaymentMethodsHandler {
	return &gamePaymentMethodsHandler{
		iDao: dao.NewGamePaymentMethodsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGamePaymentMethodsCache(database.GetCacheType()),
		),
	}
}

// Create a new gamePaymentMethods
// @Summary Create a new gamePaymentMethods
// @Description Creates a new gamePaymentMethods entity using the provided data in the request body.
// @Tags gamePaymentMethods
// @Accept json
// @Produce json
// @Param data body types.CreateGamePaymentMethodsRequest true "gamePaymentMethods information"
// @Success 200 {object} types.CreateGamePaymentMethodsReply{}
// @Router /api/v1/gamePaymentMethods [post]
// @Security BearerAuth
func (h *gamePaymentMethodsHandler) Create(c *gin.Context) {
	form := &types.CreateGamePaymentMethodsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gamePaymentMethods := &model.GamePaymentMethods{}
	err = copier.Copy(gamePaymentMethods, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGamePaymentMethods)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gamePaymentMethods)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gamePaymentMethods.ID})
}

// DeleteByID delete a gamePaymentMethods by id
// @Summary Delete a gamePaymentMethods by id
// @Description Deletes a existing gamePaymentMethods identified by the given id in the path.
// @Tags gamePaymentMethods
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGamePaymentMethodsByIDReply{}
// @Router /api/v1/gamePaymentMethods/{id} [delete]
// @Security BearerAuth
func (h *gamePaymentMethodsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGamePaymentMethodsIDFromPath(c)
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

// UpdateByID update a gamePaymentMethods by id
// @Summary Update a gamePaymentMethods by id
// @Description Updates the specified gamePaymentMethods by given id in the path, support partial update.
// @Tags gamePaymentMethods
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGamePaymentMethodsByIDRequest true "gamePaymentMethods information"
// @Success 200 {object} types.UpdateGamePaymentMethodsByIDReply{}
// @Router /api/v1/gamePaymentMethods/{id} [put]
// @Security BearerAuth
func (h *gamePaymentMethodsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGamePaymentMethodsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGamePaymentMethodsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gamePaymentMethods := &model.GamePaymentMethods{}
	err = copier.Copy(gamePaymentMethods, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGamePaymentMethods)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gamePaymentMethods)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gamePaymentMethods by id
// @Summary Get a gamePaymentMethods by id
// @Description Gets detailed information of a gamePaymentMethods specified by the given id in the path.
// @Tags gamePaymentMethods
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGamePaymentMethodsByIDReply{}
// @Router /api/v1/gamePaymentMethods/{id} [get]
// @Security BearerAuth
func (h *gamePaymentMethodsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGamePaymentMethodsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gamePaymentMethods, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GamePaymentMethodsObjDetail{}
	err = copier.Copy(data, gamePaymentMethods)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGamePaymentMethods)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gamePaymentMethods": data})
}

// List get a paginated list of gamePaymentMethodss by custom conditions
// @Summary Get a paginated list of gamePaymentMethodss by custom conditions
// @Description Returns a paginated list of gamePaymentMethods based on query filters, including page number and size.
// @Tags gamePaymentMethods
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGamePaymentMethodssReply{}
// @Router /api/v1/gamePaymentMethods/list [post]
// @Security BearerAuth
func (h *gamePaymentMethodsHandler) List(c *gin.Context) {
	form := &types.ListGamePaymentMethodssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gamePaymentMethodss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGamePaymentMethodss(gamePaymentMethodss)
	if err != nil {
		response.Error(c, ecode.ErrListGamePaymentMethods)
		return
	}

	response.Success(c, gin.H{
		"gamePaymentMethodss": data,
		"total":               total,
	})
}

func getGamePaymentMethodsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGamePaymentMethods(gamePaymentMethods *model.GamePaymentMethods) (*types.GamePaymentMethodsObjDetail, error) {
	data := &types.GamePaymentMethodsObjDetail{}
	err := copier.Copy(data, gamePaymentMethods)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGamePaymentMethodss(fromValues []*model.GamePaymentMethods) ([]*types.GamePaymentMethodsObjDetail, error) {
	toValues := []*types.GamePaymentMethodsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGamePaymentMethods(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
