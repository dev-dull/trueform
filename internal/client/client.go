package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultTimeout     = 10 * time.Second
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

	if c.isConnected() {
		c.connMu.Unlock()
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

	// Create a net.Dialer with explicit timeouts to ensure TCP connection attempts timeout
	netDialer := &net.Dialer{
		Timeout:   c.timeout,
		KeepAlive: 30 * time.Second,
	}

	dialer := websocket.Dialer{
		TLSClientConfig:  tlsConfig,
		HandshakeTimeout: c.timeout,
		NetDialContext:   netDialer.DialContext,
	}

	// Create a context with timeout for the connection attempt
	connectCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Connect
	conn, _, err := dialer.DialContext(connectCtx, u.String(), http.Header{})
	if err != nil {
		return NewConnectionError(c.host, err)
	}

	// Set initial read deadline
	conn.SetReadDeadline(time.Now().Add(c.timeout))

	c.conn = conn
	c.setConnected(true)

	// Start response reader
	c.wg.Add(1)
	go c.readResponses()

	// Release the lock before calling authenticate, which calls Call(), which needs the lock
	c.connMu.Unlock()

	// Authenticate with API key
	if err := c.authenticate(ctx); err != nil {
		c.connMu.Lock()
		c.close()
		c.connMu.Unlock()
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

	// Send request with write deadline
	c.connMu.Lock()
	c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	err := c.conn.WriteJSON(req)
	c.connMu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
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
		return fmt.Errorf("request timeout after %v", c.timeout)
	}
}

// readResponses reads responses from the WebSocket connection
func (c *Client) readResponses() {
	defer func() {
		c.wg.Done()
	}()

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
			// Check if it's a timeout - if so, check if we should continue
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Check if context is cancelled
				select {
				case <-c.ctx.Done():
					return
				default:
					// Refresh deadline and continue
					c.connMu.Lock()
					if c.conn != nil {
						c.conn.SetReadDeadline(time.Now().Add(c.timeout))
					}
					c.connMu.Unlock()
					continue
				}
			}
			// Other connection error - mark as disconnected
			c.setConnected(false)
			return
		}


		// Successfully read a response - refresh deadline for next read
		c.connMu.Lock()
		if c.conn != nil {
			c.conn.SetReadDeadline(time.Now().Add(c.timeout))
		}
		c.connMu.Unlock()

		// Route response to waiting caller
		c.responsesMu.Lock()
		if ch, ok := c.responses[resp.ID]; ok {
			ch <- &resp
		} else {
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

// WaitForJob waits for a TrueNAS job to complete and returns the result
func (c *Client) WaitForJob(ctx context.Context, jobID int64, timeout time.Duration) (map[string]interface{}, error) {
	deadline := time.Now().Add(timeout)
	pollInterval := 2 * time.Second

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for job %d to complete", jobID)
		}

		var job map[string]interface{}
		err := c.Call(ctx, "core.get_jobs", []interface{}{
			[][]interface{}{{"id", "=", jobID}},
		}, &job)

		// The API returns an array, get the first element
		var jobs []map[string]interface{}
		err = c.Call(ctx, "core.get_jobs", []interface{}{
			[][]interface{}{{"id", "=", jobID}},
		}, &jobs)
		if err != nil {
			return nil, fmt.Errorf("failed to query job status: %w", err)
		}

		if len(jobs) == 0 {
			return nil, fmt.Errorf("job %d not found", jobID)
		}

		job = jobs[0]
		state, _ := job["state"].(string)

		switch state {
		case "SUCCESS":
			if result, ok := job["result"].(map[string]interface{}); ok {
				return result, nil
			}
			// Some jobs return simple values or nil
			return job, nil
		case "FAILED":
			errMsg := "job failed"
			if e, ok := job["error"].(string); ok {
				errMsg = e
			}
			return nil, fmt.Errorf("job %d failed: %s", jobID, errMsg)
		case "ABORTED":
			return nil, fmt.Errorf("job %d was aborted", jobID)
		default:
			// Job still running, wait and poll again
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(pollInterval):
			}
		}
	}
}

// CreateWithJob creates a resource and waits for the job to complete
func (c *Client) CreateWithJob(ctx context.Context, resource string, data interface{}, timeout time.Duration) (map[string]interface{}, error) {
	method := resource + ".create"

	var jobID float64
	err := c.Call(ctx, method, []interface{}{data}, &jobID)
	if err != nil {
		return nil, err
	}

	return c.WaitForJob(ctx, int64(jobID), timeout)
}
