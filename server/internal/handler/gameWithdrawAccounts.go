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

var _ GameWithdrawAccountsHandler = (*gameWithdrawAccountsHandler)(nil)

// GameWithdrawAccountsHandler defining the handler interface
type GameWithdrawAccountsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Current(c *gin.Context)
	CreateCurrent(c *gin.Context)
}

type gameWithdrawAccountsHandler struct {
	iDao dao.GameWithdrawAccountsDao
}

// NewGameWithdrawAccountsHandler creating the handler interface
func NewGameWithdrawAccountsHandler() GameWithdrawAccountsHandler {
	return &gameWithdrawAccountsHandler{
		iDao: dao.NewGameWithdrawAccountsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameWithdrawAccountsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameWithdrawAccounts
// @Summary Create a new gameWithdrawAccounts
// @Description Creates a new gameWithdrawAccounts entity using the provided data in the request body.
// @Tags gameWithdrawAccounts
// @Accept json
// @Produce json
// @Param data body types.CreateGameWithdrawAccountsRequest true "gameWithdrawAccounts information"
// @Success 200 {object} types.CreateGameWithdrawAccountsReply{}
// @Router /api/v1/gameWithdrawAccounts [post]
// @Security BearerAuth
func (h *gameWithdrawAccountsHandler) Create(c *gin.Context) {
	form := &types.CreateGameWithdrawAccountsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameWithdrawAccounts := &model.GameWithdrawAccounts{}
	err = copier.Copy(gameWithdrawAccounts, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameWithdrawAccounts)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameWithdrawAccounts)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameWithdrawAccounts.ID})
}

// DeleteByID delete a gameWithdrawAccounts by id
// @Summary Delete a gameWithdrawAccounts by id
// @Description Deletes a existing gameWithdrawAccounts identified by the given id in the path.
// @Tags gameWithdrawAccounts
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameWithdrawAccountsByIDReply{}
// @Router /api/v1/gameWithdrawAccounts/{id} [delete]
// @Security BearerAuth
func (h *gameWithdrawAccountsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameWithdrawAccountsIDFromPath(c)
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

// UpdateByID update a gameWithdrawAccounts by id
// @Summary Update a gameWithdrawAccounts by id
// @Description Updates the specified gameWithdrawAccounts by given id in the path, support partial update.
// @Tags gameWithdrawAccounts
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameWithdrawAccountsByIDRequest true "gameWithdrawAccounts information"
// @Success 200 {object} types.UpdateGameWithdrawAccountsByIDReply{}
// @Router /api/v1/gameWithdrawAccounts/{id} [put]
// @Security BearerAuth
func (h *gameWithdrawAccountsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameWithdrawAccountsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameWithdrawAccountsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameWithdrawAccounts := &model.GameWithdrawAccounts{}
	err = copier.Copy(gameWithdrawAccounts, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameWithdrawAccounts)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameWithdrawAccounts)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameWithdrawAccounts by id
// @Summary Get a gameWithdrawAccounts by id
// @Description Gets detailed information of a gameWithdrawAccounts specified by the given id in the path.
// @Tags gameWithdrawAccounts
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameWithdrawAccountsByIDReply{}
// @Router /api/v1/gameWithdrawAccounts/{id} [get]
// @Security BearerAuth
func (h *gameWithdrawAccountsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameWithdrawAccountsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameWithdrawAccounts, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameWithdrawAccountsObjDetail{}
	err = copier.Copy(data, gameWithdrawAccounts)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameWithdrawAccounts)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameWithdrawAccounts": data})
}

// List get a paginated list of gameWithdrawAccountss by custom conditions
// @Summary Get a paginated list of gameWithdrawAccountss by custom conditions
// @Description Returns a paginated list of gameWithdrawAccounts based on query filters, including page number and size.
// @Tags gameWithdrawAccounts
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameWithdrawAccountssReply{}
// @Router /api/v1/gameWithdrawAccounts/list [post]
// @Security BearerAuth
func (h *gameWithdrawAccountsHandler) List(c *gin.Context) {
	form := &types.ListGameWithdrawAccountssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameWithdrawAccountss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameWithdrawAccountss(gameWithdrawAccountss)
	if err != nil {
		response.Error(c, ecode.ErrListGameWithdrawAccounts)
		return
	}

	response.Success(c, gin.H{
		"gameWithdrawAccountss": data,
		"total":                 total,
	})
}

func getGameWithdrawAccountsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameWithdrawAccounts(gameWithdrawAccounts *model.GameWithdrawAccounts) (*types.GameWithdrawAccountsObjDetail, error) {
	data := &types.GameWithdrawAccountsObjDetail{}
	err := copier.Copy(data, gameWithdrawAccounts)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameWithdrawAccountss(fromValues []*model.GameWithdrawAccounts) ([]*types.GameWithdrawAccountsObjDetail, error) {
	toValues := []*types.GameWithdrawAccountsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameWithdrawAccounts(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
