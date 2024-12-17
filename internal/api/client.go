// internal/api/client.go
package api

import (
	"fmt"
	"time"

	"ovh-terminal/internal/config"
	"ovh-terminal/internal/logger"

	ovh "github.com/ovh/go-ovh/ovh"
)

// ClientOption defines a function type for configuring the client
type ClientOption func(*Client)

// RetryConfig holds retry-related configuration
type RetryConfig struct {
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	RetryOnCode []int
}

// Client wraps the OVH API client with additional functionality
type Client struct {
	client  *ovh.Client
	logger  *logger.Logger
	retry   RetryConfig
	timeout time.Duration
}

// Default configuration values
var defaultRetryConfig = RetryConfig{
	MaxRetries:  3,
	BaseDelay:   time.Second,
	MaxDelay:    time.Second * 10,
	RetryOnCode: []int{408, 429, 500, 502, 503, 504},
}

// WithRetry configures retry behavior
func WithRetry(config RetryConfig) ClientOption {
	return func(c *Client) {
		c.retry = config
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// NewClient creates a new OVH API client
func NewClient(
	cfg *config.AccountConfig,
	log *logger.Logger,
	opts ...ClientOption,
) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("account configuration is required")
	}

	// Create client with configuration
	client, err := ovh.NewClient(
		cfg.Endpoint,
		cfg.AppKey,
		cfg.AppSecret,
		cfg.ConsumerKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OVH client: %w", err)
	}

	// Create wrapped client with default settings
	c := &Client{
		client:  client,
		logger:  log,
		retry:   defaultRetryConfig,
		timeout: time.Second * 30,
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// shouldRetry determines if a request should be retried
func (c *Client) shouldRetry(err error, attempt int) bool {
	if attempt >= c.retry.MaxRetries {
		return false
	}

	if ovhErr, ok := err.(*ovh.APIError); ok {
		for _, code := range c.retry.RetryOnCode {
			if ovhErr.Code == code {
				return true
			}
		}
	}

	return false
}

// calculateDelay determines the delay before the next retry
func (c *Client) calculateDelay(attempt int) time.Duration {
	delay := c.retry.BaseDelay * time.Duration(1<<uint(attempt))
	if delay > c.retry.MaxDelay {
		delay = c.retry.MaxDelay
	}
	return delay
}

// executeWithRetry handles request execution with retry logic
func (c *Client) executeWithRetry(operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < c.retry.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := c.calculateDelay(attempt)
			c.logger.Debug("Retrying request",
				"operation", operation,
				"attempt", attempt+1,
				"delay", delay.String())
			time.Sleep(delay)
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		if !c.shouldRetry(err, attempt) {
			break
		}
	}

	return c.handleAPIError("GET", operation, lastErr)
}

// Get performs a GET request to the OVH API
func (c *Client) Get(path string, result interface{}) error {
	c.logger.Debug("Making GET request", "path", path)

	return c.executeWithRetry(path, func() error {
		return c.client.Get(path, result)
	})
}

// Post performs a POST request to the OVH API
func (c *Client) Post(path string, payload interface{}, result interface{}) error {
	c.logger.Debug("Making POST request", "path", path)

	return c.executeWithRetry(path, func() error {
		return c.client.Post(path, payload, result)
	})
}

// handleAPIError processes API errors and returns appropriate error types
func (c *Client) handleAPIError(method, path string, err error) error {
	if err == nil {
		return nil
	}

	if ovhErr, ok := err.(*ovh.APIError); ok {
		switch ovhErr.Code {
		case 401, 403:
			return NewAuthError("Authentication failed", err)
		default:
			return NewAPIError("API request failed", err, map[string]interface{}{
				"status": ovhErr.Code,
				"method": method,
				"path":   path,
			})
		}
	}
	return NewAPIError("Request failed", err, nil)
}

