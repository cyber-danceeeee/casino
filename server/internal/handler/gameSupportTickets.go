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

var _ GameSupportTicketsHandler = (*gameSupportTicketsHandler)(nil)

// GameSupportTicketsHandler defining the handler interface
type GameSupportTicketsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type gameSupportTicketsHandler struct {
	iDao dao.GameSupportTicketsDao
}

// NewGameSupportTicketsHandler creating the handler interface
func NewGameSupportTicketsHandler() GameSupportTicketsHandler {
	return &gameSupportTicketsHandler{
		iDao: dao.NewGameSupportTicketsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameSupportTicketsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameSupportTickets
// @Summary Create a new gameSupportTickets
// @Description Creates a new gameSupportTickets entity using the provided data in the request body.
// @Tags gameSupportTickets
// @Accept json
// @Produce json
// @Param data body types.CreateGameSupportTicketsRequest true "gameSupportTickets information"
// @Success 200 {object} types.CreateGameSupportTicketsReply{}
// @Router /api/v1/gameSupportTickets [post]
// @Security BearerAuth
func (h *gameSupportTicketsHandler) Create(c *gin.Context) {
	form := &types.CreateGameSupportTicketsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameSupportTickets := &model.GameSupportTickets{}
	err = copier.Copy(gameSupportTickets, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameSupportTickets)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameSupportTickets)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameSupportTickets.ID})
}

// DeleteByID delete a gameSupportTickets by id
// @Summary Delete a gameSupportTickets by id
// @Description Deletes a existing gameSupportTickets identified by the given id in the path.
// @Tags gameSupportTickets
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameSupportTicketsByIDReply{}
// @Router /api/v1/gameSupportTickets/{id} [delete]
// @Security BearerAuth
func (h *gameSupportTicketsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameSupportTicketsIDFromPath(c)
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

// UpdateByID update a gameSupportTickets by id
// @Summary Update a gameSupportTickets by id
// @Description Updates the specified gameSupportTickets by given id in the path, support partial update.
// @Tags gameSupportTickets
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameSupportTicketsByIDRequest true "gameSupportTickets information"
// @Success 200 {object} types.UpdateGameSupportTicketsByIDReply{}
// @Router /api/v1/gameSupportTickets/{id} [put]
// @Security BearerAuth
func (h *gameSupportTicketsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameSupportTicketsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameSupportTicketsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameSupportTickets := &model.GameSupportTickets{}
	err = copier.Copy(gameSupportTickets, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameSupportTickets)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameSupportTickets)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameSupportTickets by id
// @Summary Get a gameSupportTickets by id
// @Description Gets detailed information of a gameSupportTickets specified by the given id in the path.
// @Tags gameSupportTickets
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameSupportTicketsByIDReply{}
// @Router /api/v1/gameSupportTickets/{id} [get]
// @Security BearerAuth
func (h *gameSupportTicketsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameSupportTicketsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameSupportTickets, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameSupportTicketsObjDetail{}
	err = copier.Copy(data, gameSupportTickets)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameSupportTickets)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameSupportTickets": data})
}

// List get a paginated list of gameSupportTicketss by custom conditions
// @Summary Get a paginated list of gameSupportTicketss by custom conditions
// @Description Returns a paginated list of gameSupportTickets based on query filters, including page number and size.
// @Tags gameSupportTickets
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameSupportTicketssReply{}
// @Router /api/v1/gameSupportTickets/list [post]
// @Security BearerAuth
func (h *gameSupportTicketsHandler) List(c *gin.Context) {
	form := &types.ListGameSupportTicketssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameSupportTicketss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameSupportTicketss(gameSupportTicketss)
	if err != nil {
		response.Error(c, ecode.ErrListGameSupportTickets)
		return
	}

	response.Success(c, gin.H{
		"gameSupportTicketss": data,
		"total":               total,
	})
}

func getGameSupportTicketsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameSupportTickets(gameSupportTickets *model.GameSupportTickets) (*types.GameSupportTicketsObjDetail, error) {
	data := &types.GameSupportTicketsObjDetail{}
	err := copier.Copy(data, gameSupportTickets)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameSupportTicketss(fromValues []*model.GameSupportTickets) ([]*types.GameSupportTicketsObjDetail, error) {
	toValues := []*types.GameSupportTicketsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameSupportTickets(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
