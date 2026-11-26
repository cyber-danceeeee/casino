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

var _ GameBannersHandler = (*gameBannersHandler)(nil)

// GameBannersHandler defining the handler interface
type GameBannersHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Home(c *gin.Context)
}

type gameBannersHandler struct {
	iDao dao.GameBannersDao
}

// NewGameBannersHandler creating the handler interface
func NewGameBannersHandler() GameBannersHandler {
	return &gameBannersHandler{
		iDao: dao.NewGameBannersDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameBannersCache(database.GetCacheType()),
		),
	}
}

// Create a new gameBanners
// @Summary Create a new gameBanners
// @Description Creates a new gameBanners entity using the provided data in the request body.
// @Tags gameBanners
// @Accept json
// @Produce json
// @Param data body types.CreateGameBannersRequest true "gameBanners information"
// @Success 200 {object} types.CreateGameBannersReply{}
// @Router /api/v1/gameBanners [post]
// @Security BearerAuth
func (h *gameBannersHandler) Create(c *gin.Context) {
	form := &types.CreateGameBannersRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameBanners := &model.GameBanners{}
	err = copier.Copy(gameBanners, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameBanners)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameBanners)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameBanners.ID})
}

// DeleteByID delete a gameBanners by id
// @Summary Delete a gameBanners by id
// @Description Deletes a existing gameBanners identified by the given id in the path.
// @Tags gameBanners
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameBannersByIDReply{}
// @Router /api/v1/gameBanners/{id} [delete]
// @Security BearerAuth
func (h *gameBannersHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameBannersIDFromPath(c)
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

// UpdateByID update a gameBanners by id
// @Summary Update a gameBanners by id
// @Description Updates the specified gameBanners by given id in the path, support partial update.
// @Tags gameBanners
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameBannersByIDRequest true "gameBanners information"
// @Success 200 {object} types.UpdateGameBannersByIDReply{}
// @Router /api/v1/gameBanners/{id} [put]
// @Security BearerAuth
func (h *gameBannersHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameBannersIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameBannersByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameBanners := &model.GameBanners{}
	err = copier.Copy(gameBanners, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameBanners)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameBanners)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameBanners by id
// @Summary Get a gameBanners by id
// @Description Gets detailed information of a gameBanners specified by the given id in the path.
// @Tags gameBanners
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameBannersByIDReply{}
// @Router /api/v1/gameBanners/{id} [get]
// @Security BearerAuth
func (h *gameBannersHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameBannersIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameBanners, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameBannersObjDetail{}
	err = copier.Copy(data, gameBanners)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameBanners)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameBanners": data})
}

// List get a paginated list of gameBannerss by custom conditions
// @Summary Get a paginated list of gameBannerss by custom conditions
// @Description Returns a paginated list of gameBanners based on query filters, including page number and size.
// @Tags gameBanners
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameBannerssReply{}
// @Router /api/v1/gameBanners/list [post]
// @Security BearerAuth
func (h *gameBannersHandler) List(c *gin.Context) {
	form := &types.ListGameBannerssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameBannerss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameBannerss(gameBannerss)
	if err != nil {
		response.Error(c, ecode.ErrListGameBanners)
		return
	}

	response.Success(c, gin.H{
		"gameBannerss": data,
		"total":        total,
	})
}

func getGameBannersIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameBanners(gameBanners *model.GameBanners) (*types.GameBannersObjDetail, error) {
	data := &types.GameBannersObjDetail{}
	err := copier.Copy(data, gameBanners)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameBannerss(fromValues []*model.GameBanners) ([]*types.GameBannersObjDetail, error) {
	toValues := []*types.GameBannersObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameBanners(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
