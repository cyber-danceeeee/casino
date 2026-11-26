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

var _ GameHomeFeatureCardsHandler = (*gameHomeFeatureCardsHandler)(nil)

// GameHomeFeatureCardsHandler defining the handler interface
type GameHomeFeatureCardsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameHomeFeatureCardsHandler struct {
	iDao dao.GameHomeFeatureCardsDao
}

// NewGameHomeFeatureCardsHandler creating the handler interface
func NewGameHomeFeatureCardsHandler() GameHomeFeatureCardsHandler {
	return &gameHomeFeatureCardsHandler{
		iDao: dao.NewGameHomeFeatureCardsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameHomeFeatureCardsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameHomeFeatureCards
// @Summary Create a new gameHomeFeatureCards
// @Description Creates a new gameHomeFeatureCards entity using the provided data in the request body.
// @Tags gameHomeFeatureCards
// @Accept json
// @Produce json
// @Param data body types.CreateGameHomeFeatureCardsRequest true "gameHomeFeatureCards information"
// @Success 200 {object} types.CreateGameHomeFeatureCardsReply{}
// @Router /api/v1/gameHomeFeatureCards [post]
// @Security BearerAuth
func (h *gameHomeFeatureCardsHandler) Create(c *gin.Context) {
	form := &types.CreateGameHomeFeatureCardsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameHomeFeatureCards := &model.GameHomeFeatureCards{}
	err = copier.Copy(gameHomeFeatureCards, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameHomeFeatureCards)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameHomeFeatureCards)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameHomeFeatureCards.ID})
}

// DeleteByID delete a gameHomeFeatureCards by id
// @Summary Delete a gameHomeFeatureCards by id
// @Description Deletes a existing gameHomeFeatureCards identified by the given id in the path.
// @Tags gameHomeFeatureCards
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameHomeFeatureCardsByIDReply{}
// @Router /api/v1/gameHomeFeatureCards/{id} [delete]
// @Security BearerAuth
func (h *gameHomeFeatureCardsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameHomeFeatureCardsIDFromPath(c)
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

// UpdateByID update a gameHomeFeatureCards by id
// @Summary Update a gameHomeFeatureCards by id
// @Description Updates the specified gameHomeFeatureCards by given id in the path, support partial update.
// @Tags gameHomeFeatureCards
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameHomeFeatureCardsByIDRequest true "gameHomeFeatureCards information"
// @Success 200 {object} types.UpdateGameHomeFeatureCardsByIDReply{}
// @Router /api/v1/gameHomeFeatureCards/{id} [put]
// @Security BearerAuth
func (h *gameHomeFeatureCardsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameHomeFeatureCardsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameHomeFeatureCardsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameHomeFeatureCards := &model.GameHomeFeatureCards{}
	err = copier.Copy(gameHomeFeatureCards, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameHomeFeatureCards)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameHomeFeatureCards)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameHomeFeatureCards by id
// @Summary Get a gameHomeFeatureCards by id
// @Description Gets detailed information of a gameHomeFeatureCards specified by the given id in the path.
// @Tags gameHomeFeatureCards
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameHomeFeatureCardsByIDReply{}
// @Router /api/v1/gameHomeFeatureCards/{id} [get]
// @Security BearerAuth
func (h *gameHomeFeatureCardsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameHomeFeatureCardsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameHomeFeatureCards, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameHomeFeatureCardsObjDetail{}
	err = copier.Copy(data, gameHomeFeatureCards)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameHomeFeatureCards)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameHomeFeatureCards": data})
}

// List get a paginated list of gameHomeFeatureCardss by custom conditions
// @Summary Get a paginated list of gameHomeFeatureCardss by custom conditions
// @Description Returns a paginated list of gameHomeFeatureCards based on query filters, including page number and size.
// @Tags gameHomeFeatureCards
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameHomeFeatureCardssReply{}
// @Router /api/v1/gameHomeFeatureCards/list [post]
// @Security BearerAuth
func (h *gameHomeFeatureCardsHandler) List(c *gin.Context) {
	form := &types.ListGameHomeFeatureCardssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameHomeFeatureCardss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameHomeFeatureCardss(gameHomeFeatureCardss)
	if err != nil {
		response.Error(c, ecode.ErrListGameHomeFeatureCards)
		return
	}

	response.Success(c, gin.H{
		"gameHomeFeatureCardss": data,
		"total":        total,
	})
}

func getGameHomeFeatureCardsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameHomeFeatureCards(gameHomeFeatureCards *model.GameHomeFeatureCards) (*types.GameHomeFeatureCardsObjDetail, error) {
	data := &types.GameHomeFeatureCardsObjDetail{}
	err := copier.Copy(data, gameHomeFeatureCards)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameHomeFeatureCardss(fromValues []*model.GameHomeFeatureCards) ([]*types.GameHomeFeatureCardsObjDetail, error) {
	toValues := []*types.GameHomeFeatureCardsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameHomeFeatureCards(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
