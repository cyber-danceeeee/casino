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

var _ GameInvitesHandler = (*gameInvitesHandler)(nil)

// GameInvitesHandler defining the handler interface
type GameInvitesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Share(c *gin.Context)
}

type gameInvitesHandler struct {
	iDao dao.GameInvitesDao
}

// NewGameInvitesHandler creating the handler interface
func NewGameInvitesHandler() GameInvitesHandler {
	return &gameInvitesHandler{
		iDao: dao.NewGameInvitesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameInvitesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameInvites
// @Summary Create a new gameInvites
// @Description Creates a new gameInvites entity using the provided data in the request body.
// @Tags gameInvites
// @Accept json
// @Produce json
// @Param data body types.CreateGameInvitesRequest true "gameInvites information"
// @Success 200 {object} types.CreateGameInvitesReply{}
// @Router /api/v1/gameInvites [post]
// @Security BearerAuth
func (h *gameInvitesHandler) Create(c *gin.Context) {
	form := &types.CreateGameInvitesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameInvites := &model.GameInvites{}
	err = copier.Copy(gameInvites, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameInvites)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameInvites)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameInvites.ID})
}

// DeleteByID delete a gameInvites by id
// @Summary Delete a gameInvites by id
// @Description Deletes a existing gameInvites identified by the given id in the path.
// @Tags gameInvites
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameInvitesByIDReply{}
// @Router /api/v1/gameInvites/{id} [delete]
// @Security BearerAuth
func (h *gameInvitesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameInvitesIDFromPath(c)
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

// UpdateByID update a gameInvites by id
// @Summary Update a gameInvites by id
// @Description Updates the specified gameInvites by given id in the path, support partial update.
// @Tags gameInvites
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameInvitesByIDRequest true "gameInvites information"
// @Success 200 {object} types.UpdateGameInvitesByIDReply{}
// @Router /api/v1/gameInvites/{id} [put]
// @Security BearerAuth
func (h *gameInvitesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameInvitesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameInvitesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameInvites := &model.GameInvites{}
	err = copier.Copy(gameInvites, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameInvites)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameInvites)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameInvites by id
// @Summary Get a gameInvites by id
// @Description Gets detailed information of a gameInvites specified by the given id in the path.
// @Tags gameInvites
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameInvitesByIDReply{}
// @Router /api/v1/gameInvites/{id} [get]
// @Security BearerAuth
func (h *gameInvitesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameInvitesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameInvites, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameInvitesObjDetail{}
	err = copier.Copy(data, gameInvites)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameInvites)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameInvites": data})
}

// List get a paginated list of gameInvitess by custom conditions
// @Summary Get a paginated list of gameInvitess by custom conditions
// @Description Returns a paginated list of gameInvites based on query filters, including page number and size.
// @Tags gameInvites
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameInvitessReply{}
// @Router /api/v1/gameInvites/list [post]
// @Security BearerAuth
func (h *gameInvitesHandler) List(c *gin.Context) {
	form := &types.ListGameInvitessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameInvitess, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameInvitess(gameInvitess)
	if err != nil {
		response.Error(c, ecode.ErrListGameInvites)
		return
	}

	response.Success(c, gin.H{
		"gameInvitess": data,
		"total":        total,
	})
}

func getGameInvitesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameInvites(gameInvites *model.GameInvites) (*types.GameInvitesObjDetail, error) {
	data := &types.GameInvitesObjDetail{}
	err := copier.Copy(data, gameInvites)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameInvitess(fromValues []*model.GameInvites) ([]*types.GameInvitesObjDetail, error) {
	toValues := []*types.GameInvitesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameInvites(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
