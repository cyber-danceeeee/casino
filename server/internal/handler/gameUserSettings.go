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

var _ GameUserSettingsHandler = (*gameUserSettingsHandler)(nil)

// GameUserSettingsHandler defining the handler interface
type GameUserSettingsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Current(c *gin.Context)
	UpdateCurrent(c *gin.Context)
}

type gameUserSettingsHandler struct {
	iDao dao.GameUserSettingsDao
}

// NewGameUserSettingsHandler creating the handler interface
func NewGameUserSettingsHandler() GameUserSettingsHandler {
	return &gameUserSettingsHandler{
		iDao: dao.NewGameUserSettingsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameUserSettingsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameUserSettings
// @Summary Create a new gameUserSettings
// @Description Creates a new gameUserSettings entity using the provided data in the request body.
// @Tags gameUserSettings
// @Accept json
// @Produce json
// @Param data body types.CreateGameUserSettingsRequest true "gameUserSettings information"
// @Success 200 {object} types.CreateGameUserSettingsReply{}
// @Router /api/v1/gameUserSettings [post]
// @Security BearerAuth
func (h *gameUserSettingsHandler) Create(c *gin.Context) {
	form := &types.CreateGameUserSettingsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameUserSettings := &model.GameUserSettings{}
	err = copier.Copy(gameUserSettings, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameUserSettings)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameUserSettings)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameUserSettings.ID})
}

// DeleteByID delete a gameUserSettings by id
// @Summary Delete a gameUserSettings by id
// @Description Deletes a existing gameUserSettings identified by the given id in the path.
// @Tags gameUserSettings
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameUserSettingsByIDReply{}
// @Router /api/v1/gameUserSettings/{id} [delete]
// @Security BearerAuth
func (h *gameUserSettingsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameUserSettingsIDFromPath(c)
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

// UpdateByID update a gameUserSettings by id
// @Summary Update a gameUserSettings by id
// @Description Updates the specified gameUserSettings by given id in the path, support partial update.
// @Tags gameUserSettings
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameUserSettingsByIDRequest true "gameUserSettings information"
// @Success 200 {object} types.UpdateGameUserSettingsByIDReply{}
// @Router /api/v1/gameUserSettings/{id} [put]
// @Security BearerAuth
func (h *gameUserSettingsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameUserSettingsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameUserSettingsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameUserSettings := &model.GameUserSettings{}
	err = copier.Copy(gameUserSettings, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameUserSettings)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameUserSettings)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameUserSettings by id
// @Summary Get a gameUserSettings by id
// @Description Gets detailed information of a gameUserSettings specified by the given id in the path.
// @Tags gameUserSettings
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameUserSettingsByIDReply{}
// @Router /api/v1/gameUserSettings/{id} [get]
// @Security BearerAuth
func (h *gameUserSettingsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameUserSettingsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserSettings, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameUserSettingsObjDetail{}
	err = copier.Copy(data, gameUserSettings)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameUserSettings)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameUserSettings": data})
}

// List get a paginated list of gameUserSettingss by custom conditions
// @Summary Get a paginated list of gameUserSettingss by custom conditions
// @Description Returns a paginated list of gameUserSettings based on query filters, including page number and size.
// @Tags gameUserSettings
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameUserSettingssReply{}
// @Router /api/v1/gameUserSettings/list [post]
// @Security BearerAuth
func (h *gameUserSettingsHandler) List(c *gin.Context) {
	form := &types.ListGameUserSettingssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserSettingss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameUserSettingss(gameUserSettingss)
	if err != nil {
		response.Error(c, ecode.ErrListGameUserSettings)
		return
	}

	response.Success(c, gin.H{
		"gameUserSettingss": data,
		"total":             total,
	})
}

func getGameUserSettingsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameUserSettings(gameUserSettings *model.GameUserSettings) (*types.GameUserSettingsObjDetail, error) {
	data := &types.GameUserSettingsObjDetail{}
	err := copier.Copy(data, gameUserSettings)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameUserSettingss(fromValues []*model.GameUserSettings) ([]*types.GameUserSettingsObjDetail, error) {
	toValues := []*types.GameUserSettingsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameUserSettings(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
