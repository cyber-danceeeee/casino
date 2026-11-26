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

var _ GameAppDownloadRecordsHandler = (*gameAppDownloadRecordsHandler)(nil)

// GameAppDownloadRecordsHandler defining the handler interface
type GameAppDownloadRecordsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameAppDownloadRecordsHandler struct {
	iDao dao.GameAppDownloadRecordsDao
}

// NewGameAppDownloadRecordsHandler creating the handler interface
func NewGameAppDownloadRecordsHandler() GameAppDownloadRecordsHandler {
	return &gameAppDownloadRecordsHandler{
		iDao: dao.NewGameAppDownloadRecordsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameAppDownloadRecordsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameAppDownloadRecords
// @Summary Create a new gameAppDownloadRecords
// @Description Creates a new gameAppDownloadRecords entity using the provided data in the request body.
// @Tags gameAppDownloadRecords
// @Accept json
// @Produce json
// @Param data body types.CreateGameAppDownloadRecordsRequest true "gameAppDownloadRecords information"
// @Success 200 {object} types.CreateGameAppDownloadRecordsReply{}
// @Router /api/v1/gameAppDownloadRecords [post]
// @Security BearerAuth
func (h *gameAppDownloadRecordsHandler) Create(c *gin.Context) {
	form := &types.CreateGameAppDownloadRecordsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameAppDownloadRecords := &model.GameAppDownloadRecords{}
	err = copier.Copy(gameAppDownloadRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameAppDownloadRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameAppDownloadRecords)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameAppDownloadRecords.ID})
}

// DeleteByID delete a gameAppDownloadRecords by id
// @Summary Delete a gameAppDownloadRecords by id
// @Description Deletes a existing gameAppDownloadRecords identified by the given id in the path.
// @Tags gameAppDownloadRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameAppDownloadRecordsByIDReply{}
// @Router /api/v1/gameAppDownloadRecords/{id} [delete]
// @Security BearerAuth
func (h *gameAppDownloadRecordsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameAppDownloadRecordsIDFromPath(c)
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

// UpdateByID update a gameAppDownloadRecords by id
// @Summary Update a gameAppDownloadRecords by id
// @Description Updates the specified gameAppDownloadRecords by given id in the path, support partial update.
// @Tags gameAppDownloadRecords
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameAppDownloadRecordsByIDRequest true "gameAppDownloadRecords information"
// @Success 200 {object} types.UpdateGameAppDownloadRecordsByIDReply{}
// @Router /api/v1/gameAppDownloadRecords/{id} [put]
// @Security BearerAuth
func (h *gameAppDownloadRecordsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameAppDownloadRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameAppDownloadRecordsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameAppDownloadRecords := &model.GameAppDownloadRecords{}
	err = copier.Copy(gameAppDownloadRecords, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameAppDownloadRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameAppDownloadRecords)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameAppDownloadRecords by id
// @Summary Get a gameAppDownloadRecords by id
// @Description Gets detailed information of a gameAppDownloadRecords specified by the given id in the path.
// @Tags gameAppDownloadRecords
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameAppDownloadRecordsByIDReply{}
// @Router /api/v1/gameAppDownloadRecords/{id} [get]
// @Security BearerAuth
func (h *gameAppDownloadRecordsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameAppDownloadRecordsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameAppDownloadRecords, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameAppDownloadRecordsObjDetail{}
	err = copier.Copy(data, gameAppDownloadRecords)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameAppDownloadRecords)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameAppDownloadRecords": data})
}

// List get a paginated list of gameAppDownloadRecordss by custom conditions
// @Summary Get a paginated list of gameAppDownloadRecordss by custom conditions
// @Description Returns a paginated list of gameAppDownloadRecords based on query filters, including page number and size.
// @Tags gameAppDownloadRecords
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameAppDownloadRecordssReply{}
// @Router /api/v1/gameAppDownloadRecords/list [post]
// @Security BearerAuth
func (h *gameAppDownloadRecordsHandler) List(c *gin.Context) {
	form := &types.ListGameAppDownloadRecordssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameAppDownloadRecordss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameAppDownloadRecordss(gameAppDownloadRecordss)
	if err != nil {
		response.Error(c, ecode.ErrListGameAppDownloadRecords)
		return
	}

	response.Success(c, gin.H{
		"gameAppDownloadRecordss": data,
		"total":                   total,
	})
}

func getGameAppDownloadRecordsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameAppDownloadRecords(gameAppDownloadRecords *model.GameAppDownloadRecords) (*types.GameAppDownloadRecordsObjDetail, error) {
	data := &types.GameAppDownloadRecordsObjDetail{}
	err := copier.Copy(data, gameAppDownloadRecords)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameAppDownloadRecordss(fromValues []*model.GameAppDownloadRecords) ([]*types.GameAppDownloadRecordsObjDetail, error) {
	toValues := []*types.GameAppDownloadRecordsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameAppDownloadRecords(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
