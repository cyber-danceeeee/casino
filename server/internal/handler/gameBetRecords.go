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

var _ GameBetRecordsHandler = (*gameBetRecordsHandler)(nil)

// GameBetRecordsHandler defining the handler interface
type GameBetRecordsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	PlayerRecords(c *gin.Context)
}

type gameBetRecordsHandler struct {
	iDao dao.GameBetRecordsDao
}

// NewGameBetRecordsHandler creating the handler interface
func NewGameBetRecordsHandler() GameBetRecordsHandler {
	return &gameBetRecordsHandler{
		iDao: dao.NewGameBetRecordsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameBetRecordsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameBetRecords
// @Summary Create a new gameBetRecords
// @Description Creates a new gameBetRecords entity using the provided data in the request body.
// @Tags gameBetRecords
// @Accept json
// @Produce json
// @Param data body types.CreateGameBetRecordsRequest true "gameBetRecords information"
// @Success 200 {object} types.CreateGameBetRecordsReply{}
// @Router /api/v1/gameBetRecords [post]
// @Security BearerAuth
func (h *gameBetRecordsHandler) Create(c *gin.Context) {
	form := &types.CreateGameBetRecordsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameBetRecords := &model.GameBetRecords{}
	err = copier.Copy(gameBetRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameBetRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameBetRecords)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameBetRecords.ID})
}

// DeleteByID delete a gameBetRecords by id
// @Summary Delete a gameBetRecords by id
// @Description Deletes a existing gameBetRecords identified by the given id in the path.
// @Tags gameBetRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameBetRecordsByIDReply{}
// @Router /api/v1/gameBetRecords/{id} [delete]
// @Security BearerAuth
func (h *gameBetRecordsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameBetRecordsIDFromPath(c)
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

// UpdateByID update a gameBetRecords by id
// @Summary Update a gameBetRecords by id
// @Description Updates the specified gameBetRecords by given id in the path, support partial update.
// @Tags gameBetRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameBetRecordsByIDRequest true "gameBetRecords information"
// @Success 200 {object} types.UpdateGameBetRecordsByIDReply{}
// @Router /api/v1/gameBetRecords/{id} [put]
// @Security BearerAuth
func (h *gameBetRecordsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameBetRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameBetRecordsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameBetRecords := &model.GameBetRecords{}
	err = copier.Copy(gameBetRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameBetRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameBetRecords)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameBetRecords by id
// @Summary Get a gameBetRecords by id
// @Description Gets detailed information of a gameBetRecords specified by the given id in the path.
// @Tags gameBetRecords
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameBetRecordsByIDReply{}
// @Router /api/v1/gameBetRecords/{id} [get]
// @Security BearerAuth
func (h *gameBetRecordsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameBetRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameBetRecords, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameBetRecordsObjDetail{}
	err = copier.Copy(data, gameBetRecords)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameBetRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameBetRecords": data})
}

// List get a paginated list of gameBetRecordss by custom conditions
// @Summary Get a paginated list of gameBetRecordss by custom conditions
// @Description Returns a paginated list of gameBetRecords based on query filters, including page number and size.
// @Tags gameBetRecords
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameBetRecordssReply{}
// @Router /api/v1/gameBetRecords/list [post]
// @Security BearerAuth
func (h *gameBetRecordsHandler) List(c *gin.Context) {
	form := &types.ListGameBetRecordssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameBetRecordss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameBetRecordss(gameBetRecordss)
	if err != nil {
		response.Error(c, ecode.ErrListGameBetRecords)
		return
	}

	response.Success(c, gin.H{
		"gameBetRecordss": data,
		"total":           total,
	})
}

func getGameBetRecordsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameBetRecords(gameBetRecords *model.GameBetRecords) (*types.GameBetRecordsObjDetail, error) {
	data := &types.GameBetRecordsObjDetail{}
	err := copier.Copy(data, gameBetRecords)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameBetRecordss(fromValues []*model.GameBetRecords) ([]*types.GameBetRecordsObjDetail, error) {
	toValues := []*types.GameBetRecordsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameBetRecords(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
