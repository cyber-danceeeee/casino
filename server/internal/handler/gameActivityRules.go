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

var _ GameActivityRulesHandler = (*gameActivityRulesHandler)(nil)

// GameActivityRulesHandler defining the handler interface
type GameActivityRulesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameActivityRulesHandler struct {
	iDao dao.GameActivityRulesDao
}

// NewGameActivityRulesHandler creating the handler interface
func NewGameActivityRulesHandler() GameActivityRulesHandler {
	return &gameActivityRulesHandler{
		iDao: dao.NewGameActivityRulesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameActivityRulesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameActivityRules
// @Summary Create a new gameActivityRules
// @Description Creates a new gameActivityRules entity using the provided data in the request body.
// @Tags gameActivityRules
// @Accept json
// @Produce json
// @Param data body types.CreateGameActivityRulesRequest true "gameActivityRules information"
// @Success 200 {object} types.CreateGameActivityRulesReply{}
// @Router /api/v1/gameActivityRules [post]
// @Security BearerAuth
func (h *gameActivityRulesHandler) Create(c *gin.Context) {
	form := &types.CreateGameActivityRulesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameActivityRules := &model.GameActivityRules{}
	err = copier.Copy(gameActivityRules, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameActivityRules)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameActivityRules)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameActivityRules.ID})
}

// DeleteByID delete a gameActivityRules by id
// @Summary Delete a gameActivityRules by id
// @Description Deletes a existing gameActivityRules identified by the given id in the path.
// @Tags gameActivityRules
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameActivityRulesByIDReply{}
// @Router /api/v1/gameActivityRules/{id} [delete]
// @Security BearerAuth
func (h *gameActivityRulesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameActivityRulesIDFromPath(c)
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

// UpdateByID update a gameActivityRules by id
// @Summary Update a gameActivityRules by id
// @Description Updates the specified gameActivityRules by given id in the path, support partial update.
// @Tags gameActivityRules
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameActivityRulesByIDRequest true "gameActivityRules information"
// @Success 200 {object} types.UpdateGameActivityRulesByIDReply{}
// @Router /api/v1/gameActivityRules/{id} [put]
// @Security BearerAuth
func (h *gameActivityRulesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameActivityRulesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameActivityRulesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameActivityRules := &model.GameActivityRules{}
	err = copier.Copy(gameActivityRules, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameActivityRules)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameActivityRules)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameActivityRules by id
// @Summary Get a gameActivityRules by id
// @Description Gets detailed information of a gameActivityRules specified by the given id in the path.
// @Tags gameActivityRules
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameActivityRulesByIDReply{}
// @Router /api/v1/gameActivityRules/{id} [get]
// @Security BearerAuth
func (h *gameActivityRulesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameActivityRulesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameActivityRules, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameActivityRulesObjDetail{}
	err = copier.Copy(data, gameActivityRules)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameActivityRules)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameActivityRules": data})
}

// List get a paginated list of gameActivityRuless by custom conditions
// @Summary Get a paginated list of gameActivityRuless by custom conditions
// @Description Returns a paginated list of gameActivityRules based on query filters, including page number and size.
// @Tags gameActivityRules
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameActivityRulessReply{}
// @Router /api/v1/gameActivityRules/list [post]
// @Security BearerAuth
func (h *gameActivityRulesHandler) List(c *gin.Context) {
	form := &types.ListGameActivityRulessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameActivityRuless, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameActivityRuless(gameActivityRuless)
	if err != nil {
		response.Error(c, ecode.ErrListGameActivityRules)
		return
	}

	response.Success(c, gin.H{
		"gameActivityRuless": data,
		"total":              total,
	})
}

func getGameActivityRulesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameActivityRules(gameActivityRules *model.GameActivityRules) (*types.GameActivityRulesObjDetail, error) {
	data := &types.GameActivityRulesObjDetail{}
	err := copier.Copy(data, gameActivityRules)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameActivityRuless(fromValues []*model.GameActivityRules) ([]*types.GameActivityRulesObjDetail, error) {
	toValues := []*types.GameActivityRulesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameActivityRules(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
