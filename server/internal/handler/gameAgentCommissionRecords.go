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

var _ GameAgentCommissionRecordsHandler = (*gameAgentCommissionRecordsHandler)(nil)

// GameAgentCommissionRecordsHandler defining the handler interface
type GameAgentCommissionRecordsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameAgentCommissionRecordsHandler struct {
	iDao dao.GameAgentCommissionRecordsDao
}

// NewGameAgentCommissionRecordsHandler creating the handler interface
func NewGameAgentCommissionRecordsHandler() GameAgentCommissionRecordsHandler {
	return &gameAgentCommissionRecordsHandler{
		iDao: dao.NewGameAgentCommissionRecordsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameAgentCommissionRecordsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameAgentCommissionRecords
// @Summary Create a new gameAgentCommissionRecords
// @Description Creates a new gameAgentCommissionRecords entity using the provided data in the request body.
// @Tags gameAgentCommissionRecords
// @Accept json
// @Produce json
// @Param data body types.CreateGameAgentCommissionRecordsRequest true "gameAgentCommissionRecords information"
// @Success 200 {object} types.CreateGameAgentCommissionRecordsReply{}
// @Router /api/v1/gameAgentCommissionRecords [post]
// @Security BearerAuth
func (h *gameAgentCommissionRecordsHandler) Create(c *gin.Context) {
	form := &types.CreateGameAgentCommissionRecordsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameAgentCommissionRecords := &model.GameAgentCommissionRecords{}
	err = copier.Copy(gameAgentCommissionRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameAgentCommissionRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameAgentCommissionRecords)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameAgentCommissionRecords.ID})
}

// DeleteByID delete a gameAgentCommissionRecords by id
// @Summary Delete a gameAgentCommissionRecords by id
// @Description Deletes a existing gameAgentCommissionRecords identified by the given id in the path.
// @Tags gameAgentCommissionRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameAgentCommissionRecordsByIDReply{}
// @Router /api/v1/gameAgentCommissionRecords/{id} [delete]
// @Security BearerAuth
func (h *gameAgentCommissionRecordsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameAgentCommissionRecordsIDFromPath(c)
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

// UpdateByID update a gameAgentCommissionRecords by id
// @Summary Update a gameAgentCommissionRecords by id
// @Description Updates the specified gameAgentCommissionRecords by given id in the path, support partial update.
// @Tags gameAgentCommissionRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameAgentCommissionRecordsByIDRequest true "gameAgentCommissionRecords information"
// @Success 200 {object} types.UpdateGameAgentCommissionRecordsByIDReply{}
// @Router /api/v1/gameAgentCommissionRecords/{id} [put]
// @Security BearerAuth
func (h *gameAgentCommissionRecordsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameAgentCommissionRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameAgentCommissionRecordsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameAgentCommissionRecords := &model.GameAgentCommissionRecords{}
	err = copier.Copy(gameAgentCommissionRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameAgentCommissionRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameAgentCommissionRecords)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameAgentCommissionRecords by id
// @Summary Get a gameAgentCommissionRecords by id
// @Description Gets detailed information of a gameAgentCommissionRecords specified by the given id in the path.
// @Tags gameAgentCommissionRecords
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameAgentCommissionRecordsByIDReply{}
// @Router /api/v1/gameAgentCommissionRecords/{id} [get]
// @Security BearerAuth
func (h *gameAgentCommissionRecordsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameAgentCommissionRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameAgentCommissionRecords, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameAgentCommissionRecordsObjDetail{}
	err = copier.Copy(data, gameAgentCommissionRecords)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameAgentCommissionRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameAgentCommissionRecords": data})
}

// List get a paginated list of gameAgentCommissionRecordss by custom conditions
// @Summary Get a paginated list of gameAgentCommissionRecordss by custom conditions
// @Description Returns a paginated list of gameAgentCommissionRecords based on query filters, including page number and size.
// @Tags gameAgentCommissionRecords
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameAgentCommissionRecordssReply{}
// @Router /api/v1/gameAgentCommissionRecords/list [post]
// @Security BearerAuth
func (h *gameAgentCommissionRecordsHandler) List(c *gin.Context) {
	form := &types.ListGameAgentCommissionRecordssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameAgentCommissionRecordss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameAgentCommissionRecordss(gameAgentCommissionRecordss)
	if err != nil {
		response.Error(c, ecode.ErrListGameAgentCommissionRecords)
		return
	}

	response.Success(c, gin.H{
		"gameAgentCommissionRecordss": data,
		"total":        total,
	})
}

func getGameAgentCommissionRecordsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameAgentCommissionRecords(gameAgentCommissionRecords *model.GameAgentCommissionRecords) (*types.GameAgentCommissionRecordsObjDetail, error) {
	data := &types.GameAgentCommissionRecordsObjDetail{}
	err := copier.Copy(data, gameAgentCommissionRecords)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameAgentCommissionRecordss(fromValues []*model.GameAgentCommissionRecords) ([]*types.GameAgentCommissionRecordsObjDetail, error) {
	toValues := []*types.GameAgentCommissionRecordsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameAgentCommissionRecords(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
