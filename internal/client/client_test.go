package client

import (
	"encoding/json"
	"testing"
)

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name     string
		id       int64
		method   string
		params   interface{}
		wantJSON string
	}{
		{
			name:     "simple request without params",
			id:       1,
			method:   "pool.query",
			params:   nil,
			wantJSON: `{"jsonrpc":"2.0","method":"pool.query","id":1}`,
		},
		{
			name:   "request with array params",
			id:     2,
			method: "auth.login_with_api_key",
			params: []interface{}{"test-api-key"},
			wantJSON: `{"jsonrpc":"2.0","method":"auth.login_with_api_key","params":["test-api-key"],"id":2}`,
		},
		{
			name:   "request with map params",
			id:     3,
			method: "pool.dataset.create",
			params: map[string]interface{}{
				"name": "tank/test",
				"type": "FILESYSTEM",
			},
			wantJSON: `{"jsonrpc":"2.0","method":"pool.dataset.create","params":{"name":"tank/test","type":"FILESYSTEM"},"id":3}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest(tt.id, tt.method, tt.params)

			if req.JSONRPC != "2.0" {
				t.Errorf("NewRequest().JSONRPC = %v, want 2.0", req.JSONRPC)
			}
			if req.Method != tt.method {
				t.Errorf("NewRequest().Method = %v, want %v", req.Method, tt.method)
			}
			if req.ID != tt.id {
				t.Errorf("NewRequest().ID = %v, want %v", req.ID, tt.id)
			}

			// Verify JSON serialization
			data, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			// For comparison, unmarshal both and compare
			var got, want map[string]interface{}
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Failed to unmarshal got: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.wantJSON), &want); err != nil {
				t.Fatalf("Failed to unmarshal want: %v", err)
			}

			// Compare key fields
			if got["jsonrpc"] != want["jsonrpc"] {
				t.Errorf("jsonrpc mismatch: got %v, want %v", got["jsonrpc"], want["jsonrpc"])
			}
			if got["method"] != want["method"] {
				t.Errorf("method mismatch: got %v, want %v", got["method"], want["method"])
			}
			if got["id"] != want["id"] {
				t.Errorf("id mismatch: got %v, want %v", got["id"], want["id"])
			}
		})
	}
}

func TestQueryParams(t *testing.T) {
	t.Run("empty query params", func(t *testing.T) {
		params := NewQueryParams()
		if params.Limit != 0 {
			t.Errorf("NewQueryParams().Limit = %v, want 0", params.Limit)
		}
		if params.Offset != 0 {
			t.Errorf("NewQueryParams().Offset = %v, want 0", params.Offset)
		}
		if len(params.Filters) != 0 {
			t.Errorf("NewQueryParams().Filters length = %v, want 0", len(params.Filters))
		}
	})

	t.Run("with filter", func(t *testing.T) {
		params := NewQueryParams().WithFilter("name", "=", "test")
		if len(params.Filters) != 1 {
			t.Fatalf("Expected 1 filter, got %d", len(params.Filters))
		}
		filter := params.Filters[0]
		if filter[0] != "name" {
			t.Errorf("Filter field = %v, want name", filter[0])
		}
		if filter[1] != "=" {
			t.Errorf("Filter operator = %v, want =", filter[1])
		}
		if filter[2] != "test" {
			t.Errorf("Filter value = %v, want test", filter[2])
		}
	})

	t.Run("with limit and offset", func(t *testing.T) {
		params := NewQueryParams().WithLimit(10).WithOffset(20)
		if params.Limit != 10 {
			t.Errorf("Limit = %v, want 10", params.Limit)
		}
		if params.Offset != 20 {
			t.Errorf("Offset = %v, want 20", params.Offset)
		}
	})

	t.Run("with select", func(t *testing.T) {
		params := NewQueryParams().WithSelect("id", "name", "status")
		if len(params.Select) != 3 {
			t.Fatalf("Expected 3 select fields, got %d", len(params.Select))
		}
		expected := []string{"id", "name", "status"}
		for i, field := range expected {
			if params.Select[i] != field {
				t.Errorf("Select[%d] = %v, want %v", i, params.Select[i], field)
			}
		}
	})

	t.Run("chained operations", func(t *testing.T) {
		params := NewQueryParams().
			WithFilter("pool", "=", "tank").
			WithFilter("type", "=", "FILESYSTEM").
			WithLimit(5).
			WithOffset(0).
			WithSelect("name", "mountpoint")

		if len(params.Filters) != 2 {
			t.Errorf("Expected 2 filters, got %d", len(params.Filters))
		}
		if params.Limit != 5 {
			t.Errorf("Limit = %v, want 5", params.Limit)
		}
		if len(params.Select) != 2 {
			t.Errorf("Expected 2 select fields, got %d", len(params.Select))
		}
	})
}

func TestNewClientConfig(t *testing.T) {
	cfg := &Config{
		Host:      "truenas.local",
		APIKey:    "test-key",
		VerifySSL: true,
	}

	client := NewClient(cfg)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.host != cfg.Host {
		t.Errorf("client.host = %v, want %v", client.host, cfg.Host)
	}
	if client.apiKey != cfg.APIKey {
		t.Errorf("client.apiKey = %v, want %v", client.apiKey, cfg.APIKey)
	}
	if client.verifySSL != cfg.VerifySSL {
		t.Errorf("client.verifySSL = %v, want %v", client.verifySSL, cfg.VerifySSL)
	}
	if client.responses == nil {
		t.Error("client.responses map is nil")
	}
}
