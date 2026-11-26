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

var _ GameBonusRecordsHandler = (*gameBonusRecordsHandler)(nil)

// GameBonusRecordsHandler defining the handler interface
type GameBonusRecordsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameBonusRecordsHandler struct {
	iDao dao.GameBonusRecordsDao
}

// NewGameBonusRecordsHandler creating the handler interface
func NewGameBonusRecordsHandler() GameBonusRecordsHandler {
	return &gameBonusRecordsHandler{
		iDao: dao.NewGameBonusRecordsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameBonusRecordsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameBonusRecords
// @Summary Create a new gameBonusRecords
// @Description Creates a new gameBonusRecords entity using the provided data in the request body.
// @Tags gameBonusRecords
// @Accept json
// @Produce json
// @Param data body types.CreateGameBonusRecordsRequest true "gameBonusRecords information"
// @Success 200 {object} types.CreateGameBonusRecordsReply{}
// @Router /api/v1/gameBonusRecords [post]
// @Security BearerAuth
func (h *gameBonusRecordsHandler) Create(c *gin.Context) {
	form := &types.CreateGameBonusRecordsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameBonusRecords := &model.GameBonusRecords{}
	err = copier.Copy(gameBonusRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameBonusRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameBonusRecords)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameBonusRecords.ID})
}

// DeleteByID delete a gameBonusRecords by id
// @Summary Delete a gameBonusRecords by id
// @Description Deletes a existing gameBonusRecords identified by the given id in the path.
// @Tags gameBonusRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameBonusRecordsByIDReply{}
// @Router /api/v1/gameBonusRecords/{id} [delete]
// @Security BearerAuth
func (h *gameBonusRecordsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameBonusRecordsIDFromPath(c)
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

// UpdateByID update a gameBonusRecords by id
// @Summary Update a gameBonusRecords by id
// @Description Updates the specified gameBonusRecords by given id in the path, support partial update.
// @Tags gameBonusRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameBonusRecordsByIDRequest true "gameBonusRecords information"
// @Success 200 {object} types.UpdateGameBonusRecordsByIDReply{}
// @Router /api/v1/gameBonusRecords/{id} [put]
// @Security BearerAuth
func (h *gameBonusRecordsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameBonusRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameBonusRecordsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameBonusRecords := &model.GameBonusRecords{}
	err = copier.Copy(gameBonusRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameBonusRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameBonusRecords)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameBonusRecords by id
// @Summary Get a gameBonusRecords by id
// @Description Gets detailed information of a gameBonusRecords specified by the given id in the path.
// @Tags gameBonusRecords
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameBonusRecordsByIDReply{}
// @Router /api/v1/gameBonusRecords/{id} [get]
// @Security BearerAuth
func (h *gameBonusRecordsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameBonusRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameBonusRecords, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameBonusRecordsObjDetail{}
	err = copier.Copy(data, gameBonusRecords)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameBonusRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameBonusRecords": data})
}

// List get a paginated list of gameBonusRecordss by custom conditions
// @Summary Get a paginated list of gameBonusRecordss by custom conditions
// @Description Returns a paginated list of gameBonusRecords based on query filters, including page number and size.
// @Tags gameBonusRecords
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameBonusRecordssReply{}
// @Router /api/v1/gameBonusRecords/list [post]
// @Security BearerAuth
func (h *gameBonusRecordsHandler) List(c *gin.Context) {
	form := &types.ListGameBonusRecordssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameBonusRecordss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameBonusRecordss(gameBonusRecordss)
	if err != nil {
		response.Error(c, ecode.ErrListGameBonusRecords)
		return
	}

	response.Success(c, gin.H{
		"gameBonusRecordss": data,
		"total":             total,
	})
}

func getGameBonusRecordsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameBonusRecords(gameBonusRecords *model.GameBonusRecords) (*types.GameBonusRecordsObjDetail, error) {
	data := &types.GameBonusRecordsObjDetail{}
	err := copier.Copy(data, gameBonusRecords)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameBonusRecordss(fromValues []*model.GameBonusRecords) ([]*types.GameBonusRecordsObjDetail, error) {
	toValues := []*types.GameBonusRecordsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameBonusRecords(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
