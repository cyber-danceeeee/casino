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

var _ GameRankingRecordsHandler = (*gameRankingRecordsHandler)(nil)

// GameRankingRecordsHandler defining the handler interface
type GameRankingRecordsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameRankingRecordsHandler struct {
	iDao dao.GameRankingRecordsDao
}

// NewGameRankingRecordsHandler creating the handler interface
func NewGameRankingRecordsHandler() GameRankingRecordsHandler {
	return &gameRankingRecordsHandler{
		iDao: dao.NewGameRankingRecordsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameRankingRecordsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameRankingRecords
// @Summary Create a new gameRankingRecords
// @Description Creates a new gameRankingRecords entity using the provided data in the request body.
// @Tags gameRankingRecords
// @Accept json
// @Produce json
// @Param data body types.CreateGameRankingRecordsRequest true "gameRankingRecords information"
// @Success 200 {object} types.CreateGameRankingRecordsReply{}
// @Router /api/v1/gameRankingRecords [post]
// @Security BearerAuth
func (h *gameRankingRecordsHandler) Create(c *gin.Context) {
	form := &types.CreateGameRankingRecordsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameRankingRecords := &model.GameRankingRecords{}
	err = copier.Copy(gameRankingRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameRankingRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameRankingRecords)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameRankingRecords.ID})
}

// DeleteByID delete a gameRankingRecords by id
// @Summary Delete a gameRankingRecords by id
// @Description Deletes a existing gameRankingRecords identified by the given id in the path.
// @Tags gameRankingRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameRankingRecordsByIDReply{}
// @Router /api/v1/gameRankingRecords/{id} [delete]
// @Security BearerAuth
func (h *gameRankingRecordsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameRankingRecordsIDFromPath(c)
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

// UpdateByID update a gameRankingRecords by id
// @Summary Update a gameRankingRecords by id
// @Description Updates the specified gameRankingRecords by given id in the path, support partial update.
// @Tags gameRankingRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameRankingRecordsByIDRequest true "gameRankingRecords information"
// @Success 200 {object} types.UpdateGameRankingRecordsByIDReply{}
// @Router /api/v1/gameRankingRecords/{id} [put]
// @Security BearerAuth
func (h *gameRankingRecordsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameRankingRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameRankingRecordsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameRankingRecords := &model.GameRankingRecords{}
	err = copier.Copy(gameRankingRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameRankingRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameRankingRecords)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameRankingRecords by id
// @Summary Get a gameRankingRecords by id
// @Description Gets detailed information of a gameRankingRecords specified by the given id in the path.
// @Tags gameRankingRecords
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameRankingRecordsByIDReply{}
// @Router /api/v1/gameRankingRecords/{id} [get]
// @Security BearerAuth
func (h *gameRankingRecordsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameRankingRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameRankingRecords, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameRankingRecordsObjDetail{}
	err = copier.Copy(data, gameRankingRecords)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameRankingRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameRankingRecords": data})
}

// List get a paginated list of gameRankingRecordss by custom conditions
// @Summary Get a paginated list of gameRankingRecordss by custom conditions
// @Description Returns a paginated list of gameRankingRecords based on query filters, including page number and size.
// @Tags gameRankingRecords
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameRankingRecordssReply{}
// @Router /api/v1/gameRankingRecords/list [post]
// @Security BearerAuth
func (h *gameRankingRecordsHandler) List(c *gin.Context) {
	form := &types.ListGameRankingRecordssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameRankingRecordss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameRankingRecordss(gameRankingRecordss)
	if err != nil {
		response.Error(c, ecode.ErrListGameRankingRecords)
		return
	}

	response.Success(c, gin.H{
		"gameRankingRecordss": data,
		"total":               total,
	})
}

func getGameRankingRecordsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameRankingRecords(gameRankingRecords *model.GameRankingRecords) (*types.GameRankingRecordsObjDetail, error) {
	data := &types.GameRankingRecordsObjDetail{}
	err := copier.Copy(data, gameRankingRecords)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameRankingRecordss(fromValues []*model.GameRankingRecords) ([]*types.GameRankingRecordsObjDetail, error) {
	toValues := []*types.GameRankingRecordsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameRankingRecords(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
