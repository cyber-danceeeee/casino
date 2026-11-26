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

var _ GameInvitationWheelRoundsHandler = (*gameInvitationWheelRoundsHandler)(nil)

// GameInvitationWheelRoundsHandler defining the handler interface
type GameInvitationWheelRoundsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameInvitationWheelRoundsHandler struct {
	iDao dao.GameInvitationWheelRoundsDao
}

// NewGameInvitationWheelRoundsHandler creating the handler interface
func NewGameInvitationWheelRoundsHandler() GameInvitationWheelRoundsHandler {
	return &gameInvitationWheelRoundsHandler{
		iDao: dao.NewGameInvitationWheelRoundsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameInvitationWheelRoundsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameInvitationWheelRounds
// @Summary Create a new gameInvitationWheelRounds
// @Description Creates a new gameInvitationWheelRounds entity using the provided data in the request body.
// @Tags gameInvitationWheelRounds
// @Accept json
// @Produce json
// @Param data body types.CreateGameInvitationWheelRoundsRequest true "gameInvitationWheelRounds information"
// @Success 200 {object} types.CreateGameInvitationWheelRoundsReply{}
// @Router /api/v1/gameInvitationWheelRounds [post]
// @Security BearerAuth
func (h *gameInvitationWheelRoundsHandler) Create(c *gin.Context) {
	form := &types.CreateGameInvitationWheelRoundsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameInvitationWheelRounds := &model.GameInvitationWheelRounds{}
	err = copier.Copy(gameInvitationWheelRounds, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameInvitationWheelRounds)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameInvitationWheelRounds)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameInvitationWheelRounds.ID})
}

// DeleteByID delete a gameInvitationWheelRounds by id
// @Summary Delete a gameInvitationWheelRounds by id
// @Description Deletes a existing gameInvitationWheelRounds identified by the given id in the path.
// @Tags gameInvitationWheelRounds
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameInvitationWheelRoundsByIDReply{}
// @Router /api/v1/gameInvitationWheelRounds/{id} [delete]
// @Security BearerAuth
func (h *gameInvitationWheelRoundsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationWheelRoundsIDFromPath(c)
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

// UpdateByID update a gameInvitationWheelRounds by id
// @Summary Update a gameInvitationWheelRounds by id
// @Description Updates the specified gameInvitationWheelRounds by given id in the path, support partial update.
// @Tags gameInvitationWheelRounds
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameInvitationWheelRoundsByIDRequest true "gameInvitationWheelRounds information"
// @Success 200 {object} types.UpdateGameInvitationWheelRoundsByIDReply{}
// @Router /api/v1/gameInvitationWheelRounds/{id} [put]
// @Security BearerAuth
func (h *gameInvitationWheelRoundsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationWheelRoundsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameInvitationWheelRoundsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameInvitationWheelRounds := &model.GameInvitationWheelRounds{}
	err = copier.Copy(gameInvitationWheelRounds, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameInvitationWheelRounds)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameInvitationWheelRounds)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameInvitationWheelRounds by id
// @Summary Get a gameInvitationWheelRounds by id
// @Description Gets detailed information of a gameInvitationWheelRounds specified by the given id in the path.
// @Tags gameInvitationWheelRounds
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameInvitationWheelRoundsByIDReply{}
// @Router /api/v1/gameInvitationWheelRounds/{id} [get]
// @Security BearerAuth
func (h *gameInvitationWheelRoundsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationWheelRoundsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameInvitationWheelRounds, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameInvitationWheelRoundsObjDetail{}
	err = copier.Copy(data, gameInvitationWheelRounds)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameInvitationWheelRounds)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameInvitationWheelRounds": data})
}

// List get a paginated list of gameInvitationWheelRoundss by custom conditions
// @Summary Get a paginated list of gameInvitationWheelRoundss by custom conditions
// @Description Returns a paginated list of gameInvitationWheelRounds based on query filters, including page number and size.
// @Tags gameInvitationWheelRounds
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameInvitationWheelRoundssReply{}
// @Router /api/v1/gameInvitationWheelRounds/list [post]
// @Security BearerAuth
func (h *gameInvitationWheelRoundsHandler) List(c *gin.Context) {
	form := &types.ListGameInvitationWheelRoundssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameInvitationWheelRoundss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameInvitationWheelRoundss(gameInvitationWheelRoundss)
	if err != nil {
		response.Error(c, ecode.ErrListGameInvitationWheelRounds)
		return
	}

	response.Success(c, gin.H{
		"gameInvitationWheelRoundss": data,
		"total":                      total,
	})
}

func getGameInvitationWheelRoundsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameInvitationWheelRounds(gameInvitationWheelRounds *model.GameInvitationWheelRounds) (*types.GameInvitationWheelRoundsObjDetail, error) {
	data := &types.GameInvitationWheelRoundsObjDetail{}
	err := copier.Copy(data, gameInvitationWheelRounds)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameInvitationWheelRoundss(fromValues []*model.GameInvitationWheelRounds) ([]*types.GameInvitationWheelRoundsObjDetail, error) {
	toValues := []*types.GameInvitationWheelRoundsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameInvitationWheelRounds(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
