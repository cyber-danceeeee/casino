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

var _ GameUserWelfareCardsHandler = (*gameUserWelfareCardsHandler)(nil)

// GameUserWelfareCardsHandler defining the handler interface
type GameUserWelfareCardsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameUserWelfareCardsHandler struct {
	iDao dao.GameUserWelfareCardsDao
}

// NewGameUserWelfareCardsHandler creating the handler interface
func NewGameUserWelfareCardsHandler() GameUserWelfareCardsHandler {
	return &gameUserWelfareCardsHandler{
		iDao: dao.NewGameUserWelfareCardsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameUserWelfareCardsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameUserWelfareCards
// @Summary Create a new gameUserWelfareCards
// @Description Creates a new gameUserWelfareCards entity using the provided data in the request body.
// @Tags gameUserWelfareCards
// @Accept json
// @Produce json
// @Param data body types.CreateGameUserWelfareCardsRequest true "gameUserWelfareCards information"
// @Success 200 {object} types.CreateGameUserWelfareCardsReply{}
// @Router /api/v1/gameUserWelfareCards [post]
// @Security BearerAuth
func (h *gameUserWelfareCardsHandler) Create(c *gin.Context) {
	form := &types.CreateGameUserWelfareCardsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameUserWelfareCards := &model.GameUserWelfareCards{}
	err = copier.Copy(gameUserWelfareCards, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameUserWelfareCards)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameUserWelfareCards)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameUserWelfareCards.ID})
}

// DeleteByID delete a gameUserWelfareCards by id
// @Summary Delete a gameUserWelfareCards by id
// @Description Deletes a existing gameUserWelfareCards identified by the given id in the path.
// @Tags gameUserWelfareCards
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameUserWelfareCardsByIDReply{}
// @Router /api/v1/gameUserWelfareCards/{id} [delete]
// @Security BearerAuth
func (h *gameUserWelfareCardsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameUserWelfareCardsIDFromPath(c)
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

// UpdateByID update a gameUserWelfareCards by id
// @Summary Update a gameUserWelfareCards by id
// @Description Updates the specified gameUserWelfareCards by given id in the path, support partial update.
// @Tags gameUserWelfareCards
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameUserWelfareCardsByIDRequest true "gameUserWelfareCards information"
// @Success 200 {object} types.UpdateGameUserWelfareCardsByIDReply{}
// @Router /api/v1/gameUserWelfareCards/{id} [put]
// @Security BearerAuth
func (h *gameUserWelfareCardsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameUserWelfareCardsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameUserWelfareCardsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameUserWelfareCards := &model.GameUserWelfareCards{}
	err = copier.Copy(gameUserWelfareCards, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameUserWelfareCards)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameUserWelfareCards)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameUserWelfareCards by id
// @Summary Get a gameUserWelfareCards by id
// @Description Gets detailed information of a gameUserWelfareCards specified by the given id in the path.
// @Tags gameUserWelfareCards
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameUserWelfareCardsByIDReply{}
// @Router /api/v1/gameUserWelfareCards/{id} [get]
// @Security BearerAuth
func (h *gameUserWelfareCardsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameUserWelfareCardsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserWelfareCards, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameUserWelfareCardsObjDetail{}
	err = copier.Copy(data, gameUserWelfareCards)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameUserWelfareCards)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameUserWelfareCards": data})
}

// List get a paginated list of gameUserWelfareCardss by custom conditions
// @Summary Get a paginated list of gameUserWelfareCardss by custom conditions
// @Description Returns a paginated list of gameUserWelfareCards based on query filters, including page number and size.
// @Tags gameUserWelfareCards
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameUserWelfareCardssReply{}
// @Router /api/v1/gameUserWelfareCards/list [post]
// @Security BearerAuth
func (h *gameUserWelfareCardsHandler) List(c *gin.Context) {
	form := &types.ListGameUserWelfareCardssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserWelfareCardss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameUserWelfareCardss(gameUserWelfareCardss)
	if err != nil {
		response.Error(c, ecode.ErrListGameUserWelfareCards)
		return
	}

	response.Success(c, gin.H{
		"gameUserWelfareCardss": data,
		"total":                 total,
	})
}

func getGameUserWelfareCardsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameUserWelfareCards(gameUserWelfareCards *model.GameUserWelfareCards) (*types.GameUserWelfareCardsObjDetail, error) {
	data := &types.GameUserWelfareCardsObjDetail{}
	err := copier.Copy(data, gameUserWelfareCards)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameUserWelfareCardss(fromValues []*model.GameUserWelfareCards) ([]*types.GameUserWelfareCardsObjDetail, error) {
	toValues := []*types.GameUserWelfareCardsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameUserWelfareCards(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
