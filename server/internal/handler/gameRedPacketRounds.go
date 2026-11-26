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

var _ GameRedPacketRoundsHandler = (*gameRedPacketRoundsHandler)(nil)

// GameRedPacketRoundsHandler defining the handler interface
type GameRedPacketRoundsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameRedPacketRoundsHandler struct {
	iDao dao.GameRedPacketRoundsDao
}

// NewGameRedPacketRoundsHandler creating the handler interface
func NewGameRedPacketRoundsHandler() GameRedPacketRoundsHandler {
	return &gameRedPacketRoundsHandler{
		iDao: dao.NewGameRedPacketRoundsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameRedPacketRoundsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameRedPacketRounds
// @Summary Create a new gameRedPacketRounds
// @Description Creates a new gameRedPacketRounds entity using the provided data in the request body.
// @Tags gameRedPacketRounds
// @Accept json
// @Produce json
// @Param data body types.CreateGameRedPacketRoundsRequest true "gameRedPacketRounds information"
// @Success 200 {object} types.CreateGameRedPacketRoundsReply{}
// @Router /api/v1/gameRedPacketRounds [post]
// @Security BearerAuth
func (h *gameRedPacketRoundsHandler) Create(c *gin.Context) {
	form := &types.CreateGameRedPacketRoundsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameRedPacketRounds := &model.GameRedPacketRounds{}
	err = copier.Copy(gameRedPacketRounds, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameRedPacketRounds)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameRedPacketRounds)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameRedPacketRounds.ID})
}

// DeleteByID delete a gameRedPacketRounds by id
// @Summary Delete a gameRedPacketRounds by id
// @Description Deletes a existing gameRedPacketRounds identified by the given id in the path.
// @Tags gameRedPacketRounds
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameRedPacketRoundsByIDReply{}
// @Router /api/v1/gameRedPacketRounds/{id} [delete]
// @Security BearerAuth
func (h *gameRedPacketRoundsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameRedPacketRoundsIDFromPath(c)
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

// UpdateByID update a gameRedPacketRounds by id
// @Summary Update a gameRedPacketRounds by id
// @Description Updates the specified gameRedPacketRounds by given id in the path, support partial update.
// @Tags gameRedPacketRounds
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameRedPacketRoundsByIDRequest true "gameRedPacketRounds information"
// @Success 200 {object} types.UpdateGameRedPacketRoundsByIDReply{}
// @Router /api/v1/gameRedPacketRounds/{id} [put]
// @Security BearerAuth
func (h *gameRedPacketRoundsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameRedPacketRoundsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameRedPacketRoundsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameRedPacketRounds := &model.GameRedPacketRounds{}
	err = copier.Copy(gameRedPacketRounds, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameRedPacketRounds)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameRedPacketRounds)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameRedPacketRounds by id
// @Summary Get a gameRedPacketRounds by id
// @Description Gets detailed information of a gameRedPacketRounds specified by the given id in the path.
// @Tags gameRedPacketRounds
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameRedPacketRoundsByIDReply{}
// @Router /api/v1/gameRedPacketRounds/{id} [get]
// @Security BearerAuth
func (h *gameRedPacketRoundsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameRedPacketRoundsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameRedPacketRounds, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameRedPacketRoundsObjDetail{}
	err = copier.Copy(data, gameRedPacketRounds)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameRedPacketRounds)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameRedPacketRounds": data})
}

// List get a paginated list of gameRedPacketRoundss by custom conditions
// @Summary Get a paginated list of gameRedPacketRoundss by custom conditions
// @Description Returns a paginated list of gameRedPacketRounds based on query filters, including page number and size.
// @Tags gameRedPacketRounds
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameRedPacketRoundssReply{}
// @Router /api/v1/gameRedPacketRounds/list [post]
// @Security BearerAuth
func (h *gameRedPacketRoundsHandler) List(c *gin.Context) {
	form := &types.ListGameRedPacketRoundssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameRedPacketRoundss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameRedPacketRoundss(gameRedPacketRoundss)
	if err != nil {
		response.Error(c, ecode.ErrListGameRedPacketRounds)
		return
	}

	response.Success(c, gin.H{
		"gameRedPacketRoundss": data,
		"total":                total,
	})
}

func getGameRedPacketRoundsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameRedPacketRounds(gameRedPacketRounds *model.GameRedPacketRounds) (*types.GameRedPacketRoundsObjDetail, error) {
	data := &types.GameRedPacketRoundsObjDetail{}
	err := copier.Copy(data, gameRedPacketRounds)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameRedPacketRoundss(fromValues []*model.GameRedPacketRounds) ([]*types.GameRedPacketRoundsObjDetail, error) {
	toValues := []*types.GameRedPacketRoundsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameRedPacketRounds(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
