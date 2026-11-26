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

var _ GameActivityEventsHandler = (*gameActivityEventsHandler)(nil)

// GameActivityEventsHandler defining the handler interface
type GameActivityEventsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Offers(c *gin.Context)
	GetByIDOrSlug(c *gin.Context)
	Action(c *gin.Context)
}

type gameActivityEventsHandler struct {
	iDao dao.GameActivityEventsDao
}

// NewGameActivityEventsHandler creating the handler interface
func NewGameActivityEventsHandler() GameActivityEventsHandler {
	return &gameActivityEventsHandler{
		iDao: dao.NewGameActivityEventsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameActivityEventsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameActivityEvents
// @Summary Create a new gameActivityEvents
// @Description Creates a new gameActivityEvents entity using the provided data in the request body.
// @Tags gameActivityEvents
// @Accept json
// @Produce json
// @Param data body types.CreateGameActivityEventsRequest true "gameActivityEvents information"
// @Success 200 {object} types.CreateGameActivityEventsReply{}
// @Router /api/v1/gameActivityEvents [post]
// @Security BearerAuth
func (h *gameActivityEventsHandler) Create(c *gin.Context) {
	form := &types.CreateGameActivityEventsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameActivityEvents := &model.GameActivityEvents{}
	err = copier.Copy(gameActivityEvents, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameActivityEvents)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameActivityEvents)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameActivityEvents.ID})
}

// DeleteByID delete a gameActivityEvents by id
// @Summary Delete a gameActivityEvents by id
// @Description Deletes a existing gameActivityEvents identified by the given id in the path.
// @Tags gameActivityEvents
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameActivityEventsByIDReply{}
// @Router /api/v1/gameActivityEvents/{id} [delete]
// @Security BearerAuth
func (h *gameActivityEventsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameActivityEventsIDFromPath(c)
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

// UpdateByID update a gameActivityEvents by id
// @Summary Update a gameActivityEvents by id
// @Description Updates the specified gameActivityEvents by given id in the path, support partial update.
// @Tags gameActivityEvents
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameActivityEventsByIDRequest true "gameActivityEvents information"
// @Success 200 {object} types.UpdateGameActivityEventsByIDReply{}
// @Router /api/v1/gameActivityEvents/{id} [put]
// @Security BearerAuth
func (h *gameActivityEventsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameActivityEventsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameActivityEventsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameActivityEvents := &model.GameActivityEvents{}
	err = copier.Copy(gameActivityEvents, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameActivityEvents)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameActivityEvents)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameActivityEvents by id
// @Summary Get a gameActivityEvents by id
// @Description Gets detailed information of a gameActivityEvents specified by the given id in the path.
// @Tags gameActivityEvents
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameActivityEventsByIDReply{}
// @Router /api/v1/gameActivityEvents/{id} [get]
// @Security BearerAuth
func (h *gameActivityEventsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameActivityEventsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameActivityEvents, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameActivityEventsObjDetail{}
	err = copier.Copy(data, gameActivityEvents)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameActivityEvents)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameActivityEvents": data})
}

// List get a paginated list of gameActivityEventss by custom conditions
// @Summary Get a paginated list of gameActivityEventss by custom conditions
// @Description Returns a paginated list of gameActivityEvents based on query filters, including page number and size.
// @Tags gameActivityEvents
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameActivityEventssReply{}
// @Router /api/v1/gameActivityEvents/list [post]
// @Security BearerAuth
func (h *gameActivityEventsHandler) List(c *gin.Context) {
	form := &types.ListGameActivityEventssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameActivityEventss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameActivityEventss(gameActivityEventss)
	if err != nil {
		response.Error(c, ecode.ErrListGameActivityEvents)
		return
	}

	response.Success(c, gin.H{
		"gameActivityEventss": data,
		"total":               total,
	})
}

func getGameActivityEventsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameActivityEvents(gameActivityEvents *model.GameActivityEvents) (*types.GameActivityEventsObjDetail, error) {
	data := &types.GameActivityEventsObjDetail{}
	err := copier.Copy(data, gameActivityEvents)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameActivityEventss(fromValues []*model.GameActivityEvents) ([]*types.GameActivityEventsObjDetail, error) {
	toValues := []*types.GameActivityEventsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameActivityEvents(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
