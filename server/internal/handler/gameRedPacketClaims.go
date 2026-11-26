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

var _ GameRedPacketClaimsHandler = (*gameRedPacketClaimsHandler)(nil)

// GameRedPacketClaimsHandler defining the handler interface
type GameRedPacketClaimsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameRedPacketClaimsHandler struct {
	iDao dao.GameRedPacketClaimsDao
}

// NewGameRedPacketClaimsHandler creating the handler interface
func NewGameRedPacketClaimsHandler() GameRedPacketClaimsHandler {
	return &gameRedPacketClaimsHandler{
		iDao: dao.NewGameRedPacketClaimsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameRedPacketClaimsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameRedPacketClaims
// @Summary Create a new gameRedPacketClaims
// @Description Creates a new gameRedPacketClaims entity using the provided data in the request body.
// @Tags gameRedPacketClaims
// @Accept json
// @Produce json
// @Param data body types.CreateGameRedPacketClaimsRequest true "gameRedPacketClaims information"
// @Success 200 {object} types.CreateGameRedPacketClaimsReply{}
// @Router /api/v1/gameRedPacketClaims [post]
// @Security BearerAuth
func (h *gameRedPacketClaimsHandler) Create(c *gin.Context) {
	form := &types.CreateGameRedPacketClaimsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameRedPacketClaims := &model.GameRedPacketClaims{}
	err = copier.Copy(gameRedPacketClaims, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameRedPacketClaims)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameRedPacketClaims)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameRedPacketClaims.ID})
}

// DeleteByID delete a gameRedPacketClaims by id
// @Summary Delete a gameRedPacketClaims by id
// @Description Deletes a existing gameRedPacketClaims identified by the given id in the path.
// @Tags gameRedPacketClaims
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameRedPacketClaimsByIDReply{}
// @Router /api/v1/gameRedPacketClaims/{id} [delete]
// @Security BearerAuth
func (h *gameRedPacketClaimsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameRedPacketClaimsIDFromPath(c)
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

// UpdateByID update a gameRedPacketClaims by id
// @Summary Update a gameRedPacketClaims by id
// @Description Updates the specified gameRedPacketClaims by given id in the path, support partial update.
// @Tags gameRedPacketClaims
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameRedPacketClaimsByIDRequest true "gameRedPacketClaims information"
// @Success 200 {object} types.UpdateGameRedPacketClaimsByIDReply{}
// @Router /api/v1/gameRedPacketClaims/{id} [put]
// @Security BearerAuth
func (h *gameRedPacketClaimsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameRedPacketClaimsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameRedPacketClaimsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameRedPacketClaims := &model.GameRedPacketClaims{}
	err = copier.Copy(gameRedPacketClaims, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameRedPacketClaims)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameRedPacketClaims)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameRedPacketClaims by id
// @Summary Get a gameRedPacketClaims by id
// @Description Gets detailed information of a gameRedPacketClaims specified by the given id in the path.
// @Tags gameRedPacketClaims
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameRedPacketClaimsByIDReply{}
// @Router /api/v1/gameRedPacketClaims/{id} [get]
// @Security BearerAuth
func (h *gameRedPacketClaimsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameRedPacketClaimsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameRedPacketClaims, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameRedPacketClaimsObjDetail{}
	err = copier.Copy(data, gameRedPacketClaims)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameRedPacketClaims)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameRedPacketClaims": data})
}

// List get a paginated list of gameRedPacketClaimss by custom conditions
// @Summary Get a paginated list of gameRedPacketClaimss by custom conditions
// @Description Returns a paginated list of gameRedPacketClaims based on query filters, including page number and size.
// @Tags gameRedPacketClaims
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameRedPacketClaimssReply{}
// @Router /api/v1/gameRedPacketClaims/list [post]
// @Security BearerAuth
func (h *gameRedPacketClaimsHandler) List(c *gin.Context) {
	form := &types.ListGameRedPacketClaimssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameRedPacketClaimss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameRedPacketClaimss(gameRedPacketClaimss)
	if err != nil {
		response.Error(c, ecode.ErrListGameRedPacketClaims)
		return
	}

	response.Success(c, gin.H{
		"gameRedPacketClaimss": data,
		"total":        total,
	})
}

func getGameRedPacketClaimsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameRedPacketClaims(gameRedPacketClaims *model.GameRedPacketClaims) (*types.GameRedPacketClaimsObjDetail, error) {
	data := &types.GameRedPacketClaimsObjDetail{}
	err := copier.Copy(data, gameRedPacketClaims)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameRedPacketClaimss(fromValues []*model.GameRedPacketClaims) ([]*types.GameRedPacketClaimsObjDetail, error) {
	toValues := []*types.GameRedPacketClaimsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameRedPacketClaims(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
