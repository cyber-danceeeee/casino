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

var _ GameUserInvitationChestsHandler = (*gameUserInvitationChestsHandler)(nil)

// GameUserInvitationChestsHandler defining the handler interface
type GameUserInvitationChestsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameUserInvitationChestsHandler struct {
	iDao dao.GameUserInvitationChestsDao
}

// NewGameUserInvitationChestsHandler creating the handler interface
func NewGameUserInvitationChestsHandler() GameUserInvitationChestsHandler {
	return &gameUserInvitationChestsHandler{
		iDao: dao.NewGameUserInvitationChestsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameUserInvitationChestsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameUserInvitationChests
// @Summary Create a new gameUserInvitationChests
// @Description Creates a new gameUserInvitationChests entity using the provided data in the request body.
// @Tags gameUserInvitationChests
// @Accept json
// @Produce json
// @Param data body types.CreateGameUserInvitationChestsRequest true "gameUserInvitationChests information"
// @Success 200 {object} types.CreateGameUserInvitationChestsReply{}
// @Router /api/v1/gameUserInvitationChests [post]
// @Security BearerAuth
func (h *gameUserInvitationChestsHandler) Create(c *gin.Context) {
	form := &types.CreateGameUserInvitationChestsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameUserInvitationChests := &model.GameUserInvitationChests{}
	err = copier.Copy(gameUserInvitationChests, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameUserInvitationChests)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameUserInvitationChests)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameUserInvitationChests.ID})
}

// DeleteByID delete a gameUserInvitationChests by id
// @Summary Delete a gameUserInvitationChests by id
// @Description Deletes a existing gameUserInvitationChests identified by the given id in the path.
// @Tags gameUserInvitationChests
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameUserInvitationChestsByIDReply{}
// @Router /api/v1/gameUserInvitationChests/{id} [delete]
// @Security BearerAuth
func (h *gameUserInvitationChestsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameUserInvitationChestsIDFromPath(c)
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

// UpdateByID update a gameUserInvitationChests by id
// @Summary Update a gameUserInvitationChests by id
// @Description Updates the specified gameUserInvitationChests by given id in the path, support partial update.
// @Tags gameUserInvitationChests
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameUserInvitationChestsByIDRequest true "gameUserInvitationChests information"
// @Success 200 {object} types.UpdateGameUserInvitationChestsByIDReply{}
// @Router /api/v1/gameUserInvitationChests/{id} [put]
// @Security BearerAuth
func (h *gameUserInvitationChestsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameUserInvitationChestsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameUserInvitationChestsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameUserInvitationChests := &model.GameUserInvitationChests{}
	err = copier.Copy(gameUserInvitationChests, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameUserInvitationChests)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameUserInvitationChests)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameUserInvitationChests by id
// @Summary Get a gameUserInvitationChests by id
// @Description Gets detailed information of a gameUserInvitationChests specified by the given id in the path.
// @Tags gameUserInvitationChests
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameUserInvitationChestsByIDReply{}
// @Router /api/v1/gameUserInvitationChests/{id} [get]
// @Security BearerAuth
func (h *gameUserInvitationChestsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameUserInvitationChestsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserInvitationChests, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameUserInvitationChestsObjDetail{}
	err = copier.Copy(data, gameUserInvitationChests)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameUserInvitationChests)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameUserInvitationChests": data})
}

// List get a paginated list of gameUserInvitationChestss by custom conditions
// @Summary Get a paginated list of gameUserInvitationChestss by custom conditions
// @Description Returns a paginated list of gameUserInvitationChests based on query filters, including page number and size.
// @Tags gameUserInvitationChests
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameUserInvitationChestssReply{}
// @Router /api/v1/gameUserInvitationChests/list [post]
// @Security BearerAuth
func (h *gameUserInvitationChestsHandler) List(c *gin.Context) {
	form := &types.ListGameUserInvitationChestssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserInvitationChestss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameUserInvitationChestss(gameUserInvitationChestss)
	if err != nil {
		response.Error(c, ecode.ErrListGameUserInvitationChests)
		return
	}

	response.Success(c, gin.H{
		"gameUserInvitationChestss": data,
		"total":                     total,
	})
}

func getGameUserInvitationChestsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameUserInvitationChests(gameUserInvitationChests *model.GameUserInvitationChests) (*types.GameUserInvitationChestsObjDetail, error) {
	data := &types.GameUserInvitationChestsObjDetail{}
	err := copier.Copy(data, gameUserInvitationChests)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameUserInvitationChestss(fromValues []*model.GameUserInvitationChests) ([]*types.GameUserInvitationChestsObjDetail, error) {
	toValues := []*types.GameUserInvitationChestsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameUserInvitationChests(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
