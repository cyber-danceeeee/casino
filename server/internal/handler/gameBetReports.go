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

var _ GameBetReportsHandler = (*gameBetReportsHandler)(nil)

// GameBetReportsHandler defining the handler interface
type GameBetReportsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameBetReportsHandler struct {
	iDao dao.GameBetReportsDao
}

// NewGameBetReportsHandler creating the handler interface
func NewGameBetReportsHandler() GameBetReportsHandler {
	return &gameBetReportsHandler{
		iDao: dao.NewGameBetReportsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameBetReportsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameBetReports
// @Summary Create a new gameBetReports
// @Description Creates a new gameBetReports entity using the provided data in the request body.
// @Tags gameBetReports
// @Accept json
// @Produce json
// @Param data body types.CreateGameBetReportsRequest true "gameBetReports information"
// @Success 200 {object} types.CreateGameBetReportsReply{}
// @Router /api/v1/gameBetReports [post]
// @Security BearerAuth
func (h *gameBetReportsHandler) Create(c *gin.Context) {
	form := &types.CreateGameBetReportsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameBetReports := &model.GameBetReports{}
	err = copier.Copy(gameBetReports, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameBetReports)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameBetReports)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameBetReports.ID})
}

// DeleteByID delete a gameBetReports by id
// @Summary Delete a gameBetReports by id
// @Description Deletes a existing gameBetReports identified by the given id in the path.
// @Tags gameBetReports
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameBetReportsByIDReply{}
// @Router /api/v1/gameBetReports/{id} [delete]
// @Security BearerAuth
func (h *gameBetReportsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameBetReportsIDFromPath(c)
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

// UpdateByID update a gameBetReports by id
// @Summary Update a gameBetReports by id
// @Description Updates the specified gameBetReports by given id in the path, support partial update.
// @Tags gameBetReports
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameBetReportsByIDRequest true "gameBetReports information"
// @Success 200 {object} types.UpdateGameBetReportsByIDReply{}
// @Router /api/v1/gameBetReports/{id} [put]
// @Security BearerAuth
func (h *gameBetReportsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameBetReportsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameBetReportsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameBetReports := &model.GameBetReports{}
	err = copier.Copy(gameBetReports, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameBetReports)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameBetReports)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameBetReports by id
// @Summary Get a gameBetReports by id
// @Description Gets detailed information of a gameBetReports specified by the given id in the path.
// @Tags gameBetReports
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameBetReportsByIDReply{}
// @Router /api/v1/gameBetReports/{id} [get]
// @Security BearerAuth
func (h *gameBetReportsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameBetReportsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameBetReports, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameBetReportsObjDetail{}
	err = copier.Copy(data, gameBetReports)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameBetReports)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameBetReports": data})
}

// List get a paginated list of gameBetReportss by custom conditions
// @Summary Get a paginated list of gameBetReportss by custom conditions
// @Description Returns a paginated list of gameBetReports based on query filters, including page number and size.
// @Tags gameBetReports
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameBetReportssReply{}
// @Router /api/v1/gameBetReports/list [post]
// @Security BearerAuth
func (h *gameBetReportsHandler) List(c *gin.Context) {
	form := &types.ListGameBetReportssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameBetReportss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameBetReportss(gameBetReportss)
	if err != nil {
		response.Error(c, ecode.ErrListGameBetReports)
		return
	}

	response.Success(c, gin.H{
		"gameBetReportss": data,
		"total":           total,
	})
}

func getGameBetReportsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameBetReports(gameBetReports *model.GameBetReports) (*types.GameBetReportsObjDetail, error) {
	data := &types.GameBetReportsObjDetail{}
	err := copier.Copy(data, gameBetReports)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameBetReportss(fromValues []*model.GameBetReports) ([]*types.GameBetReportsObjDetail, error) {
	toValues := []*types.GameBetReportsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameBetReports(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
