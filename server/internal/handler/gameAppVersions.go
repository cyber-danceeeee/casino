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

var _ GameAppVersionsHandler = (*gameAppVersionsHandler)(nil)

// GameAppVersionsHandler defining the handler interface
type GameAppVersionsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Current(c *gin.Context)
}

type gameAppVersionsHandler struct {
	iDao dao.GameAppVersionsDao
}

// NewGameAppVersionsHandler creating the handler interface
func NewGameAppVersionsHandler() GameAppVersionsHandler {
	return &gameAppVersionsHandler{
		iDao: dao.NewGameAppVersionsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameAppVersionsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameAppVersions
// @Summary Create a new gameAppVersions
// @Description Creates a new gameAppVersions entity using the provided data in the request body.
// @Tags gameAppVersions
// @Accept json
// @Produce json
// @Param data body types.CreateGameAppVersionsRequest true "gameAppVersions information"
// @Success 200 {object} types.CreateGameAppVersionsReply{}
// @Router /api/v1/gameAppVersions [post]
// @Security BearerAuth
func (h *gameAppVersionsHandler) Create(c *gin.Context) {
	form := &types.CreateGameAppVersionsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameAppVersions := &model.GameAppVersions{}
	err = copier.Copy(gameAppVersions, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameAppVersions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameAppVersions)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameAppVersions.ID})
}

// DeleteByID delete a gameAppVersions by id
// @Summary Delete a gameAppVersions by id
// @Description Deletes a existing gameAppVersions identified by the given id in the path.
// @Tags gameAppVersions
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameAppVersionsByIDReply{}
// @Router /api/v1/gameAppVersions/{id} [delete]
// @Security BearerAuth
func (h *gameAppVersionsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameAppVersionsIDFromPath(c)
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

// UpdateByID update a gameAppVersions by id
// @Summary Update a gameAppVersions by id
// @Description Updates the specified gameAppVersions by given id in the path, support partial update.
// @Tags gameAppVersions
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameAppVersionsByIDRequest true "gameAppVersions information"
// @Success 200 {object} types.UpdateGameAppVersionsByIDReply{}
// @Router /api/v1/gameAppVersions/{id} [put]
// @Security BearerAuth
func (h *gameAppVersionsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameAppVersionsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameAppVersionsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameAppVersions := &model.GameAppVersions{}
	err = copier.Copy(gameAppVersions, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameAppVersions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameAppVersions)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameAppVersions by id
// @Summary Get a gameAppVersions by id
// @Description Gets detailed information of a gameAppVersions specified by the given id in the path.
// @Tags gameAppVersions
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameAppVersionsByIDReply{}
// @Router /api/v1/gameAppVersions/{id} [get]
// @Security BearerAuth
func (h *gameAppVersionsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameAppVersionsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameAppVersions, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameAppVersionsObjDetail{}
	err = copier.Copy(data, gameAppVersions)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameAppVersions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameAppVersions": data})
}

// List get a paginated list of gameAppVersionss by custom conditions
// @Summary Get a paginated list of gameAppVersionss by custom conditions
// @Description Returns a paginated list of gameAppVersions based on query filters, including page number and size.
// @Tags gameAppVersions
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameAppVersionssReply{}
// @Router /api/v1/gameAppVersions/list [post]
// @Security BearerAuth
func (h *gameAppVersionsHandler) List(c *gin.Context) {
	form := &types.ListGameAppVersionssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameAppVersionss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameAppVersionss(gameAppVersionss)
	if err != nil {
		response.Error(c, ecode.ErrListGameAppVersions)
		return
	}

	response.Success(c, gin.H{
		"gameAppVersionss": data,
		"total":            total,
	})
}

func getGameAppVersionsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameAppVersions(gameAppVersions *model.GameAppVersions) (*types.GameAppVersionsObjDetail, error) {
	data := &types.GameAppVersionsObjDetail{}
	err := copier.Copy(data, gameAppVersions)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameAppVersionss(fromValues []*model.GameAppVersions) ([]*types.GameAppVersionsObjDetail, error) {
	toValues := []*types.GameAppVersionsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameAppVersions(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
