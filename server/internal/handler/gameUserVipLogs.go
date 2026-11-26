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

var _ GameUserVipLogsHandler = (*gameUserVipLogsHandler)(nil)

// GameUserVipLogsHandler defining the handler interface
type GameUserVipLogsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameUserVipLogsHandler struct {
	iDao dao.GameUserVipLogsDao
}

// NewGameUserVipLogsHandler creating the handler interface
func NewGameUserVipLogsHandler() GameUserVipLogsHandler {
	return &gameUserVipLogsHandler{
		iDao: dao.NewGameUserVipLogsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameUserVipLogsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameUserVipLogs
// @Summary Create a new gameUserVipLogs
// @Description Creates a new gameUserVipLogs entity using the provided data in the request body.
// @Tags gameUserVipLogs
// @Accept json
// @Produce json
// @Param data body types.CreateGameUserVipLogsRequest true "gameUserVipLogs information"
// @Success 200 {object} types.CreateGameUserVipLogsReply{}
// @Router /api/v1/gameUserVipLogs [post]
// @Security BearerAuth
func (h *gameUserVipLogsHandler) Create(c *gin.Context) {
	form := &types.CreateGameUserVipLogsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameUserVipLogs := &model.GameUserVipLogs{}
	err = copier.Copy(gameUserVipLogs, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameUserVipLogs)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameUserVipLogs)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameUserVipLogs.ID})
}

// DeleteByID delete a gameUserVipLogs by id
// @Summary Delete a gameUserVipLogs by id
// @Description Deletes a existing gameUserVipLogs identified by the given id in the path.
// @Tags gameUserVipLogs
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameUserVipLogsByIDReply{}
// @Router /api/v1/gameUserVipLogs/{id} [delete]
// @Security BearerAuth
func (h *gameUserVipLogsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameUserVipLogsIDFromPath(c)
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

// UpdateByID update a gameUserVipLogs by id
// @Summary Update a gameUserVipLogs by id
// @Description Updates the specified gameUserVipLogs by given id in the path, support partial update.
// @Tags gameUserVipLogs
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameUserVipLogsByIDRequest true "gameUserVipLogs information"
// @Success 200 {object} types.UpdateGameUserVipLogsByIDReply{}
// @Router /api/v1/gameUserVipLogs/{id} [put]
// @Security BearerAuth
func (h *gameUserVipLogsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameUserVipLogsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameUserVipLogsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameUserVipLogs := &model.GameUserVipLogs{}
	err = copier.Copy(gameUserVipLogs, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameUserVipLogs)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameUserVipLogs)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameUserVipLogs by id
// @Summary Get a gameUserVipLogs by id
// @Description Gets detailed information of a gameUserVipLogs specified by the given id in the path.
// @Tags gameUserVipLogs
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameUserVipLogsByIDReply{}
// @Router /api/v1/gameUserVipLogs/{id} [get]
// @Security BearerAuth
func (h *gameUserVipLogsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameUserVipLogsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserVipLogs, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameUserVipLogsObjDetail{}
	err = copier.Copy(data, gameUserVipLogs)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameUserVipLogs)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameUserVipLogs": data})
}

// List get a paginated list of gameUserVipLogss by custom conditions
// @Summary Get a paginated list of gameUserVipLogss by custom conditions
// @Description Returns a paginated list of gameUserVipLogs based on query filters, including page number and size.
// @Tags gameUserVipLogs
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameUserVipLogssReply{}
// @Router /api/v1/gameUserVipLogs/list [post]
// @Security BearerAuth
func (h *gameUserVipLogsHandler) List(c *gin.Context) {
	form := &types.ListGameUserVipLogssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserVipLogss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameUserVipLogss(gameUserVipLogss)
	if err != nil {
		response.Error(c, ecode.ErrListGameUserVipLogs)
		return
	}

	response.Success(c, gin.H{
		"gameUserVipLogss": data,
		"total":            total,
	})
}

func getGameUserVipLogsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameUserVipLogs(gameUserVipLogs *model.GameUserVipLogs) (*types.GameUserVipLogsObjDetail, error) {
	data := &types.GameUserVipLogsObjDetail{}
	err := copier.Copy(data, gameUserVipLogs)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameUserVipLogss(fromValues []*model.GameUserVipLogs) ([]*types.GameUserVipLogsObjDetail, error) {
	toValues := []*types.GameUserVipLogsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameUserVipLogs(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
