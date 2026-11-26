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

var _ GameUsersHandler = (*gameUsersHandler)(nil)

// GameUsersHandler defining the handler interface
type GameUsersHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Login(c *gin.Context)
	Register(c *gin.Context)
	Logout(c *gin.Context)
	Me(c *gin.Context)
	Profile(c *gin.Context)
	PendingRewards(c *gin.Context)
	ClaimPendingReward(c *gin.Context)
	ClaimRegisterReward(c *gin.Context)
}

type gameUsersHandler struct {
	iDao dao.GameUsersDao
}

// NewGameUsersHandler creating the handler interface
func NewGameUsersHandler() GameUsersHandler {
	return &gameUsersHandler{
		iDao: dao.NewGameUsersDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameUsersCache(database.GetCacheType()),
		),
	}
}

// Create a new gameUsers
// @Summary Create a new gameUsers
// @Description Creates a new gameUsers entity using the provided data in the request body.
// @Tags gameUsers
// @Accept json
// @Produce json
// @Param data body types.CreateGameUsersRequest true "gameUsers information"
// @Success 200 {object} types.CreateGameUsersReply{}
// @Router /api/v1/gameUsers [post]
// @Security BearerAuth
func (h *gameUsersHandler) Create(c *gin.Context) {
	form := &types.CreateGameUsersRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameUsers := &model.GameUsers{}
	err = copier.Copy(gameUsers, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameUsers)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameUsers)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameUsers.ID})
}

// DeleteByID delete a gameUsers by id
// @Summary Delete a gameUsers by id
// @Description Deletes a existing gameUsers identified by the given id in the path.
// @Tags gameUsers
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameUsersByIDReply{}
// @Router /api/v1/gameUsers/{id} [delete]
// @Security BearerAuth
func (h *gameUsersHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameUsersIDFromPath(c)
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

// UpdateByID update a gameUsers by id
// @Summary Update a gameUsers by id
// @Description Updates the specified gameUsers by given id in the path, support partial update.
// @Tags gameUsers
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameUsersByIDRequest true "gameUsers information"
// @Success 200 {object} types.UpdateGameUsersByIDReply{}
// @Router /api/v1/gameUsers/{id} [put]
// @Security BearerAuth
func (h *gameUsersHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameUsersIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameUsersByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameUsers := &model.GameUsers{}
	err = copier.Copy(gameUsers, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameUsers)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameUsers)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameUsers by id
// @Summary Get a gameUsers by id
// @Description Gets detailed information of a gameUsers specified by the given id in the path.
// @Tags gameUsers
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameUsersByIDReply{}
// @Router /api/v1/gameUsers/{id} [get]
// @Security BearerAuth
func (h *gameUsersHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameUsersIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUsers, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameUsersObjDetail{}
	err = copier.Copy(data, gameUsers)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameUsers)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameUsers": data})
}

// List get a paginated list of gameUserss by custom conditions
// @Summary Get a paginated list of gameUserss by custom conditions
// @Description Returns a paginated list of gameUsers based on query filters, including page number and size.
// @Tags gameUsers
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameUserssReply{}
// @Router /api/v1/gameUsers/list [post]
// @Security BearerAuth
func (h *gameUsersHandler) List(c *gin.Context) {
	form := &types.ListGameUserssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameUserss(gameUserss)
	if err != nil {
		response.Error(c, ecode.ErrListGameUsers)
		return
	}

	response.Success(c, gin.H{
		"gameUserss": data,
		"total":      total,
	})
}

func getGameUsersIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameUsers(gameUsers *model.GameUsers) (*types.GameUsersObjDetail, error) {
	data := &types.GameUsersObjDetail{}
	err := copier.Copy(data, gameUsers)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameUserss(fromValues []*model.GameUsers) ([]*types.GameUsersObjDetail, error) {
	toValues := []*types.GameUsersObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameUsers(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
