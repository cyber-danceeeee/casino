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

var _ GameSupportMessagesHandler = (*gameSupportMessagesHandler)(nil)

// GameSupportMessagesHandler defining the handler interface
type GameSupportMessagesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameSupportMessagesHandler struct {
	iDao dao.GameSupportMessagesDao
}

// NewGameSupportMessagesHandler creating the handler interface
func NewGameSupportMessagesHandler() GameSupportMessagesHandler {
	return &gameSupportMessagesHandler{
		iDao: dao.NewGameSupportMessagesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameSupportMessagesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameSupportMessages
// @Summary Create a new gameSupportMessages
// @Description Creates a new gameSupportMessages entity using the provided data in the request body.
// @Tags gameSupportMessages
// @Accept json
// @Produce json
// @Param data body types.CreateGameSupportMessagesRequest true "gameSupportMessages information"
// @Success 200 {object} types.CreateGameSupportMessagesReply{}
// @Router /api/v1/gameSupportMessages [post]
// @Security BearerAuth
func (h *gameSupportMessagesHandler) Create(c *gin.Context) {
	form := &types.CreateGameSupportMessagesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameSupportMessages := &model.GameSupportMessages{}
	err = copier.Copy(gameSupportMessages, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameSupportMessages)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameSupportMessages)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameSupportMessages.ID})
}

// DeleteByID delete a gameSupportMessages by id
// @Summary Delete a gameSupportMessages by id
// @Description Deletes a existing gameSupportMessages identified by the given id in the path.
// @Tags gameSupportMessages
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameSupportMessagesByIDReply{}
// @Router /api/v1/gameSupportMessages/{id} [delete]
// @Security BearerAuth
func (h *gameSupportMessagesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameSupportMessagesIDFromPath(c)
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

// UpdateByID update a gameSupportMessages by id
// @Summary Update a gameSupportMessages by id
// @Description Updates the specified gameSupportMessages by given id in the path, support partial update.
// @Tags gameSupportMessages
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameSupportMessagesByIDRequest true "gameSupportMessages information"
// @Success 200 {object} types.UpdateGameSupportMessagesByIDReply{}
// @Router /api/v1/gameSupportMessages/{id} [put]
// @Security BearerAuth
func (h *gameSupportMessagesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameSupportMessagesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameSupportMessagesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameSupportMessages := &model.GameSupportMessages{}
	err = copier.Copy(gameSupportMessages, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameSupportMessages)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameSupportMessages)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameSupportMessages by id
// @Summary Get a gameSupportMessages by id
// @Description Gets detailed information of a gameSupportMessages specified by the given id in the path.
// @Tags gameSupportMessages
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameSupportMessagesByIDReply{}
// @Router /api/v1/gameSupportMessages/{id} [get]
// @Security BearerAuth
func (h *gameSupportMessagesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameSupportMessagesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameSupportMessages, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameSupportMessagesObjDetail{}
	err = copier.Copy(data, gameSupportMessages)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameSupportMessages)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameSupportMessages": data})
}

// List get a paginated list of gameSupportMessagess by custom conditions
// @Summary Get a paginated list of gameSupportMessagess by custom conditions
// @Description Returns a paginated list of gameSupportMessages based on query filters, including page number and size.
// @Tags gameSupportMessages
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameSupportMessagessReply{}
// @Router /api/v1/gameSupportMessages/list [post]
// @Security BearerAuth
func (h *gameSupportMessagesHandler) List(c *gin.Context) {
	form := &types.ListGameSupportMessagessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameSupportMessagess, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameSupportMessagess(gameSupportMessagess)
	if err != nil {
		response.Error(c, ecode.ErrListGameSupportMessages)
		return
	}

	response.Success(c, gin.H{
		"gameSupportMessagess": data,
		"total":                total,
	})
}

func getGameSupportMessagesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameSupportMessages(gameSupportMessages *model.GameSupportMessages) (*types.GameSupportMessagesObjDetail, error) {
	data := &types.GameSupportMessagesObjDetail{}
	err := copier.Copy(data, gameSupportMessages)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameSupportMessagess(fromValues []*model.GameSupportMessages) ([]*types.GameSupportMessagesObjDetail, error) {
	toValues := []*types.GameSupportMessagesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameSupportMessages(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
