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

var _ GameUserActivityRecordsHandler = (*gameUserActivityRecordsHandler)(nil)

// GameUserActivityRecordsHandler defining the handler interface
type GameUserActivityRecordsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameUserActivityRecordsHandler struct {
	iDao dao.GameUserActivityRecordsDao
}

// NewGameUserActivityRecordsHandler creating the handler interface
func NewGameUserActivityRecordsHandler() GameUserActivityRecordsHandler {
	return &gameUserActivityRecordsHandler{
		iDao: dao.NewGameUserActivityRecordsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameUserActivityRecordsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameUserActivityRecords
// @Summary Create a new gameUserActivityRecords
// @Description Creates a new gameUserActivityRecords entity using the provided data in the request body.
// @Tags gameUserActivityRecords
// @Accept json
// @Produce json
// @Param data body types.CreateGameUserActivityRecordsRequest true "gameUserActivityRecords information"
// @Success 200 {object} types.CreateGameUserActivityRecordsReply{}
// @Router /api/v1/gameUserActivityRecords [post]
// @Security BearerAuth
func (h *gameUserActivityRecordsHandler) Create(c *gin.Context) {
	form := &types.CreateGameUserActivityRecordsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameUserActivityRecords := &model.GameUserActivityRecords{}
	err = copier.Copy(gameUserActivityRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameUserActivityRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameUserActivityRecords)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameUserActivityRecords.ID})
}

// DeleteByID delete a gameUserActivityRecords by id
// @Summary Delete a gameUserActivityRecords by id
// @Description Deletes a existing gameUserActivityRecords identified by the given id in the path.
// @Tags gameUserActivityRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameUserActivityRecordsByIDReply{}
// @Router /api/v1/gameUserActivityRecords/{id} [delete]
// @Security BearerAuth
func (h *gameUserActivityRecordsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameUserActivityRecordsIDFromPath(c)
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

// UpdateByID update a gameUserActivityRecords by id
// @Summary Update a gameUserActivityRecords by id
// @Description Updates the specified gameUserActivityRecords by given id in the path, support partial update.
// @Tags gameUserActivityRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameUserActivityRecordsByIDRequest true "gameUserActivityRecords information"
// @Success 200 {object} types.UpdateGameUserActivityRecordsByIDReply{}
// @Router /api/v1/gameUserActivityRecords/{id} [put]
// @Security BearerAuth
func (h *gameUserActivityRecordsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameUserActivityRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameUserActivityRecordsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameUserActivityRecords := &model.GameUserActivityRecords{}
	err = copier.Copy(gameUserActivityRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameUserActivityRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameUserActivityRecords)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameUserActivityRecords by id
// @Summary Get a gameUserActivityRecords by id
// @Description Gets detailed information of a gameUserActivityRecords specified by the given id in the path.
// @Tags gameUserActivityRecords
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameUserActivityRecordsByIDReply{}
// @Router /api/v1/gameUserActivityRecords/{id} [get]
// @Security BearerAuth
func (h *gameUserActivityRecordsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameUserActivityRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserActivityRecords, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameUserActivityRecordsObjDetail{}
	err = copier.Copy(data, gameUserActivityRecords)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameUserActivityRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameUserActivityRecords": data})
}

// List get a paginated list of gameUserActivityRecordss by custom conditions
// @Summary Get a paginated list of gameUserActivityRecordss by custom conditions
// @Description Returns a paginated list of gameUserActivityRecords based on query filters, including page number and size.
// @Tags gameUserActivityRecords
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameUserActivityRecordssReply{}
// @Router /api/v1/gameUserActivityRecords/list [post]
// @Security BearerAuth
func (h *gameUserActivityRecordsHandler) List(c *gin.Context) {
	form := &types.ListGameUserActivityRecordssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserActivityRecordss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameUserActivityRecordss(gameUserActivityRecordss)
	if err != nil {
		response.Error(c, ecode.ErrListGameUserActivityRecords)
		return
	}

	response.Success(c, gin.H{
		"gameUserActivityRecordss": data,
		"total":                    total,
	})
}

func getGameUserActivityRecordsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameUserActivityRecords(gameUserActivityRecords *model.GameUserActivityRecords) (*types.GameUserActivityRecordsObjDetail, error) {
	data := &types.GameUserActivityRecordsObjDetail{}
	err := copier.Copy(data, gameUserActivityRecords)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameUserActivityRecordss(fromValues []*model.GameUserActivityRecords) ([]*types.GameUserActivityRecordsObjDetail, error) {
	toValues := []*types.GameUserActivityRecordsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameUserActivityRecords(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
