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

var _ GameSmsVerificationCodesHandler = (*gameSmsVerificationCodesHandler)(nil)

// GameSmsVerificationCodesHandler defining the handler interface
type GameSmsVerificationCodesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameSmsVerificationCodesHandler struct {
	iDao dao.GameSmsVerificationCodesDao
}

// NewGameSmsVerificationCodesHandler creating the handler interface
func NewGameSmsVerificationCodesHandler() GameSmsVerificationCodesHandler {
	return &gameSmsVerificationCodesHandler{
		iDao: dao.NewGameSmsVerificationCodesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameSmsVerificationCodesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameSmsVerificationCodes
// @Summary Create a new gameSmsVerificationCodes
// @Description Creates a new gameSmsVerificationCodes entity using the provided data in the request body.
// @Tags gameSmsVerificationCodes
// @Accept json
// @Produce json
// @Param data body types.CreateGameSmsVerificationCodesRequest true "gameSmsVerificationCodes information"
// @Success 200 {object} types.CreateGameSmsVerificationCodesReply{}
// @Router /api/v1/gameSmsVerificationCodes [post]
// @Security BearerAuth
func (h *gameSmsVerificationCodesHandler) Create(c *gin.Context) {
	form := &types.CreateGameSmsVerificationCodesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameSmsVerificationCodes := &model.GameSmsVerificationCodes{}
	err = copier.Copy(gameSmsVerificationCodes, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameSmsVerificationCodes)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameSmsVerificationCodes)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameSmsVerificationCodes.ID})
}

// DeleteByID delete a gameSmsVerificationCodes by id
// @Summary Delete a gameSmsVerificationCodes by id
// @Description Deletes a existing gameSmsVerificationCodes identified by the given id in the path.
// @Tags gameSmsVerificationCodes
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameSmsVerificationCodesByIDReply{}
// @Router /api/v1/gameSmsVerificationCodes/{id} [delete]
// @Security BearerAuth
func (h *gameSmsVerificationCodesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameSmsVerificationCodesIDFromPath(c)
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

// UpdateByID update a gameSmsVerificationCodes by id
// @Summary Update a gameSmsVerificationCodes by id
// @Description Updates the specified gameSmsVerificationCodes by given id in the path, support partial update.
// @Tags gameSmsVerificationCodes
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameSmsVerificationCodesByIDRequest true "gameSmsVerificationCodes information"
// @Success 200 {object} types.UpdateGameSmsVerificationCodesByIDReply{}
// @Router /api/v1/gameSmsVerificationCodes/{id} [put]
// @Security BearerAuth
func (h *gameSmsVerificationCodesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameSmsVerificationCodesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameSmsVerificationCodesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameSmsVerificationCodes := &model.GameSmsVerificationCodes{}
	err = copier.Copy(gameSmsVerificationCodes, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameSmsVerificationCodes)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameSmsVerificationCodes)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameSmsVerificationCodes by id
// @Summary Get a gameSmsVerificationCodes by id
// @Description Gets detailed information of a gameSmsVerificationCodes specified by the given id in the path.
// @Tags gameSmsVerificationCodes
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameSmsVerificationCodesByIDReply{}
// @Router /api/v1/gameSmsVerificationCodes/{id} [get]
// @Security BearerAuth
func (h *gameSmsVerificationCodesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameSmsVerificationCodesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameSmsVerificationCodes, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameSmsVerificationCodesObjDetail{}
	err = copier.Copy(data, gameSmsVerificationCodes)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameSmsVerificationCodes)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameSmsVerificationCodes": data})
}

// List get a paginated list of gameSmsVerificationCodess by custom conditions
// @Summary Get a paginated list of gameSmsVerificationCodess by custom conditions
// @Description Returns a paginated list of gameSmsVerificationCodes based on query filters, including page number and size.
// @Tags gameSmsVerificationCodes
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameSmsVerificationCodessReply{}
// @Router /api/v1/gameSmsVerificationCodes/list [post]
// @Security BearerAuth
func (h *gameSmsVerificationCodesHandler) List(c *gin.Context) {
	form := &types.ListGameSmsVerificationCodessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameSmsVerificationCodess, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameSmsVerificationCodess(gameSmsVerificationCodess)
	if err != nil {
		response.Error(c, ecode.ErrListGameSmsVerificationCodes)
		return
	}

	response.Success(c, gin.H{
		"gameSmsVerificationCodess": data,
		"total":        total,
	})
}

func getGameSmsVerificationCodesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameSmsVerificationCodes(gameSmsVerificationCodes *model.GameSmsVerificationCodes) (*types.GameSmsVerificationCodesObjDetail, error) {
	data := &types.GameSmsVerificationCodesObjDetail{}
	err := copier.Copy(data, gameSmsVerificationCodes)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameSmsVerificationCodess(fromValues []*model.GameSmsVerificationCodes) ([]*types.GameSmsVerificationCodesObjDetail, error) {
	toValues := []*types.GameSmsVerificationCodesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameSmsVerificationCodes(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
