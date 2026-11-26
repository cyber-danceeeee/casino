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

var _ GameVipRequirementsHandler = (*gameVipRequirementsHandler)(nil)

// GameVipRequirementsHandler defining the handler interface
type GameVipRequirementsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameVipRequirementsHandler struct {
	iDao dao.GameVipRequirementsDao
}

// NewGameVipRequirementsHandler creating the handler interface
func NewGameVipRequirementsHandler() GameVipRequirementsHandler {
	return &gameVipRequirementsHandler{
		iDao: dao.NewGameVipRequirementsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameVipRequirementsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameVipRequirements
// @Summary Create a new gameVipRequirements
// @Description Creates a new gameVipRequirements entity using the provided data in the request body.
// @Tags gameVipRequirements
// @Accept json
// @Produce json
// @Param data body types.CreateGameVipRequirementsRequest true "gameVipRequirements information"
// @Success 200 {object} types.CreateGameVipRequirementsReply{}
// @Router /api/v1/gameVipRequirements [post]
// @Security BearerAuth
func (h *gameVipRequirementsHandler) Create(c *gin.Context) {
	form := &types.CreateGameVipRequirementsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameVipRequirements := &model.GameVipRequirements{}
	err = copier.Copy(gameVipRequirements, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameVipRequirements)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameVipRequirements)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameVipRequirements.ID})
}

// DeleteByID delete a gameVipRequirements by id
// @Summary Delete a gameVipRequirements by id
// @Description Deletes a existing gameVipRequirements identified by the given id in the path.
// @Tags gameVipRequirements
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameVipRequirementsByIDReply{}
// @Router /api/v1/gameVipRequirements/{id} [delete]
// @Security BearerAuth
func (h *gameVipRequirementsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameVipRequirementsIDFromPath(c)
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

// UpdateByID update a gameVipRequirements by id
// @Summary Update a gameVipRequirements by id
// @Description Updates the specified gameVipRequirements by given id in the path, support partial update.
// @Tags gameVipRequirements
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameVipRequirementsByIDRequest true "gameVipRequirements information"
// @Success 200 {object} types.UpdateGameVipRequirementsByIDReply{}
// @Router /api/v1/gameVipRequirements/{id} [put]
// @Security BearerAuth
func (h *gameVipRequirementsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameVipRequirementsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameVipRequirementsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameVipRequirements := &model.GameVipRequirements{}
	err = copier.Copy(gameVipRequirements, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameVipRequirements)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameVipRequirements)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameVipRequirements by id
// @Summary Get a gameVipRequirements by id
// @Description Gets detailed information of a gameVipRequirements specified by the given id in the path.
// @Tags gameVipRequirements
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameVipRequirementsByIDReply{}
// @Router /api/v1/gameVipRequirements/{id} [get]
// @Security BearerAuth
func (h *gameVipRequirementsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameVipRequirementsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameVipRequirements, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameVipRequirementsObjDetail{}
	err = copier.Copy(data, gameVipRequirements)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameVipRequirements)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameVipRequirements": data})
}

// List get a paginated list of gameVipRequirementss by custom conditions
// @Summary Get a paginated list of gameVipRequirementss by custom conditions
// @Description Returns a paginated list of gameVipRequirements based on query filters, including page number and size.
// @Tags gameVipRequirements
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameVipRequirementssReply{}
// @Router /api/v1/gameVipRequirements/list [post]
// @Security BearerAuth
func (h *gameVipRequirementsHandler) List(c *gin.Context) {
	form := &types.ListGameVipRequirementssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameVipRequirementss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameVipRequirementss(gameVipRequirementss)
	if err != nil {
		response.Error(c, ecode.ErrListGameVipRequirements)
		return
	}

	response.Success(c, gin.H{
		"gameVipRequirementss": data,
		"total":        total,
	})
}

func getGameVipRequirementsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameVipRequirements(gameVipRequirements *model.GameVipRequirements) (*types.GameVipRequirementsObjDetail, error) {
	data := &types.GameVipRequirementsObjDetail{}
	err := copier.Copy(data, gameVipRequirements)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameVipRequirementss(fromValues []*model.GameVipRequirements) ([]*types.GameVipRequirementsObjDetail, error) {
	toValues := []*types.GameVipRequirementsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameVipRequirements(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
