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

var _ GameGameCategoriesHandler = (*gameGameCategoriesHandler)(nil)

// GameGameCategoriesHandler defining the handler interface
type GameGameCategoriesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameGameCategoriesHandler struct {
	iDao dao.GameGameCategoriesDao
}

// NewGameGameCategoriesHandler creating the handler interface
func NewGameGameCategoriesHandler() GameGameCategoriesHandler {
	return &gameGameCategoriesHandler{
		iDao: dao.NewGameGameCategoriesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameGameCategoriesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameGameCategories
// @Summary Create a new gameGameCategories
// @Description Creates a new gameGameCategories entity using the provided data in the request body.
// @Tags gameGameCategories
// @Accept json
// @Produce json
// @Param data body types.CreateGameGameCategoriesRequest true "gameGameCategories information"
// @Success 200 {object} types.CreateGameGameCategoriesReply{}
// @Router /api/v1/gameGameCategories [post]
// @Security BearerAuth
func (h *gameGameCategoriesHandler) Create(c *gin.Context) {
	form := &types.CreateGameGameCategoriesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameGameCategories := &model.GameGameCategories{}
	err = copier.Copy(gameGameCategories, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameGameCategories)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameGameCategories)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameGameCategories.ID})
}

// DeleteByID delete a gameGameCategories by id
// @Summary Delete a gameGameCategories by id
// @Description Deletes a existing gameGameCategories identified by the given id in the path.
// @Tags gameGameCategories
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameGameCategoriesByIDReply{}
// @Router /api/v1/gameGameCategories/{id} [delete]
// @Security BearerAuth
func (h *gameGameCategoriesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameGameCategoriesIDFromPath(c)
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

// UpdateByID update a gameGameCategories by id
// @Summary Update a gameGameCategories by id
// @Description Updates the specified gameGameCategories by given id in the path, support partial update.
// @Tags gameGameCategories
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameGameCategoriesByIDRequest true "gameGameCategories information"
// @Success 200 {object} types.UpdateGameGameCategoriesByIDReply{}
// @Router /api/v1/gameGameCategories/{id} [put]
// @Security BearerAuth
func (h *gameGameCategoriesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameGameCategoriesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameGameCategoriesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameGameCategories := &model.GameGameCategories{}
	err = copier.Copy(gameGameCategories, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameGameCategories)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameGameCategories)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameGameCategories by id
// @Summary Get a gameGameCategories by id
// @Description Gets detailed information of a gameGameCategories specified by the given id in the path.
// @Tags gameGameCategories
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameGameCategoriesByIDReply{}
// @Router /api/v1/gameGameCategories/{id} [get]
// @Security BearerAuth
func (h *gameGameCategoriesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameGameCategoriesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameGameCategories, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameGameCategoriesObjDetail{}
	err = copier.Copy(data, gameGameCategories)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameGameCategories)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameGameCategories": data})
}

// List get a paginated list of gameGameCategoriess by custom conditions
// @Summary Get a paginated list of gameGameCategoriess by custom conditions
// @Description Returns a paginated list of gameGameCategories based on query filters, including page number and size.
// @Tags gameGameCategories
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameGameCategoriessReply{}
// @Router /api/v1/gameGameCategories/list [post]
// @Security BearerAuth
func (h *gameGameCategoriesHandler) List(c *gin.Context) {
	form := &types.ListGameGameCategoriessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameGameCategoriess, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameGameCategoriess(gameGameCategoriess)
	if err != nil {
		response.Error(c, ecode.ErrListGameGameCategories)
		return
	}

	response.Success(c, gin.H{
		"gameGameCategoriess": data,
		"total":               total,
	})
}

func getGameGameCategoriesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameGameCategories(gameGameCategories *model.GameGameCategories) (*types.GameGameCategoriesObjDetail, error) {
	data := &types.GameGameCategoriesObjDetail{}
	err := copier.Copy(data, gameGameCategories)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameGameCategoriess(fromValues []*model.GameGameCategories) ([]*types.GameGameCategoriesObjDetail, error) {
	toValues := []*types.GameGameCategoriesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameGameCategories(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
