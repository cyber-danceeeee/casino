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

var _ GameInvitationWheelPrizesHandler = (*gameInvitationWheelPrizesHandler)(nil)

// GameInvitationWheelPrizesHandler defining the handler interface
type GameInvitationWheelPrizesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameInvitationWheelPrizesHandler struct {
	iDao dao.GameInvitationWheelPrizesDao
}

// NewGameInvitationWheelPrizesHandler creating the handler interface
func NewGameInvitationWheelPrizesHandler() GameInvitationWheelPrizesHandler {
	return &gameInvitationWheelPrizesHandler{
		iDao: dao.NewGameInvitationWheelPrizesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameInvitationWheelPrizesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameInvitationWheelPrizes
// @Summary Create a new gameInvitationWheelPrizes
// @Description Creates a new gameInvitationWheelPrizes entity using the provided data in the request body.
// @Tags gameInvitationWheelPrizes
// @Accept json
// @Produce json
// @Param data body types.CreateGameInvitationWheelPrizesRequest true "gameInvitationWheelPrizes information"
// @Success 200 {object} types.CreateGameInvitationWheelPrizesReply{}
// @Router /api/v1/gameInvitationWheelPrizes [post]
// @Security BearerAuth
func (h *gameInvitationWheelPrizesHandler) Create(c *gin.Context) {
	form := &types.CreateGameInvitationWheelPrizesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameInvitationWheelPrizes := &model.GameInvitationWheelPrizes{}
	err = copier.Copy(gameInvitationWheelPrizes, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameInvitationWheelPrizes)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameInvitationWheelPrizes)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameInvitationWheelPrizes.ID})
}

// DeleteByID delete a gameInvitationWheelPrizes by id
// @Summary Delete a gameInvitationWheelPrizes by id
// @Description Deletes a existing gameInvitationWheelPrizes identified by the given id in the path.
// @Tags gameInvitationWheelPrizes
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameInvitationWheelPrizesByIDReply{}
// @Router /api/v1/gameInvitationWheelPrizes/{id} [delete]
// @Security BearerAuth
func (h *gameInvitationWheelPrizesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationWheelPrizesIDFromPath(c)
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

// UpdateByID update a gameInvitationWheelPrizes by id
// @Summary Update a gameInvitationWheelPrizes by id
// @Description Updates the specified gameInvitationWheelPrizes by given id in the path, support partial update.
// @Tags gameInvitationWheelPrizes
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameInvitationWheelPrizesByIDRequest true "gameInvitationWheelPrizes information"
// @Success 200 {object} types.UpdateGameInvitationWheelPrizesByIDReply{}
// @Router /api/v1/gameInvitationWheelPrizes/{id} [put]
// @Security BearerAuth
func (h *gameInvitationWheelPrizesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationWheelPrizesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameInvitationWheelPrizesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameInvitationWheelPrizes := &model.GameInvitationWheelPrizes{}
	err = copier.Copy(gameInvitationWheelPrizes, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameInvitationWheelPrizes)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameInvitationWheelPrizes)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameInvitationWheelPrizes by id
// @Summary Get a gameInvitationWheelPrizes by id
// @Description Gets detailed information of a gameInvitationWheelPrizes specified by the given id in the path.
// @Tags gameInvitationWheelPrizes
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameInvitationWheelPrizesByIDReply{}
// @Router /api/v1/gameInvitationWheelPrizes/{id} [get]
// @Security BearerAuth
func (h *gameInvitationWheelPrizesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationWheelPrizesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameInvitationWheelPrizes, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameInvitationWheelPrizesObjDetail{}
	err = copier.Copy(data, gameInvitationWheelPrizes)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameInvitationWheelPrizes)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameInvitationWheelPrizes": data})
}

// List get a paginated list of gameInvitationWheelPrizess by custom conditions
// @Summary Get a paginated list of gameInvitationWheelPrizess by custom conditions
// @Description Returns a paginated list of gameInvitationWheelPrizes based on query filters, including page number and size.
// @Tags gameInvitationWheelPrizes
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameInvitationWheelPrizessReply{}
// @Router /api/v1/gameInvitationWheelPrizes/list [post]
// @Security BearerAuth
func (h *gameInvitationWheelPrizesHandler) List(c *gin.Context) {
	form := &types.ListGameInvitationWheelPrizessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameInvitationWheelPrizess, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameInvitationWheelPrizess(gameInvitationWheelPrizess)
	if err != nil {
		response.Error(c, ecode.ErrListGameInvitationWheelPrizes)
		return
	}

	response.Success(c, gin.H{
		"gameInvitationWheelPrizess": data,
		"total":                      total,
	})
}

func getGameInvitationWheelPrizesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameInvitationWheelPrizes(gameInvitationWheelPrizes *model.GameInvitationWheelPrizes) (*types.GameInvitationWheelPrizesObjDetail, error) {
	data := &types.GameInvitationWheelPrizesObjDetail{}
	err := copier.Copy(data, gameInvitationWheelPrizes)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameInvitationWheelPrizess(fromValues []*model.GameInvitationWheelPrizes) ([]*types.GameInvitationWheelPrizesObjDetail, error) {
	toValues := []*types.GameInvitationWheelPrizesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameInvitationWheelPrizes(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
