package ag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"casino/internal/game/domain"
)

const maxProviderResponseBytes = 16 << 20

type Client struct {
	cfg        Config
	httpClient *http.Client
	baseURL    string
	now        func() time.Time
}

type ClientOption func(*Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func NewClient(cfg Config, opts ...ClientOption) *Client {
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	connectTimeout := cfg.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 3 * time.Second
	}
	client := &Client{
		cfg:     cfg,
		baseURL: cfg.EndpointURL(),
		now:     time.Now,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   connectTimeout,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   connectTimeout,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func (c *Client) post(ctx context.Context, path string, payload apiRequest) (*apiEnvelope, error) {
	envelope, _, err := c.postWithTrace(ctx, path, payload)
	return envelope, err
}

func (c *Client) postWithTrace(ctx context.Context, path string, payload apiRequest) (*apiEnvelope, *apiCallTrace, error) {
	if !c.cfg.OutboundReady() {
		return nil, nil, domain.ErrProviderDisabled
	}
	if payload == nil {
		payload = apiRequest{}
	}
	if _, ok := payload["reqTraceId"]; !ok {
		payload["reqTraceId"] = "ag-" + nonce()
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}
	url := strings.TrimRight(c.baseURL, "/") + "/" + strings.TrimLeft(path, "/")
	timestamp := timestampMillis(c.now())
	requestNonce := nonce()
	signature := sign(c.cfg.SecretKey, c.cfg.MerchantCode, timestamp, requestNonce, body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-MERCHANT-CODE", c.cfg.MerchantCode)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-NONCE", requestNonce)
	req.Header.Set("X-SIGN", signature)
	req.Header.Set("X-CONTENT-PROCESSING-TYPE", signType)

	trace := &apiCallTrace{
		Request: map[string]any{
			"method":  req.Method,
			"url":     req.URL.String(),
			"headers": cloneHeader(req.Header),
			"query":   req.URL.Query(),
			"body":    decodeTraceJSON(body),
		},
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, trace, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxProviderResponseBytes+1))
	if err != nil {
		return nil, trace, err
	}
	if len(respBody) > maxProviderResponseBytes {
		return nil, trace, &providerError{Message: fmt.Sprintf("ag response exceeds %d bytes", maxProviderResponseBytes), StatusCode: resp.StatusCode}
	}
	trace.Response = map[string]any{
		"status":     resp.StatusCode,
		"statusText": resp.Status,
		"headers":    cloneHeader(resp.Header),
		"body":       decodeTraceJSON(respBody),
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, trace, traceError(trace, &providerError{Code: fmt.Sprintf("%d", resp.StatusCode), Message: fmt.Sprintf("ag http status %d", resp.StatusCode), StatusCode: resp.StatusCode})
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, trace, err
	}
	if !strings.EqualFold(strings.TrimSpace(envelope.Code), "C10000") {
		msg := strings.TrimSpace(envelope.Msg)
		if msg == "" {
			msg = "ag api error"
		}
		return nil, trace, traceError(trace, &providerError{Code: envelope.Code, Message: msg, StatusCode: resp.StatusCode})
	}
	return &envelope, trace, nil
}

func cloneHeader(header http.Header) map[string][]string {
	out := make(map[string][]string, len(header))
	for key, values := range header {
		out[key] = append([]string(nil), values...)
	}
	return out
}

func decodeTraceJSON(raw []byte) any {
	var value any
	if len(raw) > 0 && json.Unmarshal(raw, &value) == nil {
		return value
	}
	return string(raw)
}

func traceRequestJSON(trace *apiCallTrace) string {
	if trace == nil || trace.Request == nil {
		return ""
	}
	raw, _ := json.Marshal(trace.Request)
	return string(raw)
}

func traceResponseJSON(trace *apiCallTrace) string {
	if trace == nil || trace.Response == nil {
		return ""
	}
	raw, _ := json.Marshal(trace.Response)
	return string(raw)
}

func traceError(trace *apiCallTrace, err *providerError) *providerError {
	if err == nil {
		return nil
	}
	err.RequestTrace = traceRequestJSON(trace)
	err.ResponseTrace = traceResponseJSON(trace)
	return err
}
