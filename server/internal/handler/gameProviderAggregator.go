package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"

	"casino/internal/adminauth"
	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/game/domain"
	gameservice "casino/internal/game/service"
	"casino/internal/model"
)

type GameProviderAggregatorHandler interface {
	GameList(c *gin.Context)
	SyncProviderGames(c *gin.Context)
	Launch(c *gin.Context)
	LaunchDemo(c *gin.Context)
	ExitSession(c *gin.Context)
	HeartbeatSession(c *gin.Context)
	Callback(c *gin.Context)
}

type gameProviderAggregatorHandler struct {
	runtime *gameservice.Runtime
}

type gameProviderLaunchRequest struct {
	Currency  string `json:"currency"`
	Language  string `json:"language"`
	Country   string `json:"country"`
	Device    string `json:"device"`
	ReturnURL string `json:"return_url"`
	SessionID string `json:"sessionId"`
}

type gameSessionExitRequest struct {
	Reason string `json:"reason"`
}

func NewGameProviderAggregatorHandler() GameProviderAggregatorHandler {
	return &gameProviderAggregatorHandler{runtime: gameservice.NewRuntime(database.GetDB())}
}

func (h *gameProviderAggregatorHandler) GameList(c *gin.Context) {
	providerCode := strings.ToLower(strings.TrimSpace(c.Param("providerCode")))
	if providerCode == "" {
		h5Error(c, http.StatusBadRequest, "invalid provider")
		return
	}

	db := database.GetDB()
	scope := h5ContentScope(c)
	var provider model.GameGameProviders
	if err := content.ApplySite(db.Model(&model.GameGameProviders{}), "", scope).Where("provider_code = ?", providerCode).First(&provider).Error; err != nil {
		h5Error(c, http.StatusNotFound, "provider not found")
		return
	}

	var games []model.GameGames
	q, _ := h5BuildPublicGameQuery(c, db, scope, h5GameListFilters{ProviderCode: providerCode})
	if err := q.Order("game_games.hot_score DESC, game_games.sort ASC,game_games.id ASC").Find(&games).Error; err != nil {
		h5Error(c, http.StatusInternalServerError, "game list unavailable")
		return
	}

	items := h5GameDTOs(c.Request.Context(), games, map[uint64]model.GameGameProviders{provider.ID: provider})
	response.Success(c, gin.H{
		"providerCode": providerCode,
		"items":        items,
		"list":         items,
		"total":        len(items),
	})
}

func (h *gameProviderAggregatorHandler) SyncProviderGames(c *gin.Context) {
	providerCode := strings.ToLower(strings.TrimSpace(c.Param("providerCode")))
	if providerCode == "" {
		h5Error(c, http.StatusBadRequest, "invalid provider")
		return
	}
	_, ok := adminauth.FromContext(c)
	if !ok {
		h5Error(c, http.StatusForbidden, "administrator authentication is required")
		return
	}
	summary, err := h.runtime.GameService.SyncProviderGames(c.Request.Context(), providerCode)
	if err != nil {
		h5Error(c, http.StatusBadGateway, err.Error())
		return
	}
	response.Success(c, summary)
}

func (h *gameProviderAggregatorHandler) Launch(c *gin.Context) {
	h.launch(c, false)
}

func (h *gameProviderAggregatorHandler) LaunchDemo(c *gin.Context) {
	h.launch(c, true)
}

func (h *gameProviderAggregatorHandler) HeartbeatSession(c *gin.Context) {
	db := database.GetDB()
	user, ok := providerAuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	sessionID := strings.TrimSpace(c.Param("sessionId"))
	scope := h5UserContentScope(db, user.ID, h5ContentScope(c))
	if err := h.runtime.LaunchService.Heartbeat(c.Request.Context(), scope, sessionID, user.ID); err != nil {
		h5Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c)
}

func (h *gameProviderAggregatorHandler) ExitSession(c *gin.Context) {
	req := &gameSessionExitRequest{Reason: "UNKNOWN"}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(req); err != nil {
			h5Error(c, http.StatusBadRequest, "invalid request")
			return
		}
	}
	db := database.GetDB()
	user, ok := providerAuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	sessionID := strings.TrimSpace(c.Param("sessionId"))
	scope := h5UserContentScope(db, user.ID, h5ContentScope(c))
	if err := h.runtime.LaunchService.Exit(c.Request.Context(), scope, sessionID, user.ID, req.Reason); err != nil {
		h5Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c)
}

func (h *gameProviderAggregatorHandler) launch(c *gin.Context, demo bool) {
	req := &gameProviderLaunchRequest{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(req); err != nil {
			h5Error(c, http.StatusBadRequest, "invalid request")
			return
		}
	}

	db := database.GetDB()
	user, ok := providerAuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	scope := h5UserContentScope(db, user.ID, h5ContentScope(c))
	if blocked, code, message := h5MaintenanceBlocked(c, db, scope); blocked {
		h5MachineError(c, http.StatusServiceUnavailable, code, message)
		return
	}
	gameID := strings.TrimSpace(c.Param("gameId"))
	sessionID := strings.TrimSpace(req.SessionID)
	if len(sessionID) > 128 {
		h5Error(c, http.StatusBadRequest, "invalid session id")
		return
	}
	if sessionID == "" {
		sessionID = h5Token("launch")
	}
	result, err := h.runtime.LaunchService.Launch(c.Request.Context(), gameservice.LaunchRequest{

		SiteID:       scope.SiteID,
		ProviderCode: providerCodeParam(c),
		GameID:       gameID,
		UserID:       user.ID,
		Currency:     req.Currency,
		Language:     req.Language,
		Country:      req.Country,
		Device:       req.Device,
		ReturnURL:    req.ReturnURL,
		SessionID:    sessionID,
		IPAddress:    c.ClientIP(),
		Demo:         demo,
	})
	if err != nil {
		var walletError *domain.WalletError
		if errors.As(err, &walletError) {
			status := http.StatusBadRequest
			if walletError.Code == domain.WalletErrorProviderMaintenance || walletError.Code == domain.WalletErrorGameMaintenance {
				status = http.StatusServiceUnavailable
			}
			h5MachineError(c, status, string(walletError.Code), walletError.Error())
		} else {
			h5Error(c, http.StatusBadGateway, "game provider is unavailable")
		}
		return
	}

	response.Success(c, result)
}

func providerCodeParam(c *gin.Context) string {
	if value := strings.TrimSpace(c.Param("providerCode")); value != "" {
		return value
	}
	return strings.TrimSpace(c.Param("id"))
}

func (h *gameProviderAggregatorHandler) Callback(c *gin.Context) {
	startedAt := time.Now()
	providerCode := strings.ToLower(strings.TrimSpace(c.Param("providerCode")))
	adapter, err := h.runtime.CallbackRegistry.Get(providerCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": 2, "balance": 0})
		return
	}

	body, _ := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	request := domain.ProviderCallbackRequest{
		ProviderCode: providerCode,
		Operation:    callbackOperationParam(c),
		Method:       c.Request.Method,
		Path:         c.Request.URL.Path,
		Query:        c.Request.URL.Query(),
		Headers:      c.Request.Header.Clone(),
		Body:         body,
		ClientIP:     c.ClientIP(),
		ReceivedAt:   time.Now(),
	}

	ctx := c.Request.Context()
	callbackLogID := h.runtime.WalletService.BeginCallbackLog(ctx, request, c.GetHeader("X-Request-ID"))
	if err := adapter.ValidateRequest(ctx, request); err != nil {
		providerResponse := adapter.BuildErrorResponse(ctx, request, err)
		h.runtime.WalletService.FinishCallbackLog(ctx, callbackLogID, nil, nil, providerResponse, err, startedAt)
		writeProviderCallbackResponse(c, providerResponse)
		return
	}
	command, err := adapter.ParseCommand(ctx, request)
	if err != nil {
		providerResponse := adapter.BuildErrorResponse(ctx, request, err)
		h.runtime.WalletService.FinishCallbackLog(ctx, callbackLogID, nil, nil, providerResponse, err, startedAt)
		writeProviderCallbackResponse(c, providerResponse)
		return
	}
	command.ProviderCode = providerCode

	result, err := h.runtime.WalletService.Execute(ctx, *command)
	if err != nil {
		providerResponse := adapter.BuildErrorResponse(ctx, request, err)
		h.runtime.WalletService.FinishCallbackLog(ctx, callbackLogID, command, nil, providerResponse, err, startedAt)
		writeProviderCallbackResponse(c, providerResponse)
		return
	}
	providerResponse, err := adapter.BuildSuccessResponse(ctx, *command, *result)
	if err != nil {
		providerResponse = adapter.BuildErrorResponse(ctx, request, err)
		h.runtime.WalletService.FinishCallbackLog(ctx, callbackLogID, command, result, providerResponse, err, startedAt)
		writeProviderCallbackResponse(c, providerResponse)
		return
	}
	h.runtime.WalletService.FinishCallbackLog(ctx, callbackLogID, command, result, providerResponse, nil, startedAt)
	writeProviderCallbackResponse(c, providerResponse)
}

func callbackOperationParam(c *gin.Context) string {
	operation := strings.TrimSpace(c.Param("operation"))
	if operation != "" {
		return operation
	}
	path := strings.Trim(strings.ToLower(c.Request.URL.Path), "/")
	if strings.HasSuffix(path, "player/info") || strings.HasSuffix(path, "player-info") {
		return "player-info"
	}
	return ""
}

func writeProviderCallbackResponse(c *gin.Context, providerResponse *domain.ProviderCallbackResponse) {
	if providerResponse == nil {
		c.JSON(http.StatusOK, gin.H{"error": 2, "balance": 0})
		return
	}
	for key, value := range providerResponse.Headers {
		c.Header(key, value)
	}
	status := providerResponse.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	c.JSON(status, providerResponse.Body)
}

func providerAuthenticatedUser(db *gorm.DB, c *gin.Context) (model.GameUsers, bool) {
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		return model.GameUsers{}, false
	}
	return user, true
}
