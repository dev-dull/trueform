package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultTimeout     = 30 * time.Second
	defaultPingPeriod  = 30 * time.Second
	defaultPongTimeout = 60 * time.Second
	maxReconnectDelay  = 30 * time.Second
	apiPath            = "/api/current"
)

// Client represents a TrueNAS API client
type Client struct {
	host      string
	apiKey    string
	verifySSL bool
	timeout   time.Duration

	conn      *websocket.Conn
	connMu    sync.Mutex
	requestID int64

	// Response channels keyed by request ID
	responses   map[int64]chan *JSONRPCResponse
	responsesMu sync.Mutex

	// Context for managing goroutines
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Connection state
	connected   bool
	connectedMu sync.RWMutex
}

// Config holds configuration for the TrueNAS client
type Config struct {
	Host      string
	APIKey    string
	VerifySSL bool
	Timeout   time.Duration
}

// NewClient creates a new TrueNAS API client
func NewClient(cfg *Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		host:      cfg.Host,
		apiKey:    cfg.APIKey,
		verifySSL: cfg.VerifySSL,
		timeout:   timeout,
		responses: make(map[int64]chan *JSONRPCResponse),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Connect establishes a WebSocket connection and authenticates
func (c *Client) Connect(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.isConnected() {
		return nil
	}

	// Build WebSocket URL
	u := url.URL{
		Scheme: "wss",
		Host:   c.host,
		Path:   apiPath,
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !c.verifySSL,
	}

	dialer := websocket.Dialer{
		TLSClientConfig:  tlsConfig,
		HandshakeTimeout: c.timeout,
	}

	// Connect
	conn, _, err := dialer.DialContext(ctx, u.String(), http.Header{})
	if err != nil {
		return NewConnectionError("failed to connect to TrueNAS", err)
	}

	c.conn = conn
	c.setConnected(true)

	// Start response reader
	c.wg.Add(1)
	go c.readResponses()

	// Authenticate with API key
	if err := c.authenticate(ctx); err != nil {
		c.close()
		return err
	}

	return nil
}

// authenticate performs API key authentication
func (c *Client) authenticate(ctx context.Context) error {
	var result bool
	err := c.Call(ctx, "auth.login_with_api_key", []interface{}{c.apiKey}, &result)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	if !result {
		return fmt.Errorf("authentication failed: invalid API key")
	}
	return nil
}

// Call makes a JSON-RPC call and waits for the response
func (c *Client) Call(ctx context.Context, method string, params interface{}, result interface{}) error {
	// Ensure we're connected
	if !c.isConnected() {
		if err := c.Connect(ctx); err != nil {
			return err
		}
	}

	// Generate request ID
	id := atomic.AddInt64(&c.requestID, 1)

	// Create response channel
	respChan := make(chan *JSONRPCResponse, 1)
	c.responsesMu.Lock()
	c.responses[id] = respChan
	c.responsesMu.Unlock()

	defer func() {
		c.responsesMu.Lock()
		delete(c.responses, id)
		c.responsesMu.Unlock()
	}()

	// Build request
	req := NewRequest(id, method, params)

	// Send request
	c.connMu.Lock()
	err := c.conn.WriteJSON(req)
	c.connMu.Unlock()

	if err != nil {
		return NewConnectionError("failed to send request", err)
	}

	// Wait for response with timeout
	select {
	case resp := <-respChan:
		if resp.Error != nil {
			return NewAPIError(resp.Error)
		}
		if result != nil && resp.Result != nil {
			if err := json.Unmarshal(resp.Result, result); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(c.timeout):
		return fmt.Errorf("request timeout")
	}
}

// readResponses reads responses from the WebSocket connection
func (c *Client) readResponses() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.connMu.Lock()
		conn := c.conn
		c.connMu.Unlock()

		if conn == nil {
			return
		}

		var resp JSONRPCResponse
		if err := conn.ReadJSON(&resp); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.setConnected(false)
				return
			}
			// Connection error - mark as disconnected
			c.setConnected(false)
			return
		}

		// Route response to waiting caller
		c.responsesMu.Lock()
		if ch, ok := c.responses[resp.ID]; ok {
			ch <- &resp
		}
		c.responsesMu.Unlock()
	}
}

// Close closes the client connection
func (c *Client) Close() error {
	c.cancel()
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.close()
}

func (c *Client) close() error {
	c.setConnected(false)
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

func (c *Client) isConnected() bool {
	c.connectedMu.RLock()
	defer c.connectedMu.RUnlock()
	return c.connected
}

func (c *Client) setConnected(connected bool) {
	c.connectedMu.Lock()
	defer c.connectedMu.Unlock()
	c.connected = connected
}

// Query performs a query operation with optional filtering
func (c *Client) Query(ctx context.Context, resource string, params *QueryParams, result interface{}) error {
	method := resource + ".query"
	var args []interface{}
	if params != nil {
		// Convert filters to the format expected by TrueNAS
		queryOptions := map[string]interface{}{}
		if params.Limit > 0 {
			queryOptions["limit"] = params.Limit
		}
		if params.Offset > 0 {
			queryOptions["offset"] = params.Offset
		}
		if params.Count {
			queryOptions["count"] = params.Count
		}
		if len(params.OrderBy) > 0 {
			queryOptions["order_by"] = params.OrderBy
		}
		if len(params.Select) > 0 {
			queryOptions["select"] = params.Select
		}

		if len(params.Filters) > 0 {
			args = append(args, params.Filters)
		} else {
			args = append(args, []interface{}{})
		}
		if len(queryOptions) > 0 {
			args = append(args, queryOptions)
		}
	}
	return c.Call(ctx, method, args, result)
}

// GetInstance retrieves a single instance by ID
func (c *Client) GetInstance(ctx context.Context, resource string, id interface{}, result interface{}) error {
	method := resource + ".get_instance"
	return c.Call(ctx, method, []interface{}{id}, result)
}

// Create creates a new resource
func (c *Client) Create(ctx context.Context, resource string, data interface{}, result interface{}) error {
	method := resource + ".create"
	return c.Call(ctx, method, []interface{}{data}, result)
}

// Update updates an existing resource
func (c *Client) Update(ctx context.Context, resource string, id interface{}, data interface{}, result interface{}) error {
	method := resource + ".update"
	return c.Call(ctx, method, []interface{}{id, data}, result)
}

// Delete deletes a resource
func (c *Client) Delete(ctx context.Context, resource string, id interface{}) error {
	method := resource + ".delete"
	return c.Call(ctx, method, []interface{}{id}, nil)
}

// DeleteWithOptions deletes a resource with additional options
func (c *Client) DeleteWithOptions(ctx context.Context, resource string, id interface{}, options interface{}) error {
	method := resource + ".delete"
	return c.Call(ctx, method, []interface{}{id, options}, nil)
}
