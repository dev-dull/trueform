package client

import "encoding/json"

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      int64       `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	ID      int64           `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// NewRequest creates a new JSON-RPC 2.0 request
func NewRequest(id int64, method string, params interface{}) *JSONRPCRequest {
	return &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}
}

// QueryParams represents common query parameters for list operations
type QueryParams struct {
	Limit   int                      `json:"limit,omitempty"`
	Offset  int                      `json:"offset,omitempty"`
	Count   bool                     `json:"count,omitempty"`
	OrderBy []string                 `json:"order_by,omitempty"`
	Select  []string                 `json:"select,omitempty"`
	Filters [][]interface{}          `json:"filters,omitempty"`
	Options map[string]interface{}   `json:"options,omitempty"`
}

// NewQueryParams creates a new QueryParams with defaults
func NewQueryParams() *QueryParams {
	return &QueryParams{}
}

// WithFilter adds a filter to the query
func (q *QueryParams) WithFilter(field string, operator string, value interface{}) *QueryParams {
	q.Filters = append(q.Filters, []interface{}{field, operator, value})
	return q
}

// WithLimit sets the limit for the query
func (q *QueryParams) WithLimit(limit int) *QueryParams {
	q.Limit = limit
	return q
}

// WithOffset sets the offset for the query
func (q *QueryParams) WithOffset(offset int) *QueryParams {
	q.Offset = offset
	return q
}

// WithSelect sets the fields to select
func (q *QueryParams) WithSelect(fields ...string) *QueryParams {
	q.Select = fields
	return q
}
