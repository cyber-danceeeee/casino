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

var _ GameRankingPeriodsHandler = (*gameRankingPeriodsHandler)(nil)

// GameRankingPeriodsHandler defining the handler interface
type GameRankingPeriodsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameRankingPeriodsHandler struct {
	iDao dao.GameRankingPeriodsDao
}

// NewGameRankingPeriodsHandler creating the handler interface
func NewGameRankingPeriodsHandler() GameRankingPeriodsHandler {
	return &gameRankingPeriodsHandler{
		iDao: dao.NewGameRankingPeriodsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameRankingPeriodsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameRankingPeriods
// @Summary Create a new gameRankingPeriods
// @Description Creates a new gameRankingPeriods entity using the provided data in the request body.
// @Tags gameRankingPeriods
// @Accept json
// @Produce json
// @Param data body types.CreateGameRankingPeriodsRequest true "gameRankingPeriods information"
// @Success 200 {object} types.CreateGameRankingPeriodsReply{}
// @Router /api/v1/gameRankingPeriods [post]
// @Security BearerAuth
func (h *gameRankingPeriodsHandler) Create(c *gin.Context) {
	form := &types.CreateGameRankingPeriodsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameRankingPeriods := &model.GameRankingPeriods{}
	err = copier.Copy(gameRankingPeriods, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameRankingPeriods)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameRankingPeriods)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameRankingPeriods.ID})
}

// DeleteByID delete a gameRankingPeriods by id
// @Summary Delete a gameRankingPeriods by id
// @Description Deletes a existing gameRankingPeriods identified by the given id in the path.
// @Tags gameRankingPeriods
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameRankingPeriodsByIDReply{}
// @Router /api/v1/gameRankingPeriods/{id} [delete]
// @Security BearerAuth
func (h *gameRankingPeriodsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameRankingPeriodsIDFromPath(c)
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

// UpdateByID update a gameRankingPeriods by id
// @Summary Update a gameRankingPeriods by id
// @Description Updates the specified gameRankingPeriods by given id in the path, support partial update.
// @Tags gameRankingPeriods
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameRankingPeriodsByIDRequest true "gameRankingPeriods information"
// @Success 200 {object} types.UpdateGameRankingPeriodsByIDReply{}
// @Router /api/v1/gameRankingPeriods/{id} [put]
// @Security BearerAuth
func (h *gameRankingPeriodsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameRankingPeriodsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameRankingPeriodsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameRankingPeriods := &model.GameRankingPeriods{}
	err = copier.Copy(gameRankingPeriods, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameRankingPeriods)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameRankingPeriods)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameRankingPeriods by id
// @Summary Get a gameRankingPeriods by id
// @Description Gets detailed information of a gameRankingPeriods specified by the given id in the path.
// @Tags gameRankingPeriods
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameRankingPeriodsByIDReply{}
// @Router /api/v1/gameRankingPeriods/{id} [get]
// @Security BearerAuth
func (h *gameRankingPeriodsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameRankingPeriodsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameRankingPeriods, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameRankingPeriodsObjDetail{}
	err = copier.Copy(data, gameRankingPeriods)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameRankingPeriods)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameRankingPeriods": data})
}

// List get a paginated list of gameRankingPeriodss by custom conditions
// @Summary Get a paginated list of gameRankingPeriodss by custom conditions
// @Description Returns a paginated list of gameRankingPeriods based on query filters, including page number and size.
// @Tags gameRankingPeriods
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameRankingPeriodssReply{}
// @Router /api/v1/gameRankingPeriods/list [post]
// @Security BearerAuth
func (h *gameRankingPeriodsHandler) List(c *gin.Context) {
	form := &types.ListGameRankingPeriodssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameRankingPeriodss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameRankingPeriodss(gameRankingPeriodss)
	if err != nil {
		response.Error(c, ecode.ErrListGameRankingPeriods)
		return
	}

	response.Success(c, gin.H{
		"gameRankingPeriodss": data,
		"total":        total,
	})
}

func getGameRankingPeriodsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameRankingPeriods(gameRankingPeriods *model.GameRankingPeriods) (*types.GameRankingPeriodsObjDetail, error) {
	data := &types.GameRankingPeriodsObjDetail{}
	err := copier.Copy(data, gameRankingPeriods)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameRankingPeriodss(fromValues []*model.GameRankingPeriods) ([]*types.GameRankingPeriodsObjDetail, error) {
	toValues := []*types.GameRankingPeriodsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameRankingPeriods(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
