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

var _ GameSupportQuestionsHandler = (*gameSupportQuestionsHandler)(nil)

// GameSupportQuestionsHandler defining the handler interface
type GameSupportQuestionsHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Support(c *gin.Context)
	NotificationStatus(c *gin.Context)
	MarkNoticesRead(c *gin.Context)
}

type gameSupportQuestionsHandler struct {
	iDao dao.GameSupportQuestionsDao
}

// NewGameSupportQuestionsHandler creating the handler interface
func NewGameSupportQuestionsHandler() GameSupportQuestionsHandler {
	return &gameSupportQuestionsHandler{
		iDao: dao.NewGameSupportQuestionsDao(
			database.GetDB(), // db driver is mysql
			cache.NewGameSupportQuestionsCache(database.GetCacheType()),
		),
	}
}

// Create a new gameSupportQuestions
// @Summary Create a new gameSupportQuestions
// @Description Creates a new gameSupportQuestions entity using the provided data in the request body.
// @Tags gameSupportQuestions
// @Accept json
// @Produce json
// @Param data body types.CreateGameSupportQuestionsRequest true "gameSupportQuestions information"
// @Success 200 {object} types.CreateGameSupportQuestionsReply{}
// @Router /api/v1/gameSupportQuestions [post]
// @Security BearerAuth
func (h *gameSupportQuestionsHandler) Create(c *gin.Context) {
	form := &types.CreateGameSupportQuestionsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	gameSupportQuestions := &model.GameSupportQuestions{}
	err = copier.Copy(gameSupportQuestions, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateGameSupportQuestions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, gameSupportQuestions)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": gameSupportQuestions.ID})
}

// DeleteByID delete a gameSupportQuestions by id
// @Summary Delete a gameSupportQuestions by id
// @Description Deletes a existing gameSupportQuestions identified by the given id in the path.
// @Tags gameSupportQuestions
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteGameSupportQuestionsByIDReply{}
// @Router /api/v1/gameSupportQuestions/{id} [delete]
// @Security BearerAuth
func (h *gameSupportQuestionsHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getGameSupportQuestionsIDFromPath(c)
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

// UpdateByID update a gameSupportQuestions by id
// @Summary Update a gameSupportQuestions by id
// @Description Updates the specified gameSupportQuestions by given id in the path, support partial update.
// @Tags gameSupportQuestions
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateGameSupportQuestionsByIDRequest true "gameSupportQuestions information"
// @Success 200 {object} types.UpdateGameSupportQuestionsByIDReply{}
// @Router /api/v1/gameSupportQuestions/{id} [put]
// @Security BearerAuth
func (h *gameSupportQuestionsHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getGameSupportQuestionsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateGameSupportQuestionsByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	gameSupportQuestions := &model.GameSupportQuestions{}
	err = copier.Copy(gameSupportQuestions, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDGameSupportQuestions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, gameSupportQuestions)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a gameSupportQuestions by id
// @Summary Get a gameSupportQuestions by id
// @Description Gets detailed information of a gameSupportQuestions specified by the given id in the path.
// @Tags gameSupportQuestions
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetGameSupportQuestionsByIDReply{}
// @Router /api/v1/gameSupportQuestions/{id} [get]
// @Security BearerAuth
func (h *gameSupportQuestionsHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getGameSupportQuestionsIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameSupportQuestions, err := h.iDao.GetByID(ctx, id)
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

	data := &types.GameSupportQuestionsObjDetail{}
	err = copier.Copy(data, gameSupportQuestions)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDGameSupportQuestions)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"gameSupportQuestions": data})
}

// List get a paginated list of gameSupportQuestionss by custom conditions
// @Summary Get a paginated list of gameSupportQuestionss by custom conditions
// @Description Returns a paginated list of gameSupportQuestions based on query filters, including page number and size.
// @Tags gameSupportQuestions
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListGameSupportQuestionssReply{}
// @Router /api/v1/gameSupportQuestions/list [post]
// @Security BearerAuth
func (h *gameSupportQuestionsHandler) List(c *gin.Context) {
	form := &types.ListGameSupportQuestionssRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	gameSupportQuestionss, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertGameSupportQuestionss(gameSupportQuestionss)
	if err != nil {
		response.Error(c, ecode.ErrListGameSupportQuestions)
		return
	}

	response.Success(c, gin.H{
		"gameSupportQuestionss": data,
		"total":                 total,
	})
}

func getGameSupportQuestionsIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertGameSupportQuestions(gameSupportQuestions *model.GameSupportQuestions) (*types.GameSupportQuestionsObjDetail, error) {
	data := &types.GameSupportQuestionsObjDetail{}
	err := copier.Copy(data, gameSupportQuestions)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertGameSupportQuestionss(fromValues []*model.GameSupportQuestions) ([]*types.GameSupportQuestionsObjDetail, error) {
	toValues := []*types.GameSupportQuestionsObjDetail{}
	for _, v := range fromValues {
		data, err := convertGameSupportQuestions(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
