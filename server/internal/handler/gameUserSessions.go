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

var _ GameUserSessionsHandler = (*gameUserSessionsHandler)(nil)

// GameUserSessionsHandler defining the handler interface
type GameUserSessionsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameUserSessionsHandler struct {
	iDao dao.GameUserSessionsDao
}

// NewGameUserSessionsHandler creating the handler interface
func NewGameUserSessionsHandler() GameUserSessionsHandler {
	return &gameUserSessionsHandler{
		iDao: dao.NewGameUserSessionsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameUserSessionsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameUserSessions
// @Summary Create a new gameUserSessions
// @Description Creates a new gameUserSessions entity using the provided data in the request body.
// @Tags gameUserSessions
// @Accept json
// @Produce json
// @Param data body types.CreateGameUserSessionsRequest true "gameUserSessions information"
// @Success 200 {object} types.CreateGameUserSessionsReply{}
// @Router /api/v1/gameUserSessions [post]
// @Security BearerAuth
func (h *gameUserSessionsHandler) Create(c *gin.Context) {
	form := &types.CreateGameUserSessionsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameUserSessions := &model.GameUserSessions{}
	err = copier.Copy(gameUserSessions, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameUserSessions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameUserSessions)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameUserSessions.ID})
}

// DeleteByID delete a gameUserSessions by id
// @Summary Delete a gameUserSessions by id
// @Description Deletes a existing gameUserSessions identified by the given id in the path.
// @Tags gameUserSessions
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameUserSessionsByIDReply{}
// @Router /api/v1/gameUserSessions/{id} [delete]
// @Security BearerAuth
func (h *gameUserSessionsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameUserSessionsIDFromPath(c)
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

// UpdateByID update a gameUserSessions by id
// @Summary Update a gameUserSessions by id
// @Description Updates the specified gameUserSessions by given id in the path, support partial update.
// @Tags gameUserSessions
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameUserSessionsByIDRequest true "gameUserSessions information"
// @Success 200 {object} types.UpdateGameUserSessionsByIDReply{}
// @Router /api/v1/gameUserSessions/{id} [put]
// @Security BearerAuth
func (h *gameUserSessionsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameUserSessionsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameUserSessionsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameUserSessions := &model.GameUserSessions{}
	err = copier.Copy(gameUserSessions, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameUserSessions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameUserSessions)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameUserSessions by id
// @Summary Get a gameUserSessions by id
// @Description Gets detailed information of a gameUserSessions specified by the given id in the path.
// @Tags gameUserSessions
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameUserSessionsByIDReply{}
// @Router /api/v1/gameUserSessions/{id} [get]
// @Security BearerAuth
func (h *gameUserSessionsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameUserSessionsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserSessions, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameUserSessionsObjDetail{}
	err = copier.Copy(data, gameUserSessions)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameUserSessions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameUserSessions": data})
}

// List get a paginated list of gameUserSessionss by custom conditions
// @Summary Get a paginated list of gameUserSessionss by custom conditions
// @Description Returns a paginated list of gameUserSessions based on query filters, including page number and size.
// @Tags gameUserSessions
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameUserSessionssReply{}
// @Router /api/v1/gameUserSessions/list [post]
// @Security BearerAuth
func (h *gameUserSessionsHandler) List(c *gin.Context) {
	form := &types.ListGameUserSessionssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserSessionss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameUserSessionss(gameUserSessionss)
	if err != nil {
		response.Error(c, ecode.ErrListGameUserSessions)
		return
	}

	response.Success(c, gin.H{
		"gameUserSessionss": data,
		"total":             total,
	})
}

func getGameUserSessionsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameUserSessions(gameUserSessions *model.GameUserSessions) (*types.GameUserSessionsObjDetail, error) {
	data := &types.GameUserSessionsObjDetail{}
	err := copier.Copy(data, gameUserSessions)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameUserSessionss(fromValues []*model.GameUserSessions) ([]*types.GameUserSessionsObjDetail, error) {
	toValues := []*types.GameUserSessionsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameUserSessions(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
