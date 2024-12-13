// internal/api/client.go
package api

import (
	"fmt"

	"ovh-terminal/internal/config"
	"ovh-terminal/internal/logger"

	ovh "github.com/ovh/go-ovh/ovh"
)

// Client wraps the OVH API client with additional functionality
type Client struct {
	client *ovh.Client
	logger *logger.Logger
}

// NewClient creates a new OVH API client
func NewClient(cfg *config.AccountConfig, log *logger.Logger) (*Client, error) {
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

	return &Client{
		client: client,
		logger: log,
	}, nil
}

// Get performs a GET request to the OVH API
func (c *Client) Get(path string, result interface{}) error {
	c.logger.Debug("Making GET request", "path", path)

	err := c.client.Get(path, result)
	if err != nil {
		// Handle different error types
		if ovhErr, ok := err.(*ovh.APIError); ok {
			switch ovhErr.Code {
			case 401, 403:
				return NewAuthError("Authentication failed", err)
			default:
				return NewAPIError("API request failed", err, map[string]interface{}{
					"status": ovhErr.Code,
					"method": "GET",
					"path":   path,
				})
			}
		}
		return NewAPIError("Request failed", err, nil)
	}

	c.logger.Debug("GET request successful", "path", path)
	return nil
}

// Post performs a POST request to the OVH API
func (c *Client) Post(path string, payload interface{}, result interface{}) error {
	c.logger.Debug("Making POST request", "path", path)

	err := c.client.Post(path, payload, result)
	if err != nil {
		if ovhErr, ok := err.(*ovh.APIError); ok {
			switch ovhErr.Code {
			case 401, 403:
				return NewAuthError("Authentication failed", err)
			default:
				return NewAPIError("API request failed", err, map[string]interface{}{
					"status": ovhErr.Code,
					"method": "POST",
					"path":   path,
				})
			}
		}
		return NewAPIError("Request failed", err, nil)
	}

	c.logger.Debug("POST request successful", "path", path)
	return nil
}
