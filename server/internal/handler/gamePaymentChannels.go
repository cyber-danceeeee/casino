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

var _ GamePaymentChannelsHandler = (*gamePaymentChannelsHandler)(nil)

// GamePaymentChannelsHandler defining the handler interface
type GamePaymentChannelsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gamePaymentChannelsHandler struct {
	iDao dao.GamePaymentChannelsDao
}

// NewGamePaymentChannelsHandler creating the handler interface
func NewGamePaymentChannelsHandler() GamePaymentChannelsHandler {
	return &gamePaymentChannelsHandler{
		iDao: dao.NewGamePaymentChannelsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGamePaymentChannelsCache(database.GetCacheType()),
		),
	}
}

// Create a new gamePaymentChannels
// @Summary Create a new gamePaymentChannels
// @Description Creates a new gamePaymentChannels entity using the provided data in the request body.
// @Tags gamePaymentChannels
// @Accept json
// @Produce json
// @Param data body types.CreateGamePaymentChannelsRequest true "gamePaymentChannels information"
// @Success 200 {object} types.CreateGamePaymentChannelsReply{}
// @Router /api/v1/gamePaymentChannels [post]
// @Security BearerAuth
func (h *gamePaymentChannelsHandler) Create(c *gin.Context) {
	form := &types.CreateGamePaymentChannelsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gamePaymentChannels := &model.GamePaymentChannels{}
	err = copier.Copy(gamePaymentChannels, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGamePaymentChannels)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gamePaymentChannels)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gamePaymentChannels.ID})
}

// DeleteByID delete a gamePaymentChannels by id
// @Summary Delete a gamePaymentChannels by id
// @Description Deletes a existing gamePaymentChannels identified by the given id in the path.
// @Tags gamePaymentChannels
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGamePaymentChannelsByIDReply{}
// @Router /api/v1/gamePaymentChannels/{id} [delete]
// @Security BearerAuth
func (h *gamePaymentChannelsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGamePaymentChannelsIDFromPath(c)
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

// UpdateByID update a gamePaymentChannels by id
// @Summary Update a gamePaymentChannels by id
// @Description Updates the specified gamePaymentChannels by given id in the path, support partial update.
// @Tags gamePaymentChannels
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGamePaymentChannelsByIDRequest true "gamePaymentChannels information"
// @Success 200 {object} types.UpdateGamePaymentChannelsByIDReply{}
// @Router /api/v1/gamePaymentChannels/{id} [put]
// @Security BearerAuth
func (h *gamePaymentChannelsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGamePaymentChannelsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGamePaymentChannelsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gamePaymentChannels := &model.GamePaymentChannels{}
	err = copier.Copy(gamePaymentChannels, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGamePaymentChannels)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gamePaymentChannels)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gamePaymentChannels by id
// @Summary Get a gamePaymentChannels by id
// @Description Gets detailed information of a gamePaymentChannels specified by the given id in the path.
// @Tags gamePaymentChannels
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGamePaymentChannelsByIDReply{}
// @Router /api/v1/gamePaymentChannels/{id} [get]
// @Security BearerAuth
func (h *gamePaymentChannelsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGamePaymentChannelsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gamePaymentChannels, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GamePaymentChannelsObjDetail{}
	err = copier.Copy(data, gamePaymentChannels)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGamePaymentChannels)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gamePaymentChannels": data})
}

// List get a paginated list of gamePaymentChannelss by custom conditions
// @Summary Get a paginated list of gamePaymentChannelss by custom conditions
// @Description Returns a paginated list of gamePaymentChannels based on query filters, including page number and size.
// @Tags gamePaymentChannels
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGamePaymentChannelssReply{}
// @Router /api/v1/gamePaymentChannels/list [post]
// @Security BearerAuth
func (h *gamePaymentChannelsHandler) List(c *gin.Context) {
	form := &types.ListGamePaymentChannelssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gamePaymentChannelss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGamePaymentChannelss(gamePaymentChannelss)
	if err != nil {
		response.Error(c, ecode.ErrListGamePaymentChannels)
		return
	}

	response.Success(c, gin.H{
		"gamePaymentChannelss": data,
		"total":        total,
	})
}

func getGamePaymentChannelsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGamePaymentChannels(gamePaymentChannels *model.GamePaymentChannels) (*types.GamePaymentChannelsObjDetail, error) {
	data := &types.GamePaymentChannelsObjDetail{}
	err := copier.Copy(data, gamePaymentChannels)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGamePaymentChannelss(fromValues []*model.GamePaymentChannels) ([]*types.GamePaymentChannelsObjDetail, error) {
	toValues := []*types.GamePaymentChannelsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGamePaymentChannels(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
