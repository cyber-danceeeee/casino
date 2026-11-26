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

var _ GameActivityRewardRulesHandler = (*gameActivityRewardRulesHandler)(nil)

// GameActivityRewardRulesHandler defining the handler interface
type GameActivityRewardRulesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameActivityRewardRulesHandler struct {
	iDao dao.GameActivityRewardRulesDao
}

// NewGameActivityRewardRulesHandler creating the handler interface
func NewGameActivityRewardRulesHandler() GameActivityRewardRulesHandler {
	return &gameActivityRewardRulesHandler{
		iDao: dao.NewGameActivityRewardRulesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameActivityRewardRulesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameActivityRewardRules
// @Summary Create a new gameActivityRewardRules
// @Description Creates a new gameActivityRewardRules entity using the provided data in the request body.
// @Tags gameActivityRewardRules
// @Accept json
// @Produce json
// @Param data body types.CreateGameActivityRewardRulesRequest true "gameActivityRewardRules information"
// @Success 200 {object} types.CreateGameActivityRewardRulesReply{}
// @Router /api/v1/gameActivityRewardRules [post]
// @Security BearerAuth
func (h *gameActivityRewardRulesHandler) Create(c *gin.Context) {
	form := &types.CreateGameActivityRewardRulesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameActivityRewardRules := &model.GameActivityRewardRules{}
	err = copier.Copy(gameActivityRewardRules, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameActivityRewardRules)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameActivityRewardRules)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameActivityRewardRules.ID})
}

// DeleteByID delete a gameActivityRewardRules by id
// @Summary Delete a gameActivityRewardRules by id
// @Description Deletes a existing gameActivityRewardRules identified by the given id in the path.
// @Tags gameActivityRewardRules
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameActivityRewardRulesByIDReply{}
// @Router /api/v1/gameActivityRewardRules/{id} [delete]
// @Security BearerAuth
func (h *gameActivityRewardRulesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameActivityRewardRulesIDFromPath(c)
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

// UpdateByID update a gameActivityRewardRules by id
// @Summary Update a gameActivityRewardRules by id
// @Description Updates the specified gameActivityRewardRules by given id in the path, support partial update.
// @Tags gameActivityRewardRules
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameActivityRewardRulesByIDRequest true "gameActivityRewardRules information"
// @Success 200 {object} types.UpdateGameActivityRewardRulesByIDReply{}
// @Router /api/v1/gameActivityRewardRules/{id} [put]
// @Security BearerAuth
func (h *gameActivityRewardRulesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameActivityRewardRulesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameActivityRewardRulesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameActivityRewardRules := &model.GameActivityRewardRules{}
	err = copier.Copy(gameActivityRewardRules, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameActivityRewardRules)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameActivityRewardRules)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameActivityRewardRules by id
// @Summary Get a gameActivityRewardRules by id
// @Description Gets detailed information of a gameActivityRewardRules specified by the given id in the path.
// @Tags gameActivityRewardRules
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameActivityRewardRulesByIDReply{}
// @Router /api/v1/gameActivityRewardRules/{id} [get]
// @Security BearerAuth
func (h *gameActivityRewardRulesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameActivityRewardRulesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameActivityRewardRules, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameActivityRewardRulesObjDetail{}
	err = copier.Copy(data, gameActivityRewardRules)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameActivityRewardRules)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameActivityRewardRules": data})
}

// List get a paginated list of gameActivityRewardRuless by custom conditions
// @Summary Get a paginated list of gameActivityRewardRuless by custom conditions
// @Description Returns a paginated list of gameActivityRewardRules based on query filters, including page number and size.
// @Tags gameActivityRewardRules
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameActivityRewardRulessReply{}
// @Router /api/v1/gameActivityRewardRules/list [post]
// @Security BearerAuth
func (h *gameActivityRewardRulesHandler) List(c *gin.Context) {
	form := &types.ListGameActivityRewardRulessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameActivityRewardRuless, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameActivityRewardRuless(gameActivityRewardRuless)
	if err != nil {
		response.Error(c, ecode.ErrListGameActivityRewardRules)
		return
	}

	response.Success(c, gin.H{
		"gameActivityRewardRuless": data,
		"total":        total,
	})
}

func getGameActivityRewardRulesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameActivityRewardRules(gameActivityRewardRules *model.GameActivityRewardRules) (*types.GameActivityRewardRulesObjDetail, error) {
	data := &types.GameActivityRewardRulesObjDetail{}
	err := copier.Copy(data, gameActivityRewardRules)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameActivityRewardRuless(fromValues []*model.GameActivityRewardRules) ([]*types.GameActivityRewardRulesObjDetail, error) {
	toValues := []*types.GameActivityRewardRulesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameActivityRewardRules(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
