package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
)

// TestServer wraps an httptest.Server for integration testing
type TestServer struct {
	server *httptest.Server
}

// NewTestServer creates a real HTTP test server
func NewTestServer(handler http.Handler) *TestServer {
	return &TestServer{server: httptest.NewServer(handler)}
}

// Close shuts down the test server
func (s *TestServer) Close() {
	s.server.Close()
}

// Request represents a test HTTP request
type Request struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
	Token   string
}

// Response represents a test HTTP response
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// Do executes a test request against the real HTTP server
func (s *TestServer) Do(req Request) (*Response, error) {
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	url := s.server.URL + req.Path
	httpReq, err := http.NewRequest(req.Method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	if req.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.Token)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}

// GET performs a GET request
func (s *TestServer) GET(path string, token string) (*Response, error) {
	return s.Do(Request{
		Method: http.MethodGet,
		Path:   path,
		Token:  token,
	})
}

// POST performs a POST request
func (s *TestServer) POST(path string, body interface{}, token string) (*Response, error) {
	return s.Do(Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
		Token:  token,
	})
}

// PUT performs a PUT request
func (s *TestServer) PUT(path string, body interface{}, token string) (*Response, error) {
	return s.Do(Request{
		Method: http.MethodPut,
		Path:   path,
		Body:   body,
		Token:  token,
	})
}

// DELETE performs a DELETE request
func (s *TestServer) DELETE(path string, token string) (*Response, error) {
	return s.Do(Request{
		Method: http.MethodDelete,
		Path:   path,
		Token:  token,
	})
}

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success   bool            `json:"success"`
	Message   string          `json:"message"`
	ErrorCode string          `json:"error_code,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Errors    json.RawMessage `json:"errors,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
}

// ParseResponse parses the response body into APIResponse
func (r *Response) ParseResponse() (*APIResponse, error) {
	var resp APIResponse
	err := json.Unmarshal(r.Body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ParseData parses the response data into a specific type
func (r *Response) ParseData(v interface{}) error {
	var resp struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(r.Body, &resp); err != nil {
		return err
	}
	return json.Unmarshal(resp.Data, v)
}

// ParseList parses the response data as a list
func (r *Response) ParseList(key string) ([]map[string]interface{}, error) {
	resp, err := r.ParseResponse()
	if err != nil {
		return nil, err
	}
	var dataMap map[string]interface{}
	if err := json.Unmarshal(resp.Data, &dataMap); err != nil {
		return nil, err
	}
	items, ok := dataMap[key].([]interface{})
	if !ok {
		return nil, nil
	}
	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = item.(map[string]interface{})
	}
	return result, nil
}
