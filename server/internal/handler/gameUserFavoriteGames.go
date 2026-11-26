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

var _ GameUserFavoriteGamesHandler = (*gameUserFavoriteGamesHandler)(nil)

// GameUserFavoriteGamesHandler defining the handler interface
type GameUserFavoriteGamesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameUserFavoriteGamesHandler struct {
	iDao dao.GameUserFavoriteGamesDao
}

// NewGameUserFavoriteGamesHandler creating the handler interface
func NewGameUserFavoriteGamesHandler() GameUserFavoriteGamesHandler {
	return &gameUserFavoriteGamesHandler{
		iDao: dao.NewGameUserFavoriteGamesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameUserFavoriteGamesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameUserFavoriteGames
// @Summary Create a new gameUserFavoriteGames
// @Description Creates a new gameUserFavoriteGames entity using the provided data in the request body.
// @Tags gameUserFavoriteGames
// @Accept json
// @Produce json
// @Param data body types.CreateGameUserFavoriteGamesRequest true "gameUserFavoriteGames information"
// @Success 200 {object} types.CreateGameUserFavoriteGamesReply{}
// @Router /api/v1/gameUserFavoriteGames [post]
// @Security BearerAuth
func (h *gameUserFavoriteGamesHandler) Create(c *gin.Context) {
	form := &types.CreateGameUserFavoriteGamesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameUserFavoriteGames := &model.GameUserFavoriteGames{}
	err = copier.Copy(gameUserFavoriteGames, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameUserFavoriteGames)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameUserFavoriteGames)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameUserFavoriteGames.ID})
}

// DeleteByID delete a gameUserFavoriteGames by id
// @Summary Delete a gameUserFavoriteGames by id
// @Description Deletes a existing gameUserFavoriteGames identified by the given id in the path.
// @Tags gameUserFavoriteGames
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameUserFavoriteGamesByIDReply{}
// @Router /api/v1/gameUserFavoriteGames/{id} [delete]
// @Security BearerAuth
func (h *gameUserFavoriteGamesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameUserFavoriteGamesIDFromPath(c)
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

// UpdateByID update a gameUserFavoriteGames by id
// @Summary Update a gameUserFavoriteGames by id
// @Description Updates the specified gameUserFavoriteGames by given id in the path, support partial update.
// @Tags gameUserFavoriteGames
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameUserFavoriteGamesByIDRequest true "gameUserFavoriteGames information"
// @Success 200 {object} types.UpdateGameUserFavoriteGamesByIDReply{}
// @Router /api/v1/gameUserFavoriteGames/{id} [put]
// @Security BearerAuth
func (h *gameUserFavoriteGamesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameUserFavoriteGamesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameUserFavoriteGamesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameUserFavoriteGames := &model.GameUserFavoriteGames{}
	err = copier.Copy(gameUserFavoriteGames, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameUserFavoriteGames)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameUserFavoriteGames)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameUserFavoriteGames by id
// @Summary Get a gameUserFavoriteGames by id
// @Description Gets detailed information of a gameUserFavoriteGames specified by the given id in the path.
// @Tags gameUserFavoriteGames
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameUserFavoriteGamesByIDReply{}
// @Router /api/v1/gameUserFavoriteGames/{id} [get]
// @Security BearerAuth
func (h *gameUserFavoriteGamesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameUserFavoriteGamesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserFavoriteGames, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameUserFavoriteGamesObjDetail{}
	err = copier.Copy(data, gameUserFavoriteGames)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameUserFavoriteGames)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameUserFavoriteGames": data})
}

// List get a paginated list of gameUserFavoriteGamess by custom conditions
// @Summary Get a paginated list of gameUserFavoriteGamess by custom conditions
// @Description Returns a paginated list of gameUserFavoriteGames based on query filters, including page number and size.
// @Tags gameUserFavoriteGames
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameUserFavoriteGamessReply{}
// @Router /api/v1/gameUserFavoriteGames/list [post]
// @Security BearerAuth
func (h *gameUserFavoriteGamesHandler) List(c *gin.Context) {
	form := &types.ListGameUserFavoriteGamessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameUserFavoriteGamess, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameUserFavoriteGamess(gameUserFavoriteGamess)
	if err != nil {
		response.Error(c, ecode.ErrListGameUserFavoriteGames)
		return
	}

	response.Success(c, gin.H{
		"gameUserFavoriteGamess": data,
		"total":        total,
	})
}

func getGameUserFavoriteGamesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameUserFavoriteGames(gameUserFavoriteGames *model.GameUserFavoriteGames) (*types.GameUserFavoriteGamesObjDetail, error) {
	data := &types.GameUserFavoriteGamesObjDetail{}
	err := copier.Copy(data, gameUserFavoriteGames)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameUserFavoriteGamess(fromValues []*model.GameUserFavoriteGames) ([]*types.GameUserFavoriteGamesObjDetail, error) {
	toValues := []*types.GameUserFavoriteGamesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameUserFavoriteGames(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
