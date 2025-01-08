/*
 Copyright © 2021-2025 Dell Inc. or its subsidiaries. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"

	types "github.com/dell/gopowermax/v2/types/v100"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type stubTypeWithMetaData struct{}

func (s stubTypeWithMetaData) MetaData() http.Header {
	h := make(http.Header)
	h.Set("foo", "bar")
	return h
}

func Test_addMetaData(t *testing.T) {
	tests := []struct {
		name           string
		givenHeader    map[string]string
		expectedHeader map[string]string
		body           interface{}
	}{
		{"nil header is a noop", nil, nil, nil},
		{"nil body is a noop", nil, nil, nil},
		{"header is updated", make(map[string]string), map[string]string{"Foo": "bar"}, stubTypeWithMetaData{}},
		{"header is not updated", make(map[string]string), map[string]string{}, struct{}{}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			addMetaData(tt.givenHeader, tt.body)

			switch {
			case tt.givenHeader == nil:
				if tt.givenHeader != nil {
					t.Errorf("(%s): expected %s, actual %s", tt.body, tt.expectedHeader, tt.givenHeader)
				}
			case tt.body == nil:
				if len(tt.givenHeader) != 0 {
					t.Errorf("(%s): expected %s, actual %s", tt.body, tt.expectedHeader, tt.givenHeader)
				}
			default:
				if !reflect.DeepEqual(tt.expectedHeader, tt.givenHeader) {
					t.Errorf("(%s): expected %s, actual %s", tt.body, tt.expectedHeader, tt.givenHeader)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		opts        ClientOptions
		debug       bool
		expectError bool
	}{
		{
			name:        "Valid host without options",
			host:        "http://example.com",
			opts:        ClientOptions{},
			debug:       false,
			expectError: false,
		},
		{
			name:        "Empty host",
			host:        "",
			opts:        ClientOptions{},
			debug:       false,
			expectError: true,
		},
		{
			name: "Valid host with timeout",
			host: "http://example.com",
			opts: ClientOptions{
				Timeout: 10 * time.Second,
			},
			debug:       false,
			expectError: false,
		},
		{
			name: "Valid host with insecure option",
			host: "http://example.com",
			opts: ClientOptions{
				Insecure: true,
			},
			debug:       false,
			expectError: false,
		},
		{
			name: "Valid host with dummy cert file",
			host: "http://example.com",
			opts: ClientOptions{
				CertFile: "../mock/cert.pem",
			},
			debug:       false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.host, tt.opts, tt.debug)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected nil error, but got: %v", err)
				}
				if c == nil {
					t.Error("Expected non-nil client, but got nil")
				}
			}
		})
	}
}

func (m *MockClient) GetHTTPClient() *http.Client {
	return m.http
}

func (m *MockClient) DoWithHeaders(
	ctx context.Context,
	method, path string,
	headers map[string]string,
	body, resp interface{},
) error {
	args := m.Called(ctx, method, path, headers, body, resp)
	return args.Error(0)
}

func TestGetHTTPClient(t *testing.T) {
	mockClient := &MockClient{http: &http.Client{}}
	client := &client{http: mockClient.GetHTTPClient()}

	assert.Equal(t, mockClient.GetHTTPClient(), client.GetHTTPClient())
}

func TestBeginsWithSlash(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"/path", true},
		{"path/", false},
		{"/", true},
	}

	for _, test := range tests {
		result := beginsWithSlash(test.input)
		if result != test.expected {
			t.Errorf("beginsWithSlash(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestEndsWithSlash(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"/path/", true},
		{"path/", true},
		{"/", true},
		{"path", false},
	}

	for _, test := range tests {
		result := endsWithSlash(test.input)
		if result != test.expected {
			t.Errorf("endsWithSlash(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

type MockClient struct {
	*client
	http *http.Client
	mock.Mock
}

func (m *MockClient) DoAndGetResponseBody(
	ctx context.Context,
	method, uri string,
	headers map[string]string,
	body interface{},
) (*http.Response, error) {
	args := m.Called(ctx, method, uri, headers, body)
	resp := args.Get(0)
	if resp == nil {
		return nil, args.Error(1)
	}
	return resp.(*http.Response), args.Error(1)
}

func TestDoWithHeaders(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	httpClient := &http.Client{
		Transport: &MockTransport{mockHTTPClient: mockHTTPClient},
	}
	mockClient := &client{
		http:     httpClient,
		host:     "https://example.com",
		token:    "mockToken",
		showHTTP: false,
	}

	tests := []struct {
		name          string
		method        string
		uri           string
		headers       map[string]string
		body          interface{}
		resp          interface{}
		mockResponse  *http.Response
		mockDoError   error
		expectedError string
	}{
		{
			name:   "Successful GET request with response",
			method: http.MethodGet,
			uri:    "/test",
			headers: map[string]string{
				"Custom-Header": "value",
			},
			body: nil,
			resp: &map[string]interface{}{},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"success": true}`)),
			},
			mockDoError:   nil,
			expectedError: "",
		},
		{
			name:          "Nil response",
			method:        http.MethodGet,
			uri:           "/test",
			headers:       nil,
			body:          nil,
			resp:          nil,
			mockResponse:  nil,
			mockDoError:   nil,
			expectedError: "",
		},
		{
			name:   "Failed to decode response body",
			method: http.MethodGet,
			uri:    "/test",
			headers: map[string]string{
				"Custom-Header": "value",
			},
			body: nil,
			resp: &map[string]interface{}{},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"invalid_json":`)),
			},
			mockDoError:   nil,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPClient.On("Do", mock.Anything).Return(tt.mockResponse, tt.mockDoError)

			err := mockClient.DoWithHeaders(
				context.TODO(),
				tt.method,
				tt.uri,
				tt.headers,
				tt.body,
				tt.resp,
			)

			if tt.expectedError != "" {
				if err == nil || err.Error() != tt.expectedError {
					t.Errorf("Expected error: %v, but got: %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
			}

			mockHTTPClient.AssertExpectations(t)
		})
	}
}

// MockHTTPClient is a mock http client
type MockHTTPClient struct {
	mock.Mock
	http.Client
}

// Do sends an HTTP request to the API
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	resp := args.Get(0)
	if resp == nil {
		return nil, args.Error(1)
	}
	return resp.(*http.Response), args.Error(1)
}

// MockTransport is a mock http transport
type MockTransport struct {
	mockHTTPClient *MockHTTPClient
	Response       *http.Response
}

// RoundTrip sends an HTTP request to the API
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.mockHTTPClient.Do(req)
}

func TestDoAndGetResponseBody(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	httpClient := &http.Client{
		Transport: &MockTransport{mockHTTPClient: mockHTTPClient},
	}
	c := &client{
		http:     httpClient,
		host:     "https://example.com",
		token:    "mockToken",
		showHTTP: false,
	}

	tests := []struct {
		name          string
		method        string
		uri           string
		headers       map[string]string
		body          interface{}
		mockResponse  *http.Response
		mockError     error
		expectedError string
	}{
		{
			name:   "Successful GET request",
			method: http.MethodGet,
			uri:    "/test",
			headers: map[string]string{
				"Custom-Header": "value",
			},
			body: nil,
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"success": true}`)),
			},
			mockError:     nil,
			expectedError: "",
		},
		{
			name:   "Failed POST request with invalid JSON body",
			method: http.MethodPost,
			uri:    "/test",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			body:          make(chan int), // invalid JSON body
			mockResponse:  nil,
			mockError:     errors.New("unsupported type error"),
			expectedError: "json: unsupported type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPClient.On("Do", mock.Anything).Return(tt.mockResponse, tt.mockError)

			res, err := c.DoAndGetResponseBody(
				context.Background(),
				tt.method,
				tt.uri,
				tt.headers,
				tt.body,
			)

			if tt.expectedError != "" {
				if err == nil || err.Error() != tt.expectedError {
					t.Errorf("Expected error: %v, but got: %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				if !reflect.DeepEqual(res, tt.mockResponse) {
					t.Errorf("Expected response: %v, but got: %v", tt.mockResponse, res)
				}
			}

			mockHTTPClient.AssertExpectations(t)
		})
	}
}

func TestGetToken(t *testing.T) {
	// Test case: token is not set
	c := &client{}
	if c.GetToken() != "" {
		t.Errorf("GetToken() = %v, want %v", c.GetToken(), "")
	}

	// Test case: token is set
	c.token = "testToken"
	if c.GetToken() != "testToken" {
		t.Errorf("GetToken() = %v, want %v", c.GetToken(), "testToken")
	}
}

func TestSetToken(t *testing.T) {
	c := &client{}

	c.SetToken("token1")
	if c.token != "token1" {
		t.Errorf("Expected token to be 'token1', got '%s'", c.token)
	}

	c.SetToken("token2")
	if c.token != "token2" {
		t.Errorf("Expected token to be 'token2', got '%s'", c.token)
	}
}

func TestParseJSONError(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		statusCode    int
		expectedError *types.Error
	}{
		{
			name:         "Valid JSON error response",
			responseBody: `{"Message": "error occurred"}`,
			statusCode:   http.StatusBadRequest,
			expectedError: &types.Error{
				HTTPStatusCode: http.StatusBadRequest,
				Message:        "error occurred",
			},
		},
		{
			name:         "Invalid JSON error response",
			responseBody: `invalid json`,
			statusCode:   http.StatusInternalServerError,
			expectedError: &types.Error{
				HTTPStatusCode: http.StatusInternalServerError,
				Message:        http.StatusText(http.StatusInternalServerError),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.responseBody)),
			}

			c := &client{}
			err := c.ParseJSONError(r)

			assert.Error(t, err)
			if e, ok := err.(*types.Error); ok {
				assert.Equal(t, tt.expectedError.HTTPStatusCode, e.HTTPStatusCode)
				assert.Equal(t, tt.expectedError.Message, e.Message)
			}
		})
	}
}

func TestDoLog(t *testing.T) {
	type fields struct {
		debug bool
	}
	type args struct {
		l   func(args ...interface{})
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "debug is true",
			fields: fields{
				debug: true,
			},
			args: args{
				l:   log.Println,
				msg: "test message",
			},
		},
		{
			name: "debug is false",
			fields: fields{
				debug: false,
			},
			args: args{
				l:   log.Println,
				msg: "test message",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			c := &client{
				debug: tt.fields.debug,
			}
			c.doLog(tt.args.l, tt.args.msg)
		})
	}
}
