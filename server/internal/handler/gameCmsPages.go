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

var _ GameCmsPagesHandler = (*gameCmsPagesHandler)(nil)

// GameCmsPagesHandler defining the handler interface
type GameCmsPagesHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Page(c *gin.Context)
}

type gameCmsPagesHandler struct {
	iDao dao.GameCmsPagesDao
}

// NewGameCmsPagesHandler creating the handler interface
func NewGameCmsPagesHandler() GameCmsPagesHandler {
	return &gameCmsPagesHandler{
		iDao: dao.NewGameCmsPagesDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameCmsPagesCache(database.GetCacheType()),
		),
	}
}

// Create a new gameCmsPages
// @Summary Create a new gameCmsPages
// @Description Creates a new gameCmsPages entity using the provided data in the request body.
// @Tags gameCmsPages
// @Accept json
// @Produce json
// @Param data body types.CreateGameCmsPagesRequest true "gameCmsPages information"
// @Success 200 {object} types.CreateGameCmsPagesReply{}
// @Router /api/v1/gameCmsPages [post]
// @Security BearerAuth
func (h *gameCmsPagesHandler) Create(c *gin.Context) {
	form := &types.CreateGameCmsPagesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameCmsPages := &model.GameCmsPages{}
	err = copier.Copy(gameCmsPages, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameCmsPages)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameCmsPages)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameCmsPages.ID})
}

// DeleteByID delete a gameCmsPages by id
// @Summary Delete a gameCmsPages by id
// @Description Deletes a existing gameCmsPages identified by the given id in the path.
// @Tags gameCmsPages
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameCmsPagesByIDReply{}
// @Router /api/v1/gameCmsPages/{id} [delete]
// @Security BearerAuth
func (h *gameCmsPagesHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameCmsPagesIDFromPath(c)
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

// UpdateByID update a gameCmsPages by id
// @Summary Update a gameCmsPages by id
// @Description Updates the specified gameCmsPages by given id in the path, support partial update.
// @Tags gameCmsPages
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameCmsPagesByIDRequest true "gameCmsPages information"
// @Success 200 {object} types.UpdateGameCmsPagesByIDReply{}
// @Router /api/v1/gameCmsPages/{id} [put]
// @Security BearerAuth
func (h *gameCmsPagesHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameCmsPagesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameCmsPagesByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameCmsPages := &model.GameCmsPages{}
	err = copier.Copy(gameCmsPages, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameCmsPages)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameCmsPages)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameCmsPages by id
// @Summary Get a gameCmsPages by id
// @Description Gets detailed information of a gameCmsPages specified by the given id in the path.
// @Tags gameCmsPages
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameCmsPagesByIDReply{}
// @Router /api/v1/gameCmsPages/{id} [get]
// @Security BearerAuth
func (h *gameCmsPagesHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameCmsPagesIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameCmsPages, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameCmsPagesObjDetail{}
	err = copier.Copy(data, gameCmsPages)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameCmsPages)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameCmsPages": data})
}

// List get a paginated list of gameCmsPagess by custom conditions
// @Summary Get a paginated list of gameCmsPagess by custom conditions
// @Description Returns a paginated list of gameCmsPages based on query filters, including page number and size.
// @Tags gameCmsPages
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameCmsPagessReply{}
// @Router /api/v1/gameCmsPages/list [post]
// @Security BearerAuth
func (h *gameCmsPagesHandler) List(c *gin.Context) {
	form := &types.ListGameCmsPagessRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameCmsPagess, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameCmsPagess(gameCmsPagess)
	if err != nil {
		response.Error(c, ecode.ErrListGameCmsPages)
		return
	}

	response.Success(c, gin.H{
		"gameCmsPagess": data,
		"total":         total,
	})
}

func getGameCmsPagesIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameCmsPages(gameCmsPages *model.GameCmsPages) (*types.GameCmsPagesObjDetail, error) {
	data := &types.GameCmsPagesObjDetail{}
	err := copier.Copy(data, gameCmsPages)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameCmsPagess(fromValues []*model.GameCmsPages) ([]*types.GameCmsPagesObjDetail, error) {
	toValues := []*types.GameCmsPagesObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameCmsPages(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
