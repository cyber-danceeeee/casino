package handler

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/adminauth"
	"casino/internal/database"
	"casino/internal/gameadmin"
	"casino/internal/storage"
)

type AdminGamesHandler interface {
	ListProviders(*gin.Context)
	GetProvider(*gin.Context)
	SetProviderStatus(*gin.Context)
	SetProviderMaintenance(*gin.Context)
	TestProvider(*gin.Context)
	SyncProvider(*gin.Context)
	CancelProviderSync(*gin.Context)
	DeleteProvider(*gin.Context)
	SyncAll(*gin.Context)
	ListProviderGameTypes(*gin.Context)
	ListGames(*gin.Context)
	ListImportableGames(*gin.Context)
	ListImportableProviderGameTypes(*gin.Context)
	ImportGames(*gin.Context)
	GetGame(*gin.Context)
	UpdateGame(*gin.Context)
	BatchUpdateGames(*gin.Context)
	SetGameStatus(*gin.Context)
	SetGameCategories(*gin.Context)
	ApplyProviderCategoryMapping(*gin.Context)
	SetGameTags(*gin.Context)
	SetGameCollections(*gin.Context)
	LaunchDemo(*gin.Context)
	ListCategories(*gin.Context)
	GetCategory(*gin.Context)
	ListCategoryGames(*gin.Context)
	ListCategoryGameCandidates(*gin.Context)
	ListCategoryOperationLogs(*gin.Context)
	CreateCategory(*gin.Context)
	UpdateCategory(*gin.Context)
	AddCategoryGames(*gin.Context)
	UpdateCategoryGameSort(*gin.Context)
	RemoveCategoryGame(*gin.Context)
	DeleteCategory(*gin.Context)
	ListTags(*gin.Context)
	GetTag(*gin.Context)
	CreateTag(*gin.Context)
	UpdateTag(*gin.Context)
	SetTagGames(*gin.Context)
	DeleteTag(*gin.Context)
	ListCollections(*gin.Context)
	GetCollection(*gin.Context)
	CreateCollection(*gin.Context)
	UpdateCollection(*gin.Context)
	DeleteCollection(*gin.Context)
	ListLobbySections(*gin.Context)
	GetLobbySection(*gin.Context)
	CreateLobbySection(*gin.Context)
	UpdateLobbySection(*gin.Context)
	DeleteLobbySection(*gin.Context)
	ListSyncLogs(*gin.Context)
	GetSyncLog(*gin.Context)
	ListSyncLogItems(*gin.Context)
	GetSyncLogRawResponse(*gin.Context)
	ListSyncLogOperationLogs(*gin.Context)
	ListLaunchLogs(*gin.Context)
	GetLaunchLog(*gin.Context)
	UploadGameAsset(*gin.Context)
}
type adminGamesHandler struct{ service *gameadmin.Service }

func NewAdminGamesHandler() AdminGamesHandler {
	return &adminGamesHandler{service: gameadmin.New(database.GetDB())}
}

func (h *adminGamesHandler) ListProviders(c *gin.Context) {
	data, err := h.service.ListProviders(gameListQuery(c, providerQueryKeys))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) GetProvider(c *gin.Context) {
	data, err := h.service.ProviderDetail(c.Param("code"), adminContentScope(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) SetProviderStatus(c *gin.Context) {
	var in gameadmin.ReasonCommand
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.SetProviderStatus(c.Param("code"), in, gameActor(c)))
}
func (h *adminGamesHandler) SetProviderMaintenance(c *gin.Context) {
	var in gameadmin.ProviderMaintenanceCommand
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.SetProviderMaintenance(c.Param("code"), in, gameActor(c)))
}
func (h *adminGamesHandler) TestProvider(c *gin.Context) {
	data, err := h.service.TestProvider(c.Request.Context(), c.Param("code"), gameActor(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) SyncProvider(c *gin.Context) {
	var in gameadmin.SyncCommand
	if !gameBind(c, &in) {
		return
	}
	data, err := h.service.SyncProvider(c.Request.Context(), c.Param("code"), in.Reason, gameActor(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) CancelProviderSync(c *gin.Context) {
	var in gameadmin.SyncCommand
	if !gameBind(c, &in) {
		return
	}
	data, err := h.service.CancelProviderSync(c.Request.Context(), c.Param("code"), in.Reason, gameActor(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) DeleteProvider(c *gin.Context) {
	var in gameadmin.ReasonCommand
	if !gameBind(c, &in) {
		return
	}
	data, err := h.service.DeleteProvider(c.Request.Context(), c.Param("code"), in, gameActor(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) SyncAll(c *gin.Context) {
	var in gameadmin.SyncCommand
	if !gameBind(c, &in) {
		return
	}
	data, err := h.service.SyncAll(c.Request.Context(), in.Reason, gameActor(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) ListProviderGameTypes(c *gin.Context) {
	data, err := h.service.ListProviderGameTypes(c.Param("code"), adminContentScope(c))
	gameWrite(c, data, err)
}

func (h *adminGamesHandler) ListGames(c *gin.Context) {
	data, err := h.service.ListGames(gameListQuery(c, gameQueryKeys))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) ListImportableGames(c *gin.Context) {
	data, err := h.service.ListImportableGames(gameListQuery(c, gameQueryKeys))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) ListImportableProviderGameTypes(c *gin.Context) {
	data, err := h.service.ListImportableProviderGameTypes(c.Param("code"), adminContentScope(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) ImportGames(c *gin.Context) {
	var in gameadmin.GameImportInput
	if !gameBind(c, &in) {
		return
	}
	data, err := h.service.ImportGames(in, gameActor(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) GetGame(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.GameDetail(id, adminContentScope(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) UpdateGame(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.GameInput
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.UpdateGame(id, in, gameActor(c)))
}
func (h *adminGamesHandler) BatchUpdateGames(c *gin.Context) {
	var in gameadmin.GameBatchInput
	if !gameBind(c, &in) {
		return
	}
	data, err := h.service.BatchUpdateGames(in, gameActor(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) SetGameStatus(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.ReasonCommand
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.SetGameStatus(id, in, gameActor(c)))
}
func (h *adminGamesHandler) SetGameCategories(c *gin.Context) { h.gameRelations(c, "category") }
func (h *adminGamesHandler) ApplyProviderCategoryMapping(c *gin.Context) {
	var in gameadmin.ProviderCategoryMappingInput
	if !gameBind(c, &in) {
		return
	}
	data, err := h.service.ApplyProviderCategoryMapping(in, gameActor(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) SetGameTags(c *gin.Context)        { h.gameRelations(c, "tag") }
func (h *adminGamesHandler) SetGameCollections(c *gin.Context) { h.gameRelations(c, "collection") }
func (h *adminGamesHandler) gameRelations(c *gin.Context, kind string) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.RelationInput
	if !gameBind(c, &in) {
		return
	}
	var err error
	switch kind {
	case "category":
		err = h.service.SetGameCategories(id, in, gameActor(c))
	case "tag":
		err = h.service.SetGameTags(id, in, gameActor(c))
	default:
		err = h.service.SetGameCollections(id, in, gameActor(c))
	}
	gameWrite(c, nil, err)
}
func (h *adminGamesHandler) LaunchDemo(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.LaunchDemo(c.Request.Context(), id, adminContentScope(c))
	gameWrite(c, data, err)
}

func (h *adminGamesHandler) ListCategories(c *gin.Context) {
	data, err := h.service.ListCategories(gameListQuery(c, taxonomyQueryKeys))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) GetCategory(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.CategoryDetail(id, adminContentScope(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) ListCategoryGames(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.CategoryGames(id, gameListQuery(c, []string{"keyword"}))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) ListCategoryGameCandidates(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.CategoryGameCandidates(id, gameListQuery(c, categoryGameCandidateQueryKeys))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) ListCategoryOperationLogs(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.CategoryOperationLogs(id, gameListQuery(c, nil))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) CreateCategory(c *gin.Context) {
	var in gameadmin.CategoryInput
	if !gameBind(c, &in) {
		return
	}
	id, err := h.service.SaveCategory(0, in, gameActor(c))
	gameWrite(c, map[string]any{"id": id}, err)
}
func (h *adminGamesHandler) UpdateCategory(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.CategoryInput
	if !gameBind(c, &in) {
		return
	}
	_, err := h.service.SaveCategory(id, in, gameActor(c))
	gameWrite(c, nil, err)
}
func (h *adminGamesHandler) DeleteCategory(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.ReasonCommand
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.DeleteCategory(id, in.Reason, gameActor(c)))
}
func (h *adminGamesHandler) AddCategoryGames(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.CategoryGamesAddInput
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.AddCategoryGames(id, in, gameActor(c)))
}
func (h *adminGamesHandler) UpdateCategoryGameSort(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	gameID, err := strconv.ParseUint(c.Param("gameID"), 10, 64)
	if err != nil || gameID == 0 {
		gameError(c, &gameadmin.Error{Code: "INVALID_REQUEST", Message: "invalid game id"})
		return
	}
	var in gameadmin.CategoryGameSortInput
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.UpdateCategoryGameSort(id, gameID, in, gameActor(c)))
}
func (h *adminGamesHandler) RemoveCategoryGame(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	gameID, err := strconv.ParseUint(c.Param("gameID"), 10, 64)
	if err != nil || gameID == 0 {
		gameError(c, &gameadmin.Error{Code: "INVALID_REQUEST", Message: "invalid game id"})
		return
	}
	var in struct {
		Reason string `json:"reason" binding:"max=500"`
	}
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.RemoveCategoryGame(id, gameID, in.Reason, gameActor(c)))
}

func (h *adminGamesHandler) ListTags(c *gin.Context) {
	data, err := h.service.ListTags(gameListQuery(c, taxonomyQueryKeys))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) GetTag(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.TagDetail(id, adminContentScope(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) CreateTag(c *gin.Context) {
	var in gameadmin.TagInput
	if !gameBind(c, &in) {
		return
	}
	id, err := h.service.SaveTag(0, in, gameActor(c))
	gameWrite(c, map[string]any{"id": id}, err)
}
func (h *adminGamesHandler) UpdateTag(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.TagInput
	if !gameBind(c, &in) {
		return
	}
	_, err := h.service.SaveTag(id, in, gameActor(c))
	gameWrite(c, nil, err)
}
func (h *adminGamesHandler) DeleteTag(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.ReasonCommand
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.DeleteTag(id, in.Reason, gameActor(c)))
}
func (h *adminGamesHandler) SetTagGames(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.RelationInput
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.SetTagGames(id, in, gameActor(c)))
}

func (h *adminGamesHandler) ListCollections(c *gin.Context) {
	data, err := h.service.ListCollections(gameListQuery(c, collectionQueryKeys))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) GetCollection(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.CollectionDetail(id, adminContentScope(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) CreateCollection(c *gin.Context) {
	var in gameadmin.CollectionInput
	if !gameBind(c, &in) {
		return
	}
	id, err := h.service.SaveCollection(0, in, gameActor(c))
	gameWrite(c, map[string]any{"id": id}, err)
}
func (h *adminGamesHandler) UpdateCollection(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.CollectionInput
	if !gameBind(c, &in) {
		return
	}
	_, err := h.service.SaveCollection(id, in, gameActor(c))
	gameWrite(c, nil, err)
}
func (h *adminGamesHandler) DeleteCollection(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.ReasonCommand
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.DeleteCollection(id, in.Reason, gameActor(c)))
}

func (h *adminGamesHandler) ListLobbySections(c *gin.Context) {
	data, err := h.service.ListLobbySections(gameListQuery(c, lobbySectionQueryKeys))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) GetLobbySection(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.LobbySectionDetail(id, adminContentScope(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) CreateLobbySection(c *gin.Context) {
	var in gameadmin.LobbySectionInput
	if !gameBind(c, &in) {
		return
	}
	id, err := h.service.SaveLobbySection(0, in, gameActor(c))
	gameWrite(c, map[string]any{"id": id}, err)
}
func (h *adminGamesHandler) UpdateLobbySection(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in gameadmin.LobbySectionInput
	if !gameBind(c, &in) {
		return
	}
	_, err := h.service.SaveLobbySection(id, in, gameActor(c))
	gameWrite(c, nil, err)
}
func (h *adminGamesHandler) DeleteLobbySection(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	var in struct {
		Reason string `json:"reason" binding:"required,max=500"`
	}
	if !gameBind(c, &in) {
		return
	}
	gameWrite(c, nil, h.service.DeleteLobbySection(id, in.Reason, gameActor(c)))
}

func (h *adminGamesHandler) ListSyncLogs(c *gin.Context) {
	data, err := h.service.ListSyncLogs(gameListQuery(c, syncQueryKeys))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) GetSyncLog(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.SyncLogDetail(id, adminContentScope(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) ListSyncLogItems(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.SyncLogItems(id, c.Query("type"), gameListQuery(c, nil))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) GetSyncLogRawResponse(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.SyncLogRawResponse(id, adminContentScope(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) ListSyncLogOperationLogs(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.SyncLogOperationLogs(id, gameListQuery(c, nil))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) ListLaunchLogs(c *gin.Context) {
	data, err := h.service.ListLaunchLogs(gameListQuery(c, launchQueryKeys))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) GetLaunchLog(c *gin.Context) {
	id, ok := gamePathID(c)
	if !ok {
		return
	}
	data, err := h.service.LaunchLogDetail(id, adminContentScope(c))
	gameWrite(c, data, err)
}
func (h *adminGamesHandler) UploadGameAsset(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		gameError(c, &gameadmin.Error{Code: "FILE_REQUIRED", Message: "file is required"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".avif": true, ".gif": true, ".svg": true}
	if !allowed[ext] || file.Size <= 0 || file.Size > 10<<20 {
		gameError(c, &gameadmin.Error{Code: "INVALID_ASSET", Message: "unsupported file type or size"})
		return
	}
	name, err := storage.RandomName(ext)
	if err != nil {
		gameError(c, err)
		return
	}
	asset, err := storage.UploadFileHeader(c.Request.Context(), "admin/uploads/game/image/"+name, file)
	if err != nil {
		gameError(c, err)
		return
	}
	if err = h.service.RecordAssetUpload("image", asset.Key, gameActor(c)); err != nil {
		gameError(c, err)
		return
	}
	response.Success(c, asset)
}

func gameWrite(c *gin.Context, data any, err error) {
	if err != nil {
		gameError(c, err)
		return
	}
	response.Success(c, data)
}
func gameBind(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		gameError(c, &gameadmin.Error{Code: "INVALID_REQUEST", Message: "invalid request body"})
		return false
	}
	return true
}
func gamePathID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		gameError(c, &gameadmin.Error{Code: "INVALID_REQUEST", Message: "invalid id"})
		return 0, false
	}
	return id, true
}
func gameActor(c *gin.Context) gameadmin.Actor {
	actor, _ := adminauth.FromContext(c)
	correlation := c.GetHeader("X-Request-ID")
	if correlation == "" {
		correlation = c.GetString("request_id")
	}
	return gameadmin.Actor{ID: actor.ID, Username: actor.Username, Scope: adminContentScope(c), IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"), CorrelationID: correlation}
}
func gameListQuery(c *gin.Context, keys []string) gameadmin.ListQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	values := map[string]string{}
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			values[key] = value
		}
	}
	return gameadmin.ListQuery{Page: page, PageSize: pageSize, Values: values, Scope: adminContentScope(c)}
}
func gameError(c *gin.Context, err error) {
	var domain *gameadmin.Error
	if errors.As(err, &domain) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "msg": domain.Message, "data": gin.H{"errorCode": domain.Code}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "msg": "internal error", "data": gin.H{"errorCode": "INTERNAL_ERROR"}})
}

var providerQueryKeys = []string{"providerCode", "providerName", "status", "environment", "lastSyncFrom", "lastSyncTo"}
var gameQueryKeys = []string{"gameID", "gameName", "providerCode", "providerGameID", "providerCategory", "platformCategory", "gameType", "enabled", "supportsDemo", "supportsFreeSpins", "isHot", "isFeatured", "isNew", "country", "currency", "lastSyncedFrom", "lastSyncedTo"}
var categoryGameCandidateQueryKeys = []string{"providerCode", "source", "categoryCode", "keyword"}
var taxonomyQueryKeys = []string{"code", "nameZh", "nameEn", "enabled"}
var collectionQueryKeys = []string{"code", "nameZh", "nameEn", "enabled", "startFrom", "startTo"}
var lobbySectionQueryKeys = []string{"sectionCode", "sectionType", "enabled"}
var syncQueryKeys = []string{"providerCode", "status", "triggeredBy", "startedFrom", "startedTo", "completedFrom", "completedTo"}
var launchQueryKeys = []string{"userID", "username", "providerCode", "gameID", "providerGameID", "launchType", "status", "currency", "language", "country", "device", "createdFrom", "createdTo"}
