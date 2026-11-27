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

var _ GameGamesHandler = (*gameGamesHandler)(nil)

// GameGamesHandler defining the handler interface
type GameGamesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Bootstrap(c *gin.Context)
	PublicList(c *gin.Context)
	PublicDetail(c *gin.Context)
	Launch(c *gin.Context)
	FavoriteStatus(c *gin.Context)
	ToggleFavorite(c *gin.Context)
}

type gameGamesHandler struct {
	iDao dao.GameGamesDao
}

// NewGameGamesHandler creating the handler interface
func NewGameGamesHandler() GameGamesHandler {
	return &gameGamesHandler{
		iDao: dao.NewGameGamesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameGamesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameGames
// @Summary Create a new gameGames
// @Description Creates a new gameGames entity using the provided data in the request body.
// @Tags gameGames
// @Accept json
// @Produce json
// @Param data body types.CreateGameGamesRequest true "gameGames information"
// @Success 200 {object} types.CreateGameGamesReply{}
// @Router /api/v1/gameGames [post]
// @Security BearerAuth
func (h *gameGamesHandler) Create(c *gin.Context) {
	form := &types.CreateGameGamesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameGames := &model.GameGames{}
	err = copier.Copy(gameGames, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameGames)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameGames)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameGames.ID})
}

// DeleteByID delete a gameGames by id
// @Summary Delete a gameGames by id
// @Description Deletes a existing gameGames identified by the given id in the path.
// @Tags gameGames
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameGamesByIDReply{}
// @Router /api/v1/gameGames/{id} [delete]
// @Security BearerAuth
func (h *gameGamesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameGamesIDFromPath(c)
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

// UpdateByID update a gameGames by id
// @Summary Update a gameGames by id
// @Description Updates the specified gameGames by given id in the path, support partial update.
// @Tags gameGames
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameGamesByIDRequest true "gameGames information"
// @Success 200 {object} types.UpdateGameGamesByIDReply{}
// @Router /api/v1/gameGames/{id} [put]
// @Security BearerAuth
func (h *gameGamesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameGamesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameGamesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameGames := &model.GameGames{}
	err = copier.Copy(gameGames, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameGames)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameGames)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameGames by id
// @Summary Get a gameGames by id
// @Description Gets detailed information of a gameGames specified by the given id in the path.
// @Tags gameGames
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameGamesByIDReply{}
// @Router /api/v1/gameGames/{id} [get]
// @Security BearerAuth
func (h *gameGamesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameGamesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameGames, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameGamesObjDetail{}
	err = copier.Copy(data, gameGames)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameGames)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameGames": data})
}

// List get a paginated list of gameGamess by custom conditions
// @Summary Get a paginated list of gameGamess by custom conditions
// @Description Returns a paginated list of gameGames based on query filters, including page number and size.
// @Tags gameGames
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameGamessReply{}
// @Router /api/v1/gameGames/list [post]
// @Security BearerAuth
func (h *gameGamesHandler) List(c *gin.Context) {
	form := &types.ListGameGamessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameGamess, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameGamess(gameGamess)
	if err != nil {
		response.Error(c, ecode.ErrListGameGames)
		return
	}

	response.Success(c, gin.H{
		"gameGamess": data,
		"total":      total,
	})
}

func getGameGamesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameGames(gameGames *model.GameGames) (*types.GameGamesObjDetail, error) {
	data := &types.GameGamesObjDetail{}
	err := copier.Copy(data, gameGames)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameGamess(fromValues []*model.GameGames) ([]*types.GameGamesObjDetail, error) {
	toValues := []*types.GameGamesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameGames(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
