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

var _ GameUserWalletsHandler = (*gameUserWalletsHandler)(nil)

// GameUserWalletsHandler defining the handler interface
type GameUserWalletsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Summary(c *gin.Context)
	CashAudits(c *gin.Context)
}

type gameUserWalletsHandler struct {
	iDao dao.GameUserWalletsDao
}

// NewGameUserWalletsHandler creating the handler interface
func NewGameUserWalletsHandler() GameUserWalletsHandler {
	return &gameUserWalletsHandler{
		iDao: dao.NewGameUserWalletsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameUserWalletsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameUserWallets
// @Summary Create a new gameUserWallets
// @Description Creates a new gameUserWallets entity using the provided data in the request body.
// @Tags gameUserWallets
// @Accept json
// @Produce json
// @Param data body types.CreateGameUserWalletsRequest true "gameUserWallets information"
// @Success 200 {object} types.CreateGameUserWalletsReply{}
// @Router /api/v1/gameUserWallets [post]
// @Security BearerAuth
func (h *gameUserWalletsHandler) Create(c *gin.Context) {
	form := &types.CreateGameUserWalletsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameUserWallets := &model.GameUserWallets{}
	err = copier.Copy(gameUserWallets, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameUserWallets)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameUserWallets)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameUserWallets.ID})
}

// DeleteByID delete a gameUserWallets by id
// @Summary Delete a gameUserWallets by id
// @Description Deletes a existing gameUserWallets identified by the given id in the path.
// @Tags gameUserWallets
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameUserWalletsByIDReply{}
// @Router /api/v1/gameUserWallets/{id} [delete]
// @Security BearerAuth
func (h *gameUserWalletsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameUserWalletsIDFromPath(c)
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

// UpdateByID update a gameUserWallets by id
// @Summary Update a gameUserWallets by id
// @Description Updates the specified gameUserWallets by given id in the path, support partial update.
// @Tags gameUserWallets
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameUserWalletsByIDRequest true "gameUserWallets information"
// @Success 200 {object} types.UpdateGameUserWalletsByIDReply{}
// @Router /api/v1/gameUserWallets/{id} [put]
// @Security BearerAuth
func (h *gameUserWalletsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameUserWalletsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameUserWalletsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameUserWallets := &model.GameUserWallets{}
	err = copier.Copy(gameUserWallets, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameUserWallets)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameUserWallets)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameUserWallets by id
// @Summary Get a gameUserWallets by id
// @Description Gets detailed information of a gameUserWallets specified by the given id in the path.
// @Tags gameUserWallets
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameUserWalletsByIDReply{}
// @Router /api/v1/gameUserWallets/{id} [get]
// @Security BearerAuth
func (h *gameUserWalletsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameUserWalletsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserWallets, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameUserWalletsObjDetail{}
	err = copier.Copy(data, gameUserWallets)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameUserWallets)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameUserWallets": data})
}

// List get a paginated list of gameUserWalletss by custom conditions
// @Summary Get a paginated list of gameUserWalletss by custom conditions
// @Description Returns a paginated list of gameUserWallets based on query filters, including page number and size.
// @Tags gameUserWallets
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameUserWalletssReply{}
// @Router /api/v1/gameUserWallets/list [post]
// @Security BearerAuth
func (h *gameUserWalletsHandler) List(c *gin.Context) {
	form := &types.ListGameUserWalletssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserWalletss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameUserWalletss(gameUserWalletss)
	if err != nil {
		response.Error(c, ecode.ErrListGameUserWallets)
		return
	}

	response.Success(c, gin.H{
		"gameUserWalletss": data,
		"total":            total,
	})
}

func getGameUserWalletsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameUserWallets(gameUserWallets *model.GameUserWallets) (*types.GameUserWalletsObjDetail, error) {
	data := &types.GameUserWalletsObjDetail{}
	err := copier.Copy(data, gameUserWallets)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameUserWalletss(fromValues []*model.GameUserWallets) ([]*types.GameUserWalletsObjDetail, error) {
	toValues := []*types.GameUserWalletsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameUserWallets(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
