package mimo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	defaultBaseURL    = "https://api.xiaomimimo.com/v1"
	defaultTimeout    = 600 * time.Second
	defaultMaxRetries = 3
	defaultRateLimit  = 100
)

// Client 是 MiMo API 的 HTTP 客户端，支持重试和限流。
type Client struct {
	config      *ClientConfig
	httpClient  *http.Client
	mu          sync.Mutex
	lastReqAt   time.Time
	reqCount    int
	windowStart time.Time
}

// NewClient 创建一个新的 MiMo API 客户端。
func NewClient(config *ClientConfig) *Client {
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = defaultMaxRetries
	}
	if config.InitialWait == 0 {
		config.InitialWait = 1 * time.Second
	}
	if config.MaxWait == 0 {
		config.MaxWait = 30 * time.Second
	}
	if config.Multiplier == 0 {
		config.Multiplier = 2.0
	}
	if config.RateLimit == 0 {
		config.RateLimit = defaultRateLimit
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		windowStart: time.Now(),
	}
}

// doRequest 执行 HTTP 请求，包含重试和限流逻辑。
func (c *Client) doRequest(ctx context.Context, endpoint string, reqBody interface{}, respBody interface{}) error {
	if err := c.checkRateLimit(); err != nil {
		return err
	}

	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			wait := c.calculateBackoff(attempt)
			log.Printf("[MiMo] 重试 %d/%d，等待 %v", attempt, c.config.MaxRetries, wait)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		err := c.executeRequest(ctx, endpoint, reqBody, respBody)
		if err == nil {
			return nil
		}

		lastErr = err

		if apiErr, ok := err.(*APIError); ok {
			if !apiErr.IsRetryable() {
				return err
			}
			if apiErr.StatusCode == 429 && apiErr.RetryAfter > 0 {
				log.Printf("[MiMo] 收到限流响应，建议等待 %v", apiErr.RetryAfter)
			}
		}

		log.Printf("[MiMo] 请求失败: %v", err)
	}

	return fmt.Errorf("重试 %d 次后仍然失败: %w", c.config.MaxRetries, lastErr)
}

// executeRequest 执行单次 HTTP 请求。
func (c *Client) executeRequest(ctx context.Context, endpoint string, reqBody interface{}, respBody interface{}) error {
	url := c.config.BaseURL + endpoint

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.config.APIKey)
	req.Header.Set("Accept", "application/json")

	c.updateLastReqTime()

	log.Printf("[MiMo] 发送请求: %s", endpoint)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp.StatusCode, body, resp.Header)
	}

	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}

	return nil
}

// handleErrorResponse 处理错误响应。
func (c *Client) handleErrorResponse(statusCode int, body []byte, headers http.Header) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		RequestID:  headers.Get("X-Request-Id"),
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		apiErr.Type = errResp.Error.Type
		apiErr.Code = errResp.Error.Code
		apiErr.Message = errResp.Error.Message
	} else {
		apiErr.Message = string(body)
	}

	if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
		if seconds, err := time.ParseDuration(retryAfter + "s"); err == nil {
			apiErr.RetryAfter = seconds
		}
	}

	return apiErr
}

// calculateBackoff 计算指数退避时间。
func (c *Client) calculateBackoff(attempt int) time.Duration {
	wait := float64(c.config.InitialWait)
	for i := 1; i < attempt; i++ {
		wait *= c.config.Multiplier
	}

	duration := time.Duration(wait)
	if duration > c.config.MaxWait {
		duration = c.config.MaxWait
	}

	return duration
}

// checkRateLimit 检查限流。
func (c *Client) checkRateLimit() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	if now.Sub(c.windowStart) >= time.Minute {
		c.reqCount = 0
		c.windowStart = now
	}

	if c.reqCount >= c.config.RateLimit {
		wait := time.Minute - now.Sub(c.windowStart)
		return fmt.Errorf("达到限流限制 (%d RPM)，需等待 %v", c.config.RateLimit, wait)
	}

	c.reqCount++
	return nil
}

// updateLastReqTime 更新最后请求时间。
func (c *Client) updateLastReqTime() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastReqAt = time.Now()
}

// GetConfig 返回客户端配置（用于测试）。
func (c *Client) GetConfig() *ClientConfig {
	return c.config
}
