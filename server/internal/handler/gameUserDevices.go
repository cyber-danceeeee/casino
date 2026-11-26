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

var _ GameUserDevicesHandler = (*gameUserDevicesHandler)(nil)

// GameUserDevicesHandler defining the handler interface
type GameUserDevicesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameUserDevicesHandler struct {
	iDao dao.GameUserDevicesDao
}

// NewGameUserDevicesHandler creating the handler interface
func NewGameUserDevicesHandler() GameUserDevicesHandler {
	return &gameUserDevicesHandler{
		iDao: dao.NewGameUserDevicesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameUserDevicesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameUserDevices
// @Summary Create a new gameUserDevices
// @Description Creates a new gameUserDevices entity using the provided data in the request body.
// @Tags gameUserDevices
// @Accept json
// @Produce json
// @Param data body types.CreateGameUserDevicesRequest true "gameUserDevices information"
// @Success 200 {object} types.CreateGameUserDevicesReply{}
// @Router /api/v1/gameUserDevices [post]
// @Security BearerAuth
func (h *gameUserDevicesHandler) Create(c *gin.Context) {
	form := &types.CreateGameUserDevicesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameUserDevices := &model.GameUserDevices{}
	err = copier.Copy(gameUserDevices, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameUserDevices)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameUserDevices)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameUserDevices.ID})
}

// DeleteByID delete a gameUserDevices by id
// @Summary Delete a gameUserDevices by id
// @Description Deletes a existing gameUserDevices identified by the given id in the path.
// @Tags gameUserDevices
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameUserDevicesByIDReply{}
// @Router /api/v1/gameUserDevices/{id} [delete]
// @Security BearerAuth
func (h *gameUserDevicesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameUserDevicesIDFromPath(c)
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

// UpdateByID update a gameUserDevices by id
// @Summary Update a gameUserDevices by id
// @Description Updates the specified gameUserDevices by given id in the path, support partial update.
// @Tags gameUserDevices
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameUserDevicesByIDRequest true "gameUserDevices information"
// @Success 200 {object} types.UpdateGameUserDevicesByIDReply{}
// @Router /api/v1/gameUserDevices/{id} [put]
// @Security BearerAuth
func (h *gameUserDevicesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameUserDevicesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameUserDevicesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameUserDevices := &model.GameUserDevices{}
	err = copier.Copy(gameUserDevices, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameUserDevices)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameUserDevices)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameUserDevices by id
// @Summary Get a gameUserDevices by id
// @Description Gets detailed information of a gameUserDevices specified by the given id in the path.
// @Tags gameUserDevices
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameUserDevicesByIDReply{}
// @Router /api/v1/gameUserDevices/{id} [get]
// @Security BearerAuth
func (h *gameUserDevicesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameUserDevicesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserDevices, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameUserDevicesObjDetail{}
	err = copier.Copy(data, gameUserDevices)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameUserDevices)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameUserDevices": data})
}

// List get a paginated list of gameUserDevicess by custom conditions
// @Summary Get a paginated list of gameUserDevicess by custom conditions
// @Description Returns a paginated list of gameUserDevices based on query filters, including page number and size.
// @Tags gameUserDevices
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameUserDevicessReply{}
// @Router /api/v1/gameUserDevices/list [post]
// @Security BearerAuth
func (h *gameUserDevicesHandler) List(c *gin.Context) {
	form := &types.ListGameUserDevicessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserDevicess, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameUserDevicess(gameUserDevicess)
	if err != nil {
		response.Error(c, ecode.ErrListGameUserDevices)
		return
	}

	response.Success(c, gin.H{
		"gameUserDevicess": data,
		"total":        total,
	})
}

func getGameUserDevicesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameUserDevices(gameUserDevices *model.GameUserDevices) (*types.GameUserDevicesObjDetail, error) {
	data := &types.GameUserDevicesObjDetail{}
	err := copier.Copy(data, gameUserDevices)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameUserDevicess(fromValues []*model.GameUserDevices) ([]*types.GameUserDevicesObjDetail, error) {
	toValues := []*types.GameUserDevicesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameUserDevices(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
