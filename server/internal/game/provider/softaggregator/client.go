package softaggregator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"casino/internal/game/domain"
)

type Client struct {
	cfg        Config
	httpClient *http.Client
	endpoint   string
}

const maxProviderResponseBytes = 16 << 20

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
		cfg:      cfg,
		endpoint: cfg.EndpointURL(),
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

func (c *Client) call(ctx context.Context, method string, payload apiRequest, retry bool) (*apiEnvelope, error) {
	envelope, _, err := c.callWithTrace(ctx, method, payload, retry)
	return envelope, err
}

func (c *Client) callWithTrace(ctx context.Context, method string, payload apiRequest, retry bool) (*apiEnvelope, *apiCallTrace, error) {
	if !c.cfg.OutboundReady() {
		return nil, nil, domain.ErrProviderDisabled
	}
	if payload == nil {
		payload = apiRequest{}
	}
	payload["api_login"] = c.cfg.APILogin
	payload["api_password"] = c.cfg.APIPassword
	payload["method"] = method

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	var lastErr error
	var lastTrace *apiCallTrace
	attempts := 1
	if retry {
		attempts = 2
	}
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, lastTrace, ctx.Err()
			case <-time.After(150 * time.Millisecond):
			}
		}
		envelope, trace, err := c.doOnceWithTrace(ctx, body)
		lastTrace = trace
		if err == nil {
			return envelope, trace, nil
		}
		lastErr = err
		if !retry || !isRetryable(err) {
			break
		}
	}
	return nil, lastTrace, lastErr
}

func (c *Client) doOnce(ctx context.Context, body []byte) (*apiEnvelope, error) {
	envelope, _, err := c.doOnceWithTrace(ctx, body)
	return envelope, err
}

func (c *Client) doOnceWithTrace(ctx context.Context, body []byte) (*apiEnvelope, *apiCallTrace, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
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
		return nil, trace, &providerError{Message: fmt.Sprintf("softaggregator response exceeds %d bytes", maxProviderResponseBytes), StatusCode: resp.StatusCode}
	}
	trace.Response = map[string]any{
		"status":     resp.StatusCode,
		"statusText": resp.Status,
		"headers":    cloneHeader(resp.Header),
		"body":       decodeTraceJSON(respBody),
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, trace, traceError(trace, &providerError{Code: resp.StatusCode, Message: fmt.Sprintf("softaggregator http status %d", resp.StatusCode), StatusCode: resp.StatusCode})
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, trace, err
	}
	if envelope.Error != 0 {
		msg := strings.TrimSpace(envelope.Message)
		if msg == "" {
			msg = "softaggregator api error"
		}
		return nil, trace, traceError(trace, &providerError{Code: envelope.Error, Message: msg, StatusCode: resp.StatusCode})
	}
	return &envelope, trace, nil
}

func isRetryable(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	var providerErr *providerError
	if errors.As(err, &providerErr) {
		return providerErr.StatusCode == http.StatusTooManyRequests || providerErr.StatusCode >= http.StatusInternalServerError
	}
	return false
}

func RedactSensitive(value string) string {
	pattern := regexp.MustCompile(`(?i)((api_login|api_password|salt_key|password|key|token)(["']?\s*[:=]\s*))(("[^"]*")|('[^']*')|([^&\s,}]+))`)
	return pattern.ReplaceAllString(value, `${1}"REDACTED"`)
}

func traceError(trace *apiCallTrace, err *providerError) *providerError {
	if err == nil {
		return nil
	}
	err.RequestTrace = traceRequestJSON(trace)
	err.ResponseTrace = traceResponseJSON(trace)
	return err
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
