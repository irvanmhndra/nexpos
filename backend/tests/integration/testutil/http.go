package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v5"
)

// TestServer wraps an Echo instance for testing
type TestServer struct {
	Echo *echo.Echo
}

// NewTestServer creates a test server
func NewTestServer(e *echo.Echo) *TestServer {
	return &TestServer{Echo: e}
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

// Do executes a test request
func (s *TestServer) Do(req Request) (*Response, error) {
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	httpReq := httptest.NewRequest(req.Method, req.Path, bodyReader)
	httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	// Set custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set auth token if provided
	if req.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.Token)
	}

	rec := httptest.NewRecorder()
	s.Echo.ServeHTTP(rec, httpReq)

	return &Response{
		StatusCode: rec.Code,
		Body:       rec.Body.Bytes(),
		Headers:    rec.Header(),
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
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	ErrorCode string                 `json:"error_code,omitempty"`
	Errors    interface{}            `json:"errors,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
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
	var resp map[string]interface{}
	if err := json.Unmarshal(r.Body, &resp); err != nil {
		return nil, err
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	items, ok := data[key].([]interface{})
	if !ok {
		return nil, nil
	}

	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = item.(map[string]interface{})
	}
	return result, nil
}
