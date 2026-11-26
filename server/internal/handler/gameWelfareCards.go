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

var _ GameWelfareCardsHandler = (*gameWelfareCardsHandler)(nil)

// GameWelfareCardsHandler defining the handler interface
type GameWelfareCardsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameWelfareCardsHandler struct {
	iDao dao.GameWelfareCardsDao
}

// NewGameWelfareCardsHandler creating the handler interface
func NewGameWelfareCardsHandler() GameWelfareCardsHandler {
	return &gameWelfareCardsHandler{
		iDao: dao.NewGameWelfareCardsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameWelfareCardsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameWelfareCards
// @Summary Create a new gameWelfareCards
// @Description Creates a new gameWelfareCards entity using the provided data in the request body.
// @Tags gameWelfareCards
// @Accept json
// @Produce json
// @Param data body types.CreateGameWelfareCardsRequest true "gameWelfareCards information"
// @Success 200 {object} types.CreateGameWelfareCardsReply{}
// @Router /api/v1/gameWelfareCards [post]
// @Security BearerAuth
func (h *gameWelfareCardsHandler) Create(c *gin.Context) {
	form := &types.CreateGameWelfareCardsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameWelfareCards := &model.GameWelfareCards{}
	err = copier.Copy(gameWelfareCards, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameWelfareCards)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameWelfareCards)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameWelfareCards.ID})
}

// DeleteByID delete a gameWelfareCards by id
// @Summary Delete a gameWelfareCards by id
// @Description Deletes a existing gameWelfareCards identified by the given id in the path.
// @Tags gameWelfareCards
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameWelfareCardsByIDReply{}
// @Router /api/v1/gameWelfareCards/{id} [delete]
// @Security BearerAuth
func (h *gameWelfareCardsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameWelfareCardsIDFromPath(c)
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

// UpdateByID update a gameWelfareCards by id
// @Summary Update a gameWelfareCards by id
// @Description Updates the specified gameWelfareCards by given id in the path, support partial update.
// @Tags gameWelfareCards
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameWelfareCardsByIDRequest true "gameWelfareCards information"
// @Success 200 {object} types.UpdateGameWelfareCardsByIDReply{}
// @Router /api/v1/gameWelfareCards/{id} [put]
// @Security BearerAuth
func (h *gameWelfareCardsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameWelfareCardsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameWelfareCardsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameWelfareCards := &model.GameWelfareCards{}
	err = copier.Copy(gameWelfareCards, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameWelfareCards)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameWelfareCards)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameWelfareCards by id
// @Summary Get a gameWelfareCards by id
// @Description Gets detailed information of a gameWelfareCards specified by the given id in the path.
// @Tags gameWelfareCards
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameWelfareCardsByIDReply{}
// @Router /api/v1/gameWelfareCards/{id} [get]
// @Security BearerAuth
func (h *gameWelfareCardsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameWelfareCardsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameWelfareCards, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameWelfareCardsObjDetail{}
	err = copier.Copy(data, gameWelfareCards)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameWelfareCards)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameWelfareCards": data})
}

// List get a paginated list of gameWelfareCardss by custom conditions
// @Summary Get a paginated list of gameWelfareCardss by custom conditions
// @Description Returns a paginated list of gameWelfareCards based on query filters, including page number and size.
// @Tags gameWelfareCards
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameWelfareCardssReply{}
// @Router /api/v1/gameWelfareCards/list [post]
// @Security BearerAuth
func (h *gameWelfareCardsHandler) List(c *gin.Context) {
	form := &types.ListGameWelfareCardssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameWelfareCardss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameWelfareCardss(gameWelfareCardss)
	if err != nil {
		response.Error(c, ecode.ErrListGameWelfareCards)
		return
	}

	response.Success(c, gin.H{
		"gameWelfareCardss": data,
		"total":             total,
	})
}

func getGameWelfareCardsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameWelfareCards(gameWelfareCards *model.GameWelfareCards) (*types.GameWelfareCardsObjDetail, error) {
	data := &types.GameWelfareCardsObjDetail{}
	err := copier.Copy(data, gameWelfareCards)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameWelfareCardss(fromValues []*model.GameWelfareCards) ([]*types.GameWelfareCardsObjDetail, error) {
	toValues := []*types.GameWelfareCardsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameWelfareCards(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
