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

var _ GameNoticesHandler = (*gameNoticesHandler)(nil)

// GameNoticesHandler defining the handler interface
type GameNoticesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameNoticesHandler struct {
	iDao dao.GameNoticesDao
}

// NewGameNoticesHandler creating the handler interface
func NewGameNoticesHandler() GameNoticesHandler {
	return &gameNoticesHandler{
		iDao: dao.NewGameNoticesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameNoticesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameNotices
// @Summary Create a new gameNotices
// @Description Creates a new gameNotices entity using the provided data in the request body.
// @Tags gameNotices
// @Accept json
// @Produce json
// @Param data body types.CreateGameNoticesRequest true "gameNotices information"
// @Success 200 {object} types.CreateGameNoticesReply{}
// @Router /api/v1/gameNotices [post]
// @Security BearerAuth
func (h *gameNoticesHandler) Create(c *gin.Context) {
	form := &types.CreateGameNoticesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameNotices := &model.GameNotices{}
	err = copier.Copy(gameNotices, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameNotices)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameNotices)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameNotices.ID})
}

// DeleteByID delete a gameNotices by id
// @Summary Delete a gameNotices by id
// @Description Deletes a existing gameNotices identified by the given id in the path.
// @Tags gameNotices
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameNoticesByIDReply{}
// @Router /api/v1/gameNotices/{id} [delete]
// @Security BearerAuth
func (h *gameNoticesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameNoticesIDFromPath(c)
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

// UpdateByID update a gameNotices by id
// @Summary Update a gameNotices by id
// @Description Updates the specified gameNotices by given id in the path, support partial update.
// @Tags gameNotices
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameNoticesByIDRequest true "gameNotices information"
// @Success 200 {object} types.UpdateGameNoticesByIDReply{}
// @Router /api/v1/gameNotices/{id} [put]
// @Security BearerAuth
func (h *gameNoticesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameNoticesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameNoticesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameNotices := &model.GameNotices{}
	err = copier.Copy(gameNotices, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameNotices)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameNotices)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameNotices by id
// @Summary Get a gameNotices by id
// @Description Gets detailed information of a gameNotices specified by the given id in the path.
// @Tags gameNotices
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameNoticesByIDReply{}
// @Router /api/v1/gameNotices/{id} [get]
// @Security BearerAuth
func (h *gameNoticesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameNoticesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameNotices, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameNoticesObjDetail{}
	err = copier.Copy(data, gameNotices)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameNotices)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameNotices": data})
}

// List get a paginated list of gameNoticess by custom conditions
// @Summary Get a paginated list of gameNoticess by custom conditions
// @Description Returns a paginated list of gameNotices based on query filters, including page number and size.
// @Tags gameNotices
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameNoticessReply{}
// @Router /api/v1/gameNotices/list [post]
// @Security BearerAuth
func (h *gameNoticesHandler) List(c *gin.Context) {
	form := &types.ListGameNoticessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameNoticess, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameNoticess(gameNoticess)
	if err != nil {
		response.Error(c, ecode.ErrListGameNotices)
		return
	}

	response.Success(c, gin.H{
		"gameNoticess": data,
		"total":        total,
	})
}

func getGameNoticesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameNotices(gameNotices *model.GameNotices) (*types.GameNoticesObjDetail, error) {
	data := &types.GameNoticesObjDetail{}
	err := copier.Copy(data, gameNotices)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameNoticess(fromValues []*model.GameNotices) ([]*types.GameNoticesObjDetail, error) {
	toValues := []*types.GameNoticesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameNotices(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
