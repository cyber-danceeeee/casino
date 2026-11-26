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

var _ GameGameProvidersHandler = (*gameGameProvidersHandler)(nil)

// GameGameProvidersHandler defining the handler interface
type GameGameProvidersHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameGameProvidersHandler struct {
	iDao dao.GameGameProvidersDao
}

// NewGameGameProvidersHandler creating the handler interface
func NewGameGameProvidersHandler() GameGameProvidersHandler {
	return &gameGameProvidersHandler{
		iDao: dao.NewGameGameProvidersDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameGameProvidersCache(database.GetCacheType()),
		),
	}
}

// Create a new gameGameProviders
// @Summary Create a new gameGameProviders
// @Description Creates a new gameGameProviders entity using the provided data in the request body.
// @Tags gameGameProviders
// @Accept json
// @Produce json
// @Param data body types.CreateGameGameProvidersRequest true "gameGameProviders information"
// @Success 200 {object} types.CreateGameGameProvidersReply{}
// @Router /api/v1/gameGameProviders [post]
// @Security BearerAuth
func (h *gameGameProvidersHandler) Create(c *gin.Context) {
	form := &types.CreateGameGameProvidersRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameGameProviders := &model.GameGameProviders{}
	err = copier.Copy(gameGameProviders, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameGameProviders)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameGameProviders)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameGameProviders.ID})
}

// DeleteByID delete a gameGameProviders by id
// @Summary Delete a gameGameProviders by id
// @Description Deletes a existing gameGameProviders identified by the given id in the path.
// @Tags gameGameProviders
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameGameProvidersByIDReply{}
// @Router /api/v1/gameGameProviders/{id} [delete]
// @Security BearerAuth
func (h *gameGameProvidersHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameGameProvidersIDFromPath(c)
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

// UpdateByID update a gameGameProviders by id
// @Summary Update a gameGameProviders by id
// @Description Updates the specified gameGameProviders by given id in the path, support partial update.
// @Tags gameGameProviders
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameGameProvidersByIDRequest true "gameGameProviders information"
// @Success 200 {object} types.UpdateGameGameProvidersByIDReply{}
// @Router /api/v1/gameGameProviders/{id} [put]
// @Security BearerAuth
func (h *gameGameProvidersHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameGameProvidersIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameGameProvidersByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameGameProviders := &model.GameGameProviders{}
	err = copier.Copy(gameGameProviders, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameGameProviders)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameGameProviders)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameGameProviders by id
// @Summary Get a gameGameProviders by id
// @Description Gets detailed information of a gameGameProviders specified by the given id in the path.
// @Tags gameGameProviders
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameGameProvidersByIDReply{}
// @Router /api/v1/gameGameProviders/{id} [get]
// @Security BearerAuth
func (h *gameGameProvidersHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameGameProvidersIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameGameProviders, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameGameProvidersObjDetail{}
	err = copier.Copy(data, gameGameProviders)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameGameProviders)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameGameProviders": data})
}

// List get a paginated list of gameGameProviderss by custom conditions
// @Summary Get a paginated list of gameGameProviderss by custom conditions
// @Description Returns a paginated list of gameGameProviders based on query filters, including page number and size.
// @Tags gameGameProviders
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameGameProviderssReply{}
// @Router /api/v1/gameGameProviders/list [post]
// @Security BearerAuth
func (h *gameGameProvidersHandler) List(c *gin.Context) {
	form := &types.ListGameGameProviderssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameGameProviderss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameGameProviderss(gameGameProviderss)
	if err != nil {
		response.Error(c, ecode.ErrListGameGameProviders)
		return
	}

	response.Success(c, gin.H{
		"gameGameProviderss": data,
		"total":        total,
	})
}

func getGameGameProvidersIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameGameProviders(gameGameProviders *model.GameGameProviders) (*types.GameGameProvidersObjDetail, error) {
	data := &types.GameGameProvidersObjDetail{}
	err := copier.Copy(data, gameGameProviders)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameGameProviderss(fromValues []*model.GameGameProviders) ([]*types.GameGameProvidersObjDetail, error) {
	toValues := []*types.GameGameProvidersObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameGameProviders(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
