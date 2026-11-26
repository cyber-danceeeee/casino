package ag

import "encoding/json"

type apiEnvelope struct {
	Success *bool           `json:"success,omitempty"`
	Code    string          `json:"code"`
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

type apiCallTrace struct {
	Request  map[string]any `json:"request,omitempty"`
	Response map[string]any `json:"response,omitempty"`
}

type apiRequest map[string]any

type categoryListResponse struct {
	CategoryDataList []categoryItem `json:"categoryDataList"`
	TotalSize        int            `json:"totalSize"`
}

type categoryItem struct {
	CategoryName string `json:"categoryName"`
	CategoryCode string `json:"categoryCode"`
}

type gameListResponse struct {
	CategoryCode string         `json:"categoryCode"`
	GameItemList []gameListItem `json:"gameItemList"`
	TotalSize    int            `json:"totalSize"`
}

type gameListItem struct {
	GameName         string `json:"gameName"`
	GameCode         string `json:"gameCode"`
	Status           int    `json:"status"`
	GameType         string `json:"gameType"`
	OfficialGameCode string `json:"officialGameCode"`

	ProviderCategoryCode string `json:"providerCategoryCode,omitempty"`
	ProviderCategoryName string `json:"providerCategoryName,omitempty"`
}

type launchResponse struct {
	GameCode   string `json:"gameCode"`
	PlayerID   string `json:"playerId"`
	GameURL    string `json:"gameUrl"`
	GameURLAlt string `json:"gameurl"`
	ExpireTime string `json:"expireTime"`
}

type providerError struct {
	Code          string
	Message       string
	StatusCode    int
	RequestTrace  string
	ResponseTrace string
}

func (e *providerError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" {
		return e.Code
	}
	return "ag provider error"
}

func (e *providerError) ProviderRequestTrace() string {
	if e == nil {
		return ""
	}
	return e.RequestTrace
}

func (e *providerError) ProviderResponseTrace() string {
	if e == nil {
		return ""
	}
	return e.ResponseTrace
}

type callbackEnvelope struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	Data    any    `json:"data"`
}
