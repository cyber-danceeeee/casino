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

var _ GameGameLaunchLogsHandler = (*gameGameLaunchLogsHandler)(nil)

// GameGameLaunchLogsHandler defining the handler interface
type GameGameLaunchLogsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameGameLaunchLogsHandler struct {
	iDao dao.GameGameLaunchLogsDao
}

// NewGameGameLaunchLogsHandler creating the handler interface
func NewGameGameLaunchLogsHandler() GameGameLaunchLogsHandler {
	return &gameGameLaunchLogsHandler{
		iDao: dao.NewGameGameLaunchLogsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameGameLaunchLogsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameGameLaunchLogs
// @Summary Create a new gameGameLaunchLogs
// @Description Creates a new gameGameLaunchLogs entity using the provided data in the request body.
// @Tags gameGameLaunchLogs
// @Accept json
// @Produce json
// @Param data body types.CreateGameGameLaunchLogsRequest true "gameGameLaunchLogs information"
// @Success 200 {object} types.CreateGameGameLaunchLogsReply{}
// @Router /api/v1/gameGameLaunchLogs [post]
// @Security BearerAuth
func (h *gameGameLaunchLogsHandler) Create(c *gin.Context) {
	form := &types.CreateGameGameLaunchLogsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameGameLaunchLogs := &model.GameGameLaunchLogs{}
	err = copier.Copy(gameGameLaunchLogs, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameGameLaunchLogs)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameGameLaunchLogs)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameGameLaunchLogs.ID})
}

// DeleteByID delete a gameGameLaunchLogs by id
// @Summary Delete a gameGameLaunchLogs by id
// @Description Deletes a existing gameGameLaunchLogs identified by the given id in the path.
// @Tags gameGameLaunchLogs
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameGameLaunchLogsByIDReply{}
// @Router /api/v1/gameGameLaunchLogs/{id} [delete]
// @Security BearerAuth
func (h *gameGameLaunchLogsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameGameLaunchLogsIDFromPath(c)
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

// UpdateByID update a gameGameLaunchLogs by id
// @Summary Update a gameGameLaunchLogs by id
// @Description Updates the specified gameGameLaunchLogs by given id in the path, support partial update.
// @Tags gameGameLaunchLogs
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameGameLaunchLogsByIDRequest true "gameGameLaunchLogs information"
// @Success 200 {object} types.UpdateGameGameLaunchLogsByIDReply{}
// @Router /api/v1/gameGameLaunchLogs/{id} [put]
// @Security BearerAuth
func (h *gameGameLaunchLogsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameGameLaunchLogsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameGameLaunchLogsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameGameLaunchLogs := &model.GameGameLaunchLogs{}
	err = copier.Copy(gameGameLaunchLogs, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameGameLaunchLogs)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameGameLaunchLogs)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameGameLaunchLogs by id
// @Summary Get a gameGameLaunchLogs by id
// @Description Gets detailed information of a gameGameLaunchLogs specified by the given id in the path.
// @Tags gameGameLaunchLogs
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameGameLaunchLogsByIDReply{}
// @Router /api/v1/gameGameLaunchLogs/{id} [get]
// @Security BearerAuth
func (h *gameGameLaunchLogsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameGameLaunchLogsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameGameLaunchLogs, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameGameLaunchLogsObjDetail{}
	err = copier.Copy(data, gameGameLaunchLogs)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameGameLaunchLogs)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameGameLaunchLogs": data})
}

// List get a paginated list of gameGameLaunchLogss by custom conditions
// @Summary Get a paginated list of gameGameLaunchLogss by custom conditions
// @Description Returns a paginated list of gameGameLaunchLogs based on query filters, including page number and size.
// @Tags gameGameLaunchLogs
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameGameLaunchLogssReply{}
// @Router /api/v1/gameGameLaunchLogs/list [post]
// @Security BearerAuth
func (h *gameGameLaunchLogsHandler) List(c *gin.Context) {
	form := &types.ListGameGameLaunchLogssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameGameLaunchLogss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameGameLaunchLogss(gameGameLaunchLogss)
	if err != nil {
		response.Error(c, ecode.ErrListGameGameLaunchLogs)
		return
	}

	response.Success(c, gin.H{
		"gameGameLaunchLogss": data,
		"total":               total,
	})
}

func getGameGameLaunchLogsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameGameLaunchLogs(gameGameLaunchLogs *model.GameGameLaunchLogs) (*types.GameGameLaunchLogsObjDetail, error) {
	data := &types.GameGameLaunchLogsObjDetail{}
	err := copier.Copy(data, gameGameLaunchLogs)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameGameLaunchLogss(fromValues []*model.GameGameLaunchLogs) ([]*types.GameGameLaunchLogsObjDetail, error) {
	toValues := []*types.GameGameLaunchLogsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameGameLaunchLogs(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
