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

var _ GameGameCategoryRelationsHandler = (*gameGameCategoryRelationsHandler)(nil)

// GameGameCategoryRelationsHandler defining the handler interface
type GameGameCategoryRelationsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameGameCategoryRelationsHandler struct {
	iDao dao.GameGameCategoryRelationsDao
}

// NewGameGameCategoryRelationsHandler creating the handler interface
func NewGameGameCategoryRelationsHandler() GameGameCategoryRelationsHandler {
	return &gameGameCategoryRelationsHandler{
		iDao: dao.NewGameGameCategoryRelationsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameGameCategoryRelationsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameGameCategoryRelations
// @Summary Create a new gameGameCategoryRelations
// @Description Creates a new gameGameCategoryRelations entity using the provided data in the request body.
// @Tags gameGameCategoryRelations
// @Accept json
// @Produce json
// @Param data body types.CreateGameGameCategoryRelationsRequest true "gameGameCategoryRelations information"
// @Success 200 {object} types.CreateGameGameCategoryRelationsReply{}
// @Router /api/v1/gameGameCategoryRelations [post]
// @Security BearerAuth
func (h *gameGameCategoryRelationsHandler) Create(c *gin.Context) {
	form := &types.CreateGameGameCategoryRelationsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameGameCategoryRelations := &model.GameGameCategoryRelations{}
	err = copier.Copy(gameGameCategoryRelations, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameGameCategoryRelations)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameGameCategoryRelations)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameGameCategoryRelations.ID})
}

// DeleteByID delete a gameGameCategoryRelations by id
// @Summary Delete a gameGameCategoryRelations by id
// @Description Deletes a existing gameGameCategoryRelations identified by the given id in the path.
// @Tags gameGameCategoryRelations
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameGameCategoryRelationsByIDReply{}
// @Router /api/v1/gameGameCategoryRelations/{id} [delete]
// @Security BearerAuth
func (h *gameGameCategoryRelationsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameGameCategoryRelationsIDFromPath(c)
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

// UpdateByID update a gameGameCategoryRelations by id
// @Summary Update a gameGameCategoryRelations by id
// @Description Updates the specified gameGameCategoryRelations by given id in the path, support partial update.
// @Tags gameGameCategoryRelations
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameGameCategoryRelationsByIDRequest true "gameGameCategoryRelations information"
// @Success 200 {object} types.UpdateGameGameCategoryRelationsByIDReply{}
// @Router /api/v1/gameGameCategoryRelations/{id} [put]
// @Security BearerAuth
func (h *gameGameCategoryRelationsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameGameCategoryRelationsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameGameCategoryRelationsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameGameCategoryRelations := &model.GameGameCategoryRelations{}
	err = copier.Copy(gameGameCategoryRelations, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameGameCategoryRelations)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameGameCategoryRelations)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameGameCategoryRelations by id
// @Summary Get a gameGameCategoryRelations by id
// @Description Gets detailed information of a gameGameCategoryRelations specified by the given id in the path.
// @Tags gameGameCategoryRelations
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameGameCategoryRelationsByIDReply{}
// @Router /api/v1/gameGameCategoryRelations/{id} [get]
// @Security BearerAuth
func (h *gameGameCategoryRelationsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameGameCategoryRelationsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameGameCategoryRelations, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameGameCategoryRelationsObjDetail{}
	err = copier.Copy(data, gameGameCategoryRelations)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameGameCategoryRelations)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameGameCategoryRelations": data})
}

// List get a paginated list of gameGameCategoryRelationss by custom conditions
// @Summary Get a paginated list of gameGameCategoryRelationss by custom conditions
// @Description Returns a paginated list of gameGameCategoryRelations based on query filters, including page number and size.
// @Tags gameGameCategoryRelations
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameGameCategoryRelationssReply{}
// @Router /api/v1/gameGameCategoryRelations/list [post]
// @Security BearerAuth
func (h *gameGameCategoryRelationsHandler) List(c *gin.Context) {
	form := &types.ListGameGameCategoryRelationssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameGameCategoryRelationss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameGameCategoryRelationss(gameGameCategoryRelationss)
	if err != nil {
		response.Error(c, ecode.ErrListGameGameCategoryRelations)
		return
	}

	response.Success(c, gin.H{
		"gameGameCategoryRelationss": data,
		"total":        total,
	})
}

func getGameGameCategoryRelationsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameGameCategoryRelations(gameGameCategoryRelations *model.GameGameCategoryRelations) (*types.GameGameCategoryRelationsObjDetail, error) {
	data := &types.GameGameCategoryRelationsObjDetail{}
	err := copier.Copy(data, gameGameCategoryRelations)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameGameCategoryRelationss(fromValues []*model.GameGameCategoryRelations) ([]*types.GameGameCategoryRelationsObjDetail, error) {
	toValues := []*types.GameGameCategoryRelationsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameGameCategoryRelations(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
