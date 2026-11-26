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

var _ GameVipLevelsHandler = (*gameVipLevelsHandler)(nil)

// GameVipLevelsHandler defining the handler interface
type GameVipLevelsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Summary(c *gin.Context)
}

type gameVipLevelsHandler struct {
	iDao dao.GameVipLevelsDao
}

// NewGameVipLevelsHandler creating the handler interface
func NewGameVipLevelsHandler() GameVipLevelsHandler {
	return &gameVipLevelsHandler{
		iDao: dao.NewGameVipLevelsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameVipLevelsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameVipLevels
// @Summary Create a new gameVipLevels
// @Description Creates a new gameVipLevels entity using the provided data in the request body.
// @Tags gameVipLevels
// @Accept json
// @Produce json
// @Param data body types.CreateGameVipLevelsRequest true "gameVipLevels information"
// @Success 200 {object} types.CreateGameVipLevelsReply{}
// @Router /api/v1/gameVipLevels [post]
// @Security BearerAuth
func (h *gameVipLevelsHandler) Create(c *gin.Context) {
	form := &types.CreateGameVipLevelsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameVipLevels := &model.GameVipLevels{}
	err = copier.Copy(gameVipLevels, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameVipLevels)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameVipLevels)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameVipLevels.ID})
}

// DeleteByID delete a gameVipLevels by id
// @Summary Delete a gameVipLevels by id
// @Description Deletes a existing gameVipLevels identified by the given id in the path.
// @Tags gameVipLevels
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameVipLevelsByIDReply{}
// @Router /api/v1/gameVipLevels/{id} [delete]
// @Security BearerAuth
func (h *gameVipLevelsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameVipLevelsIDFromPath(c)
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

// UpdateByID update a gameVipLevels by id
// @Summary Update a gameVipLevels by id
// @Description Updates the specified gameVipLevels by given id in the path, support partial update.
// @Tags gameVipLevels
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameVipLevelsByIDRequest true "gameVipLevels information"
// @Success 200 {object} types.UpdateGameVipLevelsByIDReply{}
// @Router /api/v1/gameVipLevels/{id} [put]
// @Security BearerAuth
func (h *gameVipLevelsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameVipLevelsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameVipLevelsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameVipLevels := &model.GameVipLevels{}
	err = copier.Copy(gameVipLevels, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameVipLevels)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameVipLevels)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameVipLevels by id
// @Summary Get a gameVipLevels by id
// @Description Gets detailed information of a gameVipLevels specified by the given id in the path.
// @Tags gameVipLevels
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameVipLevelsByIDReply{}
// @Router /api/v1/gameVipLevels/{id} [get]
// @Security BearerAuth
func (h *gameVipLevelsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameVipLevelsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameVipLevels, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameVipLevelsObjDetail{}
	err = copier.Copy(data, gameVipLevels)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameVipLevels)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameVipLevels": data})
}

// List get a paginated list of gameVipLevelss by custom conditions
// @Summary Get a paginated list of gameVipLevelss by custom conditions
// @Description Returns a paginated list of gameVipLevels based on query filters, including page number and size.
// @Tags gameVipLevels
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameVipLevelssReply{}
// @Router /api/v1/gameVipLevels/list [post]
// @Security BearerAuth
func (h *gameVipLevelsHandler) List(c *gin.Context) {
	form := &types.ListGameVipLevelssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameVipLevelss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameVipLevelss(gameVipLevelss)
	if err != nil {
		response.Error(c, ecode.ErrListGameVipLevels)
		return
	}

	response.Success(c, gin.H{
		"gameVipLevelss": data,
		"total":          total,
	})
}

func getGameVipLevelsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameVipLevels(gameVipLevels *model.GameVipLevels) (*types.GameVipLevelsObjDetail, error) {
	data := &types.GameVipLevelsObjDetail{}
	err := copier.Copy(data, gameVipLevels)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameVipLevelss(fromValues []*model.GameVipLevels) ([]*types.GameVipLevelsObjDetail, error) {
	toValues := []*types.GameVipLevelsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameVipLevels(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
