package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type clientHTTPSessionState struct {
	mu sync.Mutex
}

func (c *Client) callTransport(ctx context.Context, method string, params any) (json.RawMessage, error) {
	if _, ok := c.t.(*httpTransport); ok {
		c.httpSession.mu.Lock()
		defer c.httpSession.mu.Unlock()
	}
	res, err := c.t.call(ctx, method, params)
	if err == nil || method == "initialize" || !isHTTPSessionExpired(err) {
		return res, err
	}
	slog.Info("plugin: MCP HTTP session recovery started", "server", c.name)
	if initErr := c.initializeSessionWith(ctx, false, c.t.call, c.t.notify); initErr != nil {
		slog.Warn("plugin: MCP HTTP session recovery failed", "server", c.name, "stage", "initialize")
		return nil, fmt.Errorf("%w; reinitialize failed: %w", err, initErr)
	}
	res, retryErr := c.t.call(ctx, method, params)
	if retryErr != nil {
		slog.Warn("plugin: MCP HTTP session recovery failed", "server", c.name, "stage", "retry")
		return nil, retryErr
	}
	slog.Info("plugin: MCP HTTP session recovery completed", "server", c.name)
	return res, nil
}

func isHTTPSessionExpired(err error) bool {
	var expired *httpSessionExpiredError
	return errors.As(err, &expired)
}
