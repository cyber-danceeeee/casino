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

var _ GameAgentCommissionRulesHandler = (*gameAgentCommissionRulesHandler)(nil)

// GameAgentCommissionRulesHandler defining the handler interface
type GameAgentCommissionRulesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameAgentCommissionRulesHandler struct {
	iDao dao.GameAgentCommissionRulesDao
}

// NewGameAgentCommissionRulesHandler creating the handler interface
func NewGameAgentCommissionRulesHandler() GameAgentCommissionRulesHandler {
	return &gameAgentCommissionRulesHandler{
		iDao: dao.NewGameAgentCommissionRulesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameAgentCommissionRulesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameAgentCommissionRules
// @Summary Create a new gameAgentCommissionRules
// @Description Creates a new gameAgentCommissionRules entity using the provided data in the request body.
// @Tags gameAgentCommissionRules
// @Accept json
// @Produce json
// @Param data body types.CreateGameAgentCommissionRulesRequest true "gameAgentCommissionRules information"
// @Success 200 {object} types.CreateGameAgentCommissionRulesReply{}
// @Router /api/v1/gameAgentCommissionRules [post]
// @Security BearerAuth
func (h *gameAgentCommissionRulesHandler) Create(c *gin.Context) {
	form := &types.CreateGameAgentCommissionRulesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameAgentCommissionRules := &model.GameAgentCommissionRules{}
	err = copier.Copy(gameAgentCommissionRules, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameAgentCommissionRules)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameAgentCommissionRules)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameAgentCommissionRules.ID})
}

// DeleteByID delete a gameAgentCommissionRules by id
// @Summary Delete a gameAgentCommissionRules by id
// @Description Deletes a existing gameAgentCommissionRules identified by the given id in the path.
// @Tags gameAgentCommissionRules
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameAgentCommissionRulesByIDReply{}
// @Router /api/v1/gameAgentCommissionRules/{id} [delete]
// @Security BearerAuth
func (h *gameAgentCommissionRulesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameAgentCommissionRulesIDFromPath(c)
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

// UpdateByID update a gameAgentCommissionRules by id
// @Summary Update a gameAgentCommissionRules by id
// @Description Updates the specified gameAgentCommissionRules by given id in the path, support partial update.
// @Tags gameAgentCommissionRules
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameAgentCommissionRulesByIDRequest true "gameAgentCommissionRules information"
// @Success 200 {object} types.UpdateGameAgentCommissionRulesByIDReply{}
// @Router /api/v1/gameAgentCommissionRules/{id} [put]
// @Security BearerAuth
func (h *gameAgentCommissionRulesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameAgentCommissionRulesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameAgentCommissionRulesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameAgentCommissionRules := &model.GameAgentCommissionRules{}
	err = copier.Copy(gameAgentCommissionRules, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameAgentCommissionRules)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameAgentCommissionRules)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameAgentCommissionRules by id
// @Summary Get a gameAgentCommissionRules by id
// @Description Gets detailed information of a gameAgentCommissionRules specified by the given id in the path.
// @Tags gameAgentCommissionRules
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameAgentCommissionRulesByIDReply{}
// @Router /api/v1/gameAgentCommissionRules/{id} [get]
// @Security BearerAuth
func (h *gameAgentCommissionRulesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameAgentCommissionRulesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameAgentCommissionRules, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameAgentCommissionRulesObjDetail{}
	err = copier.Copy(data, gameAgentCommissionRules)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameAgentCommissionRules)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameAgentCommissionRules": data})
}

// List get a paginated list of gameAgentCommissionRuless by custom conditions
// @Summary Get a paginated list of gameAgentCommissionRuless by custom conditions
// @Description Returns a paginated list of gameAgentCommissionRules based on query filters, including page number and size.
// @Tags gameAgentCommissionRules
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameAgentCommissionRulessReply{}
// @Router /api/v1/gameAgentCommissionRules/list [post]
// @Security BearerAuth
func (h *gameAgentCommissionRulesHandler) List(c *gin.Context) {
	form := &types.ListGameAgentCommissionRulessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameAgentCommissionRuless, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameAgentCommissionRuless(gameAgentCommissionRuless)
	if err != nil {
		response.Error(c, ecode.ErrListGameAgentCommissionRules)
		return
	}

	response.Success(c, gin.H{
		"gameAgentCommissionRuless": data,
		"total":        total,
	})
}

func getGameAgentCommissionRulesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameAgentCommissionRules(gameAgentCommissionRules *model.GameAgentCommissionRules) (*types.GameAgentCommissionRulesObjDetail, error) {
	data := &types.GameAgentCommissionRulesObjDetail{}
	err := copier.Copy(data, gameAgentCommissionRules)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameAgentCommissionRuless(fromValues []*model.GameAgentCommissionRules) ([]*types.GameAgentCommissionRulesObjDetail, error) {
	toValues := []*types.GameAgentCommissionRulesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameAgentCommissionRules(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
