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

var _ GameInvitationWheelSpinsHandler = (*gameInvitationWheelSpinsHandler)(nil)

// GameInvitationWheelSpinsHandler defining the handler interface
type GameInvitationWheelSpinsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameInvitationWheelSpinsHandler struct {
	iDao dao.GameInvitationWheelSpinsDao
}

// NewGameInvitationWheelSpinsHandler creating the handler interface
func NewGameInvitationWheelSpinsHandler() GameInvitationWheelSpinsHandler {
	return &gameInvitationWheelSpinsHandler{
		iDao: dao.NewGameInvitationWheelSpinsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameInvitationWheelSpinsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameInvitationWheelSpins
// @Summary Create a new gameInvitationWheelSpins
// @Description Creates a new gameInvitationWheelSpins entity using the provided data in the request body.
// @Tags gameInvitationWheelSpins
// @Accept json
// @Produce json
// @Param data body types.CreateGameInvitationWheelSpinsRequest true "gameInvitationWheelSpins information"
// @Success 200 {object} types.CreateGameInvitationWheelSpinsReply{}
// @Router /api/v1/gameInvitationWheelSpins [post]
// @Security BearerAuth
func (h *gameInvitationWheelSpinsHandler) Create(c *gin.Context) {
	form := &types.CreateGameInvitationWheelSpinsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameInvitationWheelSpins := &model.GameInvitationWheelSpins{}
	err = copier.Copy(gameInvitationWheelSpins, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameInvitationWheelSpins)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameInvitationWheelSpins)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameInvitationWheelSpins.ID})
}

// DeleteByID delete a gameInvitationWheelSpins by id
// @Summary Delete a gameInvitationWheelSpins by id
// @Description Deletes a existing gameInvitationWheelSpins identified by the given id in the path.
// @Tags gameInvitationWheelSpins
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameInvitationWheelSpinsByIDReply{}
// @Router /api/v1/gameInvitationWheelSpins/{id} [delete]
// @Security BearerAuth
func (h *gameInvitationWheelSpinsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationWheelSpinsIDFromPath(c)
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

// UpdateByID update a gameInvitationWheelSpins by id
// @Summary Update a gameInvitationWheelSpins by id
// @Description Updates the specified gameInvitationWheelSpins by given id in the path, support partial update.
// @Tags gameInvitationWheelSpins
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameInvitationWheelSpinsByIDRequest true "gameInvitationWheelSpins information"
// @Success 200 {object} types.UpdateGameInvitationWheelSpinsByIDReply{}
// @Router /api/v1/gameInvitationWheelSpins/{id} [put]
// @Security BearerAuth
func (h *gameInvitationWheelSpinsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationWheelSpinsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameInvitationWheelSpinsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameInvitationWheelSpins := &model.GameInvitationWheelSpins{}
	err = copier.Copy(gameInvitationWheelSpins, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameInvitationWheelSpins)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameInvitationWheelSpins)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameInvitationWheelSpins by id
// @Summary Get a gameInvitationWheelSpins by id
// @Description Gets detailed information of a gameInvitationWheelSpins specified by the given id in the path.
// @Tags gameInvitationWheelSpins
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameInvitationWheelSpinsByIDReply{}
// @Router /api/v1/gameInvitationWheelSpins/{id} [get]
// @Security BearerAuth
func (h *gameInvitationWheelSpinsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationWheelSpinsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameInvitationWheelSpins, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameInvitationWheelSpinsObjDetail{}
	err = copier.Copy(data, gameInvitationWheelSpins)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameInvitationWheelSpins)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameInvitationWheelSpins": data})
}

// List get a paginated list of gameInvitationWheelSpinss by custom conditions
// @Summary Get a paginated list of gameInvitationWheelSpinss by custom conditions
// @Description Returns a paginated list of gameInvitationWheelSpins based on query filters, including page number and size.
// @Tags gameInvitationWheelSpins
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameInvitationWheelSpinssReply{}
// @Router /api/v1/gameInvitationWheelSpins/list [post]
// @Security BearerAuth
func (h *gameInvitationWheelSpinsHandler) List(c *gin.Context) {
	form := &types.ListGameInvitationWheelSpinssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameInvitationWheelSpinss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameInvitationWheelSpinss(gameInvitationWheelSpinss)
	if err != nil {
		response.Error(c, ecode.ErrListGameInvitationWheelSpins)
		return
	}

	response.Success(c, gin.H{
		"gameInvitationWheelSpinss": data,
		"total":        total,
	})
}

func getGameInvitationWheelSpinsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameInvitationWheelSpins(gameInvitationWheelSpins *model.GameInvitationWheelSpins) (*types.GameInvitationWheelSpinsObjDetail, error) {
	data := &types.GameInvitationWheelSpinsObjDetail{}
	err := copier.Copy(data, gameInvitationWheelSpins)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameInvitationWheelSpinss(fromValues []*model.GameInvitationWheelSpins) ([]*types.GameInvitationWheelSpinsObjDetail, error) {
	toValues := []*types.GameInvitationWheelSpinsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameInvitationWheelSpins(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
