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

var _ GameSupportCategoriesHandler = (*gameSupportCategoriesHandler)(nil)

// GameSupportCategoriesHandler defining the handler interface
type GameSupportCategoriesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameSupportCategoriesHandler struct {
	iDao dao.GameSupportCategoriesDao
}

// NewGameSupportCategoriesHandler creating the handler interface
func NewGameSupportCategoriesHandler() GameSupportCategoriesHandler {
	return &gameSupportCategoriesHandler{
		iDao: dao.NewGameSupportCategoriesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameSupportCategoriesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameSupportCategories
// @Summary Create a new gameSupportCategories
// @Description Creates a new gameSupportCategories entity using the provided data in the request body.
// @Tags gameSupportCategories
// @Accept json
// @Produce json
// @Param data body types.CreateGameSupportCategoriesRequest true "gameSupportCategories information"
// @Success 200 {object} types.CreateGameSupportCategoriesReply{}
// @Router /api/v1/gameSupportCategories [post]
// @Security BearerAuth
func (h *gameSupportCategoriesHandler) Create(c *gin.Context) {
	form := &types.CreateGameSupportCategoriesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameSupportCategories := &model.GameSupportCategories{}
	err = copier.Copy(gameSupportCategories, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameSupportCategories)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameSupportCategories)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameSupportCategories.ID})
}

// DeleteByID delete a gameSupportCategories by id
// @Summary Delete a gameSupportCategories by id
// @Description Deletes a existing gameSupportCategories identified by the given id in the path.
// @Tags gameSupportCategories
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameSupportCategoriesByIDReply{}
// @Router /api/v1/gameSupportCategories/{id} [delete]
// @Security BearerAuth
func (h *gameSupportCategoriesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameSupportCategoriesIDFromPath(c)
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

// UpdateByID update a gameSupportCategories by id
// @Summary Update a gameSupportCategories by id
// @Description Updates the specified gameSupportCategories by given id in the path, support partial update.
// @Tags gameSupportCategories
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameSupportCategoriesByIDRequest true "gameSupportCategories information"
// @Success 200 {object} types.UpdateGameSupportCategoriesByIDReply{}
// @Router /api/v1/gameSupportCategories/{id} [put]
// @Security BearerAuth
func (h *gameSupportCategoriesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameSupportCategoriesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameSupportCategoriesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameSupportCategories := &model.GameSupportCategories{}
	err = copier.Copy(gameSupportCategories, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameSupportCategories)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameSupportCategories)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameSupportCategories by id
// @Summary Get a gameSupportCategories by id
// @Description Gets detailed information of a gameSupportCategories specified by the given id in the path.
// @Tags gameSupportCategories
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameSupportCategoriesByIDReply{}
// @Router /api/v1/gameSupportCategories/{id} [get]
// @Security BearerAuth
func (h *gameSupportCategoriesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameSupportCategoriesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameSupportCategories, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameSupportCategoriesObjDetail{}
	err = copier.Copy(data, gameSupportCategories)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameSupportCategories)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameSupportCategories": data})
}

// List get a paginated list of gameSupportCategoriess by custom conditions
// @Summary Get a paginated list of gameSupportCategoriess by custom conditions
// @Description Returns a paginated list of gameSupportCategories based on query filters, including page number and size.
// @Tags gameSupportCategories
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameSupportCategoriessReply{}
// @Router /api/v1/gameSupportCategories/list [post]
// @Security BearerAuth
func (h *gameSupportCategoriesHandler) List(c *gin.Context) {
	form := &types.ListGameSupportCategoriessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameSupportCategoriess, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameSupportCategoriess(gameSupportCategoriess)
	if err != nil {
		response.Error(c, ecode.ErrListGameSupportCategories)
		return
	}

	response.Success(c, gin.H{
		"gameSupportCategoriess": data,
		"total":        total,
	})
}

func getGameSupportCategoriesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameSupportCategories(gameSupportCategories *model.GameSupportCategories) (*types.GameSupportCategoriesObjDetail, error) {
	data := &types.GameSupportCategoriesObjDetail{}
	err := copier.Copy(data, gameSupportCategories)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameSupportCategoriess(fromValues []*model.GameSupportCategories) ([]*types.GameSupportCategoriesObjDetail, error) {
	toValues := []*types.GameSupportCategoriesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameSupportCategories(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
