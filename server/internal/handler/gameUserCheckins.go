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

var _ GameUserCheckinsHandler = (*gameUserCheckinsHandler)(nil)

// GameUserCheckinsHandler defining the handler interface
type GameUserCheckinsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameUserCheckinsHandler struct {
	iDao dao.GameUserCheckinsDao
}

// NewGameUserCheckinsHandler creating the handler interface
func NewGameUserCheckinsHandler() GameUserCheckinsHandler {
	return &gameUserCheckinsHandler{
		iDao: dao.NewGameUserCheckinsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameUserCheckinsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameUserCheckins
// @Summary Create a new gameUserCheckins
// @Description Creates a new gameUserCheckins entity using the provided data in the request body.
// @Tags gameUserCheckins
// @Accept json
// @Produce json
// @Param data body types.CreateGameUserCheckinsRequest true "gameUserCheckins information"
// @Success 200 {object} types.CreateGameUserCheckinsReply{}
// @Router /api/v1/gameUserCheckins [post]
// @Security BearerAuth
func (h *gameUserCheckinsHandler) Create(c *gin.Context) {
	form := &types.CreateGameUserCheckinsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameUserCheckins := &model.GameUserCheckins{}
	err = copier.Copy(gameUserCheckins, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameUserCheckins)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameUserCheckins)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameUserCheckins.ID})
}

// DeleteByID delete a gameUserCheckins by id
// @Summary Delete a gameUserCheckins by id
// @Description Deletes a existing gameUserCheckins identified by the given id in the path.
// @Tags gameUserCheckins
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameUserCheckinsByIDReply{}
// @Router /api/v1/gameUserCheckins/{id} [delete]
// @Security BearerAuth
func (h *gameUserCheckinsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameUserCheckinsIDFromPath(c)
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

// UpdateByID update a gameUserCheckins by id
// @Summary Update a gameUserCheckins by id
// @Description Updates the specified gameUserCheckins by given id in the path, support partial update.
// @Tags gameUserCheckins
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameUserCheckinsByIDRequest true "gameUserCheckins information"
// @Success 200 {object} types.UpdateGameUserCheckinsByIDReply{}
// @Router /api/v1/gameUserCheckins/{id} [put]
// @Security BearerAuth
func (h *gameUserCheckinsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameUserCheckinsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameUserCheckinsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameUserCheckins := &model.GameUserCheckins{}
	err = copier.Copy(gameUserCheckins, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameUserCheckins)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameUserCheckins)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameUserCheckins by id
// @Summary Get a gameUserCheckins by id
// @Description Gets detailed information of a gameUserCheckins specified by the given id in the path.
// @Tags gameUserCheckins
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameUserCheckinsByIDReply{}
// @Router /api/v1/gameUserCheckins/{id} [get]
// @Security BearerAuth
func (h *gameUserCheckinsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameUserCheckinsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserCheckins, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameUserCheckinsObjDetail{}
	err = copier.Copy(data, gameUserCheckins)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameUserCheckins)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameUserCheckins": data})
}

// List get a paginated list of gameUserCheckinss by custom conditions
// @Summary Get a paginated list of gameUserCheckinss by custom conditions
// @Description Returns a paginated list of gameUserCheckins based on query filters, including page number and size.
// @Tags gameUserCheckins
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameUserCheckinssReply{}
// @Router /api/v1/gameUserCheckins/list [post]
// @Security BearerAuth
func (h *gameUserCheckinsHandler) List(c *gin.Context) {
	form := &types.ListGameUserCheckinssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserCheckinss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameUserCheckinss(gameUserCheckinss)
	if err != nil {
		response.Error(c, ecode.ErrListGameUserCheckins)
		return
	}

	response.Success(c, gin.H{
		"gameUserCheckinss": data,
		"total":        total,
	})
}

func getGameUserCheckinsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameUserCheckins(gameUserCheckins *model.GameUserCheckins) (*types.GameUserCheckinsObjDetail, error) {
	data := &types.GameUserCheckinsObjDetail{}
	err := copier.Copy(data, gameUserCheckins)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameUserCheckinss(fromValues []*model.GameUserCheckins) ([]*types.GameUserCheckinsObjDetail, error) {
	toValues := []*types.GameUserCheckinsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameUserCheckins(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
