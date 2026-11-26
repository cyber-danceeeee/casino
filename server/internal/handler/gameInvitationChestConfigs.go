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

var _ GameInvitationChestConfigsHandler = (*gameInvitationChestConfigsHandler)(nil)

// GameInvitationChestConfigsHandler defining the handler interface
type GameInvitationChestConfigsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameInvitationChestConfigsHandler struct {
	iDao dao.GameInvitationChestConfigsDao
}

// NewGameInvitationChestConfigsHandler creating the handler interface
func NewGameInvitationChestConfigsHandler() GameInvitationChestConfigsHandler {
	return &gameInvitationChestConfigsHandler{
		iDao: dao.NewGameInvitationChestConfigsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameInvitationChestConfigsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameInvitationChestConfigs
// @Summary Create a new gameInvitationChestConfigs
// @Description Creates a new gameInvitationChestConfigs entity using the provided data in the request body.
// @Tags gameInvitationChestConfigs
// @Accept json
// @Produce json
// @Param data body types.CreateGameInvitationChestConfigsRequest true "gameInvitationChestConfigs information"
// @Success 200 {object} types.CreateGameInvitationChestConfigsReply{}
// @Router /api/v1/gameInvitationChestConfigs [post]
// @Security BearerAuth
func (h *gameInvitationChestConfigsHandler) Create(c *gin.Context) {
	form := &types.CreateGameInvitationChestConfigsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameInvitationChestConfigs := &model.GameInvitationChestConfigs{}
	err = copier.Copy(gameInvitationChestConfigs, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameInvitationChestConfigs)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameInvitationChestConfigs)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameInvitationChestConfigs.ID})
}

// DeleteByID delete a gameInvitationChestConfigs by id
// @Summary Delete a gameInvitationChestConfigs by id
// @Description Deletes a existing gameInvitationChestConfigs identified by the given id in the path.
// @Tags gameInvitationChestConfigs
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameInvitationChestConfigsByIDReply{}
// @Router /api/v1/gameInvitationChestConfigs/{id} [delete]
// @Security BearerAuth
func (h *gameInvitationChestConfigsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationChestConfigsIDFromPath(c)
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

// UpdateByID update a gameInvitationChestConfigs by id
// @Summary Update a gameInvitationChestConfigs by id
// @Description Updates the specified gameInvitationChestConfigs by given id in the path, support partial update.
// @Tags gameInvitationChestConfigs
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameInvitationChestConfigsByIDRequest true "gameInvitationChestConfigs information"
// @Success 200 {object} types.UpdateGameInvitationChestConfigsByIDReply{}
// @Router /api/v1/gameInvitationChestConfigs/{id} [put]
// @Security BearerAuth
func (h *gameInvitationChestConfigsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationChestConfigsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameInvitationChestConfigsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameInvitationChestConfigs := &model.GameInvitationChestConfigs{}
	err = copier.Copy(gameInvitationChestConfigs, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameInvitationChestConfigs)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameInvitationChestConfigs)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameInvitationChestConfigs by id
// @Summary Get a gameInvitationChestConfigs by id
// @Description Gets detailed information of a gameInvitationChestConfigs specified by the given id in the path.
// @Tags gameInvitationChestConfigs
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameInvitationChestConfigsByIDReply{}
// @Router /api/v1/gameInvitationChestConfigs/{id} [get]
// @Security BearerAuth
func (h *gameInvitationChestConfigsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameInvitationChestConfigsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameInvitationChestConfigs, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameInvitationChestConfigsObjDetail{}
	err = copier.Copy(data, gameInvitationChestConfigs)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameInvitationChestConfigs)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameInvitationChestConfigs": data})
}

// List get a paginated list of gameInvitationChestConfigss by custom conditions
// @Summary Get a paginated list of gameInvitationChestConfigss by custom conditions
// @Description Returns a paginated list of gameInvitationChestConfigs based on query filters, including page number and size.
// @Tags gameInvitationChestConfigs
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameInvitationChestConfigssReply{}
// @Router /api/v1/gameInvitationChestConfigs/list [post]
// @Security BearerAuth
func (h *gameInvitationChestConfigsHandler) List(c *gin.Context) {
	form := &types.ListGameInvitationChestConfigssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameInvitationChestConfigss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameInvitationChestConfigss(gameInvitationChestConfigss)
	if err != nil {
		response.Error(c, ecode.ErrListGameInvitationChestConfigs)
		return
	}

	response.Success(c, gin.H{
		"gameInvitationChestConfigss": data,
		"total":        total,
	})
}

func getGameInvitationChestConfigsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameInvitationChestConfigs(gameInvitationChestConfigs *model.GameInvitationChestConfigs) (*types.GameInvitationChestConfigsObjDetail, error) {
	data := &types.GameInvitationChestConfigsObjDetail{}
	err := copier.Copy(data, gameInvitationChestConfigs)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameInvitationChestConfigss(fromValues []*model.GameInvitationChestConfigs) ([]*types.GameInvitationChestConfigsObjDetail, error) {
	toValues := []*types.GameInvitationChestConfigsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameInvitationChestConfigs(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
