package provider

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type apiClientConfig struct {
	APIKey        string
	BaseURL       string
	DefaultRegion string
	Timeout       time.Duration
}

type apiClient struct {
	httpClient    *http.Client
	apiKey        string
	baseURL       string
	defaultRegion string
}

type apiError struct {
	statusCode int
	message    string
	body       string
}

func (e *apiError) Error() string {
	if e.message != "" {
		return fmt.Sprintf("api error (%d): %s", e.statusCode, e.message)
	}
	if e.body != "" {
		return fmt.Sprintf("api error (%d): %s", e.statusCode, e.body)
	}
	return fmt.Sprintf("api error (%d)", e.statusCode)
}

func newAPIClient(cfg apiClientConfig) *apiClient {
	base := strings.TrimRight(cfg.BaseURL, "/")
	return &apiClient{
		httpClient:    &http.Client{Timeout: cfg.Timeout},
		apiKey:        cfg.APIKey,
		baseURL:       base,
		defaultRegion: cfg.DefaultRegion,
	}
}

func generateIdempotencyKey() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func (c *apiClient) get(ctx context.Context, path string, query url.Values, out any) error {
	_, err := c.do(ctx, http.MethodGet, path, query, nil, "", out)
	return err
}

func (c *apiClient) post(ctx context.Context, path string, query url.Values, body any, idempotencyKey string, out any) error {
	_, err := c.do(ctx, http.MethodPost, path, query, body, idempotencyKey, out)
	return err
}

func (c *apiClient) delete(ctx context.Context, path string, query url.Values, idempotencyKey string, out any) error {
	_, err := c.do(ctx, http.MethodDelete, path, query, nil, idempotencyKey, out)
	return err
}

func (c *apiClient) do(ctx context.Context, method, path string, query url.Values, body any, idempotencyKey string, out any) (int, error) {
	fullURL := c.baseURL + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("marshal request body: %w", err)
		}
	}

	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var bodyReader io.Reader
		if payload != nil {
			bodyReader = bytes.NewReader(payload)
		}

		req, reqErr := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if reqErr != nil {
			return 0, fmt.Errorf("build request: %w", reqErr)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-API-Key", c.apiKey)
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}

		resp, doErr := c.httpClient.Do(req)
		if doErr != nil {
			if attempt < maxAttempts {
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
				continue
			}
			return 0, fmt.Errorf("perform request: %w", doErr)
		}

		bodyBytes, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return 0, fmt.Errorf("read response body: %w", readErr)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out != nil && len(bodyBytes) > 0 {
				if err = json.Unmarshal(bodyBytes, out); err != nil {
					return resp.StatusCode, fmt.Errorf("decode response: %w", err)
				}
			}
			return resp.StatusCode, nil
		}

		if (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) && attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		apiErr := decodeAPIError(resp.StatusCode, bodyBytes)
		return resp.StatusCode, apiErr
	}

	return 0, fmt.Errorf("request failed after retries")
}

func decodeAPIError(statusCode int, body []byte) error {
	if len(body) == 0 {
		return &apiError{statusCode: statusCode}
	}

	var payload struct {
		Error        string `json:"error"`
		Message      string `json:"message"`
		NeutronError struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"NeutronError"`
	}
	_ = json.Unmarshal(body, &payload)

	msg := payload.Error
	if msg == "" {
		msg = payload.Message
	}
	if msg == "" {
		msg = payload.NeutronError.Message
	}

	return &apiError{
		statusCode: statusCode,
		message:    msg,
		body:       string(body),
	}
}

func isNotFound(err error) bool {
	var ae *apiError
	if !asAPIError(err, &ae) {
		return false
	}
	return ae.statusCode == http.StatusNotFound
}

func isConflict(err error) bool {
	var ae *apiError
	if !asAPIError(err, &ae) {
		return false
	}
	return ae.statusCode == http.StatusConflict
}

func asAPIError(err error, target **apiError) bool {
	ae, ok := err.(*apiError)
	if !ok {
		return false
	}
	*target = ae
	return true
}

func describeAPIError(err error) string {
	var ae *apiError
	if !asAPIError(err, &ae) {
		return err.Error()
	}

	switch ae.statusCode {
	case http.StatusUnauthorized:
		return "unauthorized: check api_key and workspace access"
	case http.StatusBadRequest:
		if ae.message != "" {
			return "bad request: " + ae.message
		}
		return "bad request: verify input values"
	case http.StatusConflict:
		if ae.message != "" {
			return "conflict: " + ae.message
		}
		return "conflict: resource already exists, quota reached, or request is already in-flight"
	default:
		if ae.statusCode >= 500 {
			return "server error: retry later or contact support if persistent"
		}
		return ae.Error()
	}
}
