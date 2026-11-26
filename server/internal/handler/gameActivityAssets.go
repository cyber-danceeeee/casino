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

var _ GameActivityAssetsHandler = (*gameActivityAssetsHandler)(nil)

// GameActivityAssetsHandler defining the handler interface
type GameActivityAssetsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameActivityAssetsHandler struct {
	iDao dao.GameActivityAssetsDao
}

// NewGameActivityAssetsHandler creating the handler interface
func NewGameActivityAssetsHandler() GameActivityAssetsHandler {
	return &gameActivityAssetsHandler{
		iDao: dao.NewGameActivityAssetsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameActivityAssetsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameActivityAssets
// @Summary Create a new gameActivityAssets
// @Description Creates a new gameActivityAssets entity using the provided data in the request body.
// @Tags gameActivityAssets
// @Accept json
// @Produce json
// @Param data body types.CreateGameActivityAssetsRequest true "gameActivityAssets information"
// @Success 200 {object} types.CreateGameActivityAssetsReply{}
// @Router /api/v1/gameActivityAssets [post]
// @Security BearerAuth
func (h *gameActivityAssetsHandler) Create(c *gin.Context) {
	form := &types.CreateGameActivityAssetsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameActivityAssets := &model.GameActivityAssets{}
	err = copier.Copy(gameActivityAssets, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameActivityAssets)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameActivityAssets)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameActivityAssets.ID})
}

// DeleteByID delete a gameActivityAssets by id
// @Summary Delete a gameActivityAssets by id
// @Description Deletes a existing gameActivityAssets identified by the given id in the path.
// @Tags gameActivityAssets
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameActivityAssetsByIDReply{}
// @Router /api/v1/gameActivityAssets/{id} [delete]
// @Security BearerAuth
func (h *gameActivityAssetsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameActivityAssetsIDFromPath(c)
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

// UpdateByID update a gameActivityAssets by id
// @Summary Update a gameActivityAssets by id
// @Description Updates the specified gameActivityAssets by given id in the path, support partial update.
// @Tags gameActivityAssets
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameActivityAssetsByIDRequest true "gameActivityAssets information"
// @Success 200 {object} types.UpdateGameActivityAssetsByIDReply{}
// @Router /api/v1/gameActivityAssets/{id} [put]
// @Security BearerAuth
func (h *gameActivityAssetsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameActivityAssetsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameActivityAssetsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameActivityAssets := &model.GameActivityAssets{}
	err = copier.Copy(gameActivityAssets, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameActivityAssets)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameActivityAssets)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameActivityAssets by id
// @Summary Get a gameActivityAssets by id
// @Description Gets detailed information of a gameActivityAssets specified by the given id in the path.
// @Tags gameActivityAssets
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameActivityAssetsByIDReply{}
// @Router /api/v1/gameActivityAssets/{id} [get]
// @Security BearerAuth
func (h *gameActivityAssetsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameActivityAssetsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameActivityAssets, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameActivityAssetsObjDetail{}
	err = copier.Copy(data, gameActivityAssets)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameActivityAssets)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameActivityAssets": data})
}

// List get a paginated list of gameActivityAssetss by custom conditions
// @Summary Get a paginated list of gameActivityAssetss by custom conditions
// @Description Returns a paginated list of gameActivityAssets based on query filters, including page number and size.
// @Tags gameActivityAssets
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameActivityAssetssReply{}
// @Router /api/v1/gameActivityAssets/list [post]
// @Security BearerAuth
func (h *gameActivityAssetsHandler) List(c *gin.Context) {
	form := &types.ListGameActivityAssetssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameActivityAssetss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameActivityAssetss(gameActivityAssetss)
	if err != nil {
		response.Error(c, ecode.ErrListGameActivityAssets)
		return
	}

	response.Success(c, gin.H{
		"gameActivityAssetss": data,
		"total":        total,
	})
}

func getGameActivityAssetsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameActivityAssets(gameActivityAssets *model.GameActivityAssets) (*types.GameActivityAssetsObjDetail, error) {
	data := &types.GameActivityAssetsObjDetail{}
	err := copier.Copy(data, gameActivityAssets)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameActivityAssetss(fromValues []*model.GameActivityAssets) ([]*types.GameActivityAssetsObjDetail, error) {
	toValues := []*types.GameActivityAssetsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameActivityAssets(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
