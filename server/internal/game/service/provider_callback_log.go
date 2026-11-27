package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"casino/internal/game/domain"
	"casino/internal/model"
)

func (s *WalletCallbackService) BeginCallbackLog(ctx context.Context, request domain.ProviderCallbackRequest, correlationID string) uint64 {
	providerCode := strings.ToLower(strings.TrimSpace(request.ProviderCode))
	bodyValues := callbackBodyValues(request.Body)
	providerPlayerID := callbackRequestValue(request, bodyValues, "userCode", "username", "playerId", "player_id", "uid", "userId", "user_id")
	providerTransactionID := callbackRequestValue(request, bodyValues, "transactionCode", "call_id", "transactionId", "transaction_id", "betId", "bet_id", "reqTraceId", "req_trace_id", "merchantBetId", "merchant_bet_id")
	if strings.EqualFold(callbackOperationName(request), string(domain.WalletOperationBalance)) && providerTransactionID == "" {
		providerTransactionID = "balance:" + callbackRequestValue(request, bodyValues, "timestamp", "time")
	}
	amountMinor, _ := strconv.ParseInt(callbackRequestValue(request, bodyValues, "amount", "amountMinor", "amount_minor", "betAmount", "bet_amount", "winAmount", "win_amount"), 10, 64)

	row := &model.GameProviderCallbackLogs{
		CallbackID: callbackID(), ProviderCode: providerCode,
		Operation: strings.ToUpper(strings.TrimSpace(callbackOperationName(request))), HTTPMethod: request.Method,
		ProviderPlayerID: providerPlayerID, ProviderTransactionID: providerTransactionID,
		ProviderRoundID: callbackRequestValue(request, bodyValues, "round_id", "roundId", "roundID", "providerRoundID", "provider_round_id"),
		ProviderGameID:  callbackRequestValue(request, bodyValues, "game_id", "gameId", "gameCode", "game_code", "providerGameID", "provider_game_id"),
		AmountMinor:     amountMinor, Currency: strings.ToUpper(callbackRequestValue(request, bodyValues, "currency", "currencyCode", "currency_code", "curr", "cur")),
		RequestPath: request.Path, RequestHeaders: jsonText(redactHeaders(request.Headers)),
		QueryParameters: jsonText(redactMap(request.Query)), RequestBody: redactBody(request.Body),
		SourceIP: request.ClientIP, SignatureResult: "PENDING", TimestampValidation: "PENDING",
		ProcessingResult: "PENDING", HTTPStatus: 200, NormalizedCommand: "{}", WalletResult: "{}",
		ProviderResponse: "{}", CorrelationID: correlationID,
	}
	if providerCode != "" && providerPlayerID != "" {
		if player, err := s.repo.PlayerByProviderID(ctx, providerCode, providerPlayerID); err == nil {
			row.UserID = player.UserID
			row.SiteID = player.SiteID
			row.SiteID = player.SiteID
		}
	}
	if err := s.repo.DB().WithContext(ctx).Create(row).Error; err != nil {
		return 0
	}
	return row.ID
}

func callbackRequestValue(request domain.ProviderCallbackRequest, body map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(request.QueryValue(key)); value != "" {
			return value
		}
	}
	for _, key := range keys {
		if value := callbackJSONString(body[callbackKey(key)]); value != "" {
			return value
		}
	}
	return ""
}

func callbackBodyValues(body []byte) map[string]any {
	values := map[string]any{}
	if len(body) == 0 {
		return values
	}
	var payload any
	if json.Unmarshal(body, &payload) != nil {
		return values
	}
	collectCallbackValues(values, payload)
	return values
}

func collectCallbackValues(values map[string]any, payload any) {
	switch typed := payload.(type) {
	case map[string]any:
		for key, value := range typed {
			values[callbackKey(key)] = value
			collectCallbackValues(values, value)
		}
	case []any:
		for _, value := range typed {
			collectCallbackValues(values, value)
		}
	}
}

func callbackKey(key string) string {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, " ", "")
	return normalized
}

func callbackJSONString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func (s *WalletCallbackService) FinishCallbackLog(ctx context.Context, id uint64, command *domain.WalletCommand, result *domain.WalletResult, providerResponse *domain.ProviderCallbackResponse, processingErr error, startedAt time.Time) {
	if id == 0 {
		return
	}
	updates := map[string]any{
		"duration_ms":       uint64(time.Since(startedAt).Milliseconds()),
		"processing_result": "SUCCESS", "signature_result": "VALID", "timestamp_validation": "VALID",
	}
	if command != nil {
		updates["site_id"] = command.SiteID
		updates["operation"] = string(command.Operation)
		updates["provider_player_id"] = command.ProviderPlayerID
		updates["provider_transaction_id"] = command.ProviderTransactionID
		updates["provider_round_id"] = command.ProviderRoundID
		updates["provider_game_id"] = command.ProviderGameID
		updates["amount_minor"] = command.AmountMinor
		updates["currency"] = command.Currency
		updates["normalized_command"] = jsonText(map[string]any{
			"providerCode": command.ProviderCode, "operation": command.Operation,
			"providerPlayerID": command.ProviderPlayerID, "providerTransactionID": command.ProviderTransactionID,
			"providerRoundID": command.ProviderRoundID, "providerGameID": command.ProviderGameID,
			"amountMinor": command.AmountMinor, "currency": command.Currency,
			"relatedTransactionID": command.RelatedTransactionID, "timestamp": command.Timestamp,
		})
		if player, err := s.repo.PlayerByProviderID(ctx, command.ProviderCode, command.ProviderPlayerID); err == nil {
			updates["user_id"] = player.UserID
			updates["site_id"] = player.SiteID
		}
	}
	if result != nil {
		if parsed, err := strconv.ParseUint(result.UserID, 10, 64); err == nil {
			updates["user_id"] = parsed
		}
		updates["wallet_result"] = jsonText(map[string]any{
			"success": result.Success, "internalTransactionID": result.InternalTransactionID,
			"providerTransactionID": result.ProviderTransactionID, "balanceMinor": result.BalanceMinor,
			"currency": result.Currency, "duplicate": result.Duplicate,
		})
	}
	if providerResponse != nil {
		updates["http_status"] = providerResponse.StatusCode
		updates["provider_response"] = jsonText(providerResponse.Body)
	}
	if processingErr != nil {
		updates["processing_result"] = "FAILED"
		updates["failure_reason"] = safeFailure(processingErr)
		var walletErr *domain.WalletError
		if errors.As(processingErr, &walletErr) {
			switch walletErr.Code {
			case domain.WalletErrorInvalidSignature:
				updates["signature_result"] = "INVALID"
				updates["timestamp_validation"] = "NOT_CHECKED"
			case domain.WalletErrorInvalidTimestamp:
				updates["timestamp_validation"] = "INVALID"
			default:
				updates["signature_result"] = "VALID"
				updates["timestamp_validation"] = "VALID"
			}
			if walletErr.Code == domain.WalletErrorTransactionConflict || walletErr.Code == domain.WalletErrorDuplicateTransaction {
				updates["marked_for_investigation"] = true
			}
		}
	}
	_ = s.repo.DB().WithContext(ctx).Model(&model.GameProviderCallbackLogs{}).Where("id = ?", id).Updates(updates).Error
}

func callbackOperationName(request domain.ProviderCallbackRequest) string {
	if values := request.Query["action"]; len(values) > 0 {
		return values[0]
	}
	return request.Operation
}

func redactHeaders(headers map[string][]string) map[string][]string {
	result := make(map[string][]string, len(headers))
	for key, values := range headers {
		if sensitiveKey(key) {
			result[key] = []string{"REDACTED"}
			continue
		}
		result[key] = append([]string(nil), values...)
	}
	return result
}

func redactMap(values map[string][]string) map[string][]string {
	result := make(map[string][]string, len(values))
	for key, items := range values {
		if sensitiveKey(key) {
			result[key] = []string{"REDACTED"}
			continue
		}
		result[key] = append([]string(nil), items...)
	}
	return result
}

func redactBody(body []byte) string {
	if len(body) == 0 {
		return "{}"
	}
	var value any
	if json.Unmarshal(body, &value) != nil {
		return jsonText(map[string]string{"content": "REDACTED_NON_JSON_BODY"})
	}
	return jsonText(redactJSONValue(value))
}

func redactJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			if sensitiveKey(key) {
				out[key] = "REDACTED"
			} else {
				out[key] = redactJSONValue(item)
			}
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = redactJSONValue(item)
		}
		return out
	default:
		return value
	}
}

func sensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(key), "-", "_"), " ", "_"))
	switch normalized {
	case "authorization", "cookie", "set_cookie", "key", "signature", "sign", "salt_key", "api_password", "password", "token", "access_token", "refresh_token", "launch_token", "secret":
		return true
	default:
		return false
	}
}

func jsonText(value any) string {
	data, err := json.Marshal(value)
	if err != nil || string(data) == "null" {
		return "{}"
	}
	return string(data)
}

func safeFailure(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 500 {
		return message[:500]
	}
	return message
}

func callbackID() string {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err == nil {
		return "CB" + hex.EncodeToString(buffer)
	}
	return "CB" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
