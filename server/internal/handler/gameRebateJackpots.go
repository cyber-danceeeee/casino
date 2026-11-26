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

var _ GameRebateJackpotsHandler = (*gameRebateJackpotsHandler)(nil)

// GameRebateJackpotsHandler defining the handler interface
type GameRebateJackpotsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameRebateJackpotsHandler struct {
	iDao dao.GameRebateJackpotsDao
}

// NewGameRebateJackpotsHandler creating the handler interface
func NewGameRebateJackpotsHandler() GameRebateJackpotsHandler {
	return &gameRebateJackpotsHandler{
		iDao: dao.NewGameRebateJackpotsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameRebateJackpotsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameRebateJackpots
// @Summary Create a new gameRebateJackpots
// @Description Creates a new gameRebateJackpots entity using the provided data in the request body.
// @Tags gameRebateJackpots
// @Accept json
// @Produce json
// @Param data body types.CreateGameRebateJackpotsRequest true "gameRebateJackpots information"
// @Success 200 {object} types.CreateGameRebateJackpotsReply{}
// @Router /api/v1/gameRebateJackpots [post]
// @Security BearerAuth
func (h *gameRebateJackpotsHandler) Create(c *gin.Context) {
	form := &types.CreateGameRebateJackpotsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameRebateJackpots := &model.GameRebateJackpots{}
	err = copier.Copy(gameRebateJackpots, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameRebateJackpots)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameRebateJackpots)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameRebateJackpots.ID})
}

// DeleteByID delete a gameRebateJackpots by id
// @Summary Delete a gameRebateJackpots by id
// @Description Deletes a existing gameRebateJackpots identified by the given id in the path.
// @Tags gameRebateJackpots
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameRebateJackpotsByIDReply{}
// @Router /api/v1/gameRebateJackpots/{id} [delete]
// @Security BearerAuth
func (h *gameRebateJackpotsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameRebateJackpotsIDFromPath(c)
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

// UpdateByID update a gameRebateJackpots by id
// @Summary Update a gameRebateJackpots by id
// @Description Updates the specified gameRebateJackpots by given id in the path, support partial update.
// @Tags gameRebateJackpots
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameRebateJackpotsByIDRequest true "gameRebateJackpots information"
// @Success 200 {object} types.UpdateGameRebateJackpotsByIDReply{}
// @Router /api/v1/gameRebateJackpots/{id} [put]
// @Security BearerAuth
func (h *gameRebateJackpotsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameRebateJackpotsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameRebateJackpotsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameRebateJackpots := &model.GameRebateJackpots{}
	err = copier.Copy(gameRebateJackpots, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameRebateJackpots)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameRebateJackpots)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameRebateJackpots by id
// @Summary Get a gameRebateJackpots by id
// @Description Gets detailed information of a gameRebateJackpots specified by the given id in the path.
// @Tags gameRebateJackpots
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameRebateJackpotsByIDReply{}
// @Router /api/v1/gameRebateJackpots/{id} [get]
// @Security BearerAuth
func (h *gameRebateJackpotsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameRebateJackpotsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameRebateJackpots, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameRebateJackpotsObjDetail{}
	err = copier.Copy(data, gameRebateJackpots)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameRebateJackpots)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameRebateJackpots": data})
}

// List get a paginated list of gameRebateJackpotss by custom conditions
// @Summary Get a paginated list of gameRebateJackpotss by custom conditions
// @Description Returns a paginated list of gameRebateJackpots based on query filters, including page number and size.
// @Tags gameRebateJackpots
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameRebateJackpotssReply{}
// @Router /api/v1/gameRebateJackpots/list [post]
// @Security BearerAuth
func (h *gameRebateJackpotsHandler) List(c *gin.Context) {
	form := &types.ListGameRebateJackpotssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameRebateJackpotss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameRebateJackpotss(gameRebateJackpotss)
	if err != nil {
		response.Error(c, ecode.ErrListGameRebateJackpots)
		return
	}

	response.Success(c, gin.H{
		"gameRebateJackpotss": data,
		"total":               total,
	})
}

func getGameRebateJackpotsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameRebateJackpots(gameRebateJackpots *model.GameRebateJackpots) (*types.GameRebateJackpotsObjDetail, error) {
	data := &types.GameRebateJackpotsObjDetail{}
	err := copier.Copy(data, gameRebateJackpots)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameRebateJackpotss(fromValues []*model.GameRebateJackpots) ([]*types.GameRebateJackpotsObjDetail, error) {
	toValues := []*types.GameRebateJackpotsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameRebateJackpots(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
