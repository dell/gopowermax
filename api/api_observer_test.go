/*
 Copyright © 2026 Dell Inc. or its subsidiaries. All Rights Reserved.

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
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockRequestObserver is a mock implementation of RequestObserver for testing
type mockRequestObserver struct {
	mu           sync.Mutex
	observations []RequestObservation
}

func (m *mockRequestObserver) ObservePowerMaxRequest(obs RequestObservation) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.observations = append(m.observations, obs)
}

func (m *mockRequestObserver) getObservations() []RequestObservation {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]RequestObservation(nil), m.observations...)
}

func TestSetRequestObserver(t *testing.T) {
	c := &client{
		http:              &http.Client{},
		host:              "https://example.com",
		customHTTPHeaders: NewSafeHeader(),
	}

	observer := &mockRequestObserver{}
	c.SetRequestObserver(observer)

	assert.NotNil(t, c.requestObserver)
	assert.Equal(t, observer, c.requestObserver)
}

func TestObserveRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		endpoint       string
		statusCode     int
		err            error
		duration       time.Duration
		hasObserver    bool
		expectObserved bool
	}{
		{
			name:           "Successful request with observer",
			method:         "GET",
			endpoint:       "/api/symmetrix/000123456789",
			statusCode:     200,
			err:            nil,
			duration:       100 * time.Millisecond,
			hasObserver:    true,
			expectObserved: true,
		},
		{
			name:           "Failed request with observer",
			method:         "POST",
			endpoint:       "/api/volume",
			statusCode:     500,
			err:            assert.AnError,
			duration:       50 * time.Millisecond,
			hasObserver:    true,
			expectObserved: true,
		},
		{
			name:           "Request without observer",
			method:         "GET",
			endpoint:       "/api/test",
			statusCode:     200,
			err:            nil,
			duration:       100 * time.Millisecond,
			hasObserver:    false,
			expectObserved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				http:              &http.Client{},
				host:              "https://example.com",
				customHTTPHeaders: NewSafeHeader(),
			}

			var observer *mockRequestObserver
			if tt.hasObserver {
				observer = &mockRequestObserver{}
				c.SetRequestObserver(observer)
			}

			c.observeRequest(tt.method, tt.endpoint, tt.statusCode, tt.err, tt.duration)

			if tt.expectObserved {
				assert.NotNil(t, observer)
				observations := observer.getObservations()
				assert.Len(t, observations, 1)
				obs := observations[0]
				assert.Equal(t, tt.method, obs.Method)
				assert.Equal(t, tt.endpoint, obs.Endpoint)
				assert.Equal(t, tt.statusCode, obs.StatusCode)
				assert.Equal(t, tt.err, obs.Err)
				assert.Equal(t, tt.duration, obs.Duration)
			} else {
				assert.Nil(t, observer)
			}
		})
	}
}

func TestDoAndGetResponseBodyWithObserver(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		uri                string
		responseStatusCode int
		responseBody       string
		expectError        bool
		expectObservation  bool
	}{
		{
			name:               "Successful GET request",
			method:             http.MethodGet,
			uri:                "/api/symmetrix/000123456789",
			responseStatusCode: http.StatusOK,
			responseBody:       `{"success": true}`,
			expectError:        false,
			expectObservation:  true,
		},
		{
			name:               "Failed POST request",
			method:             http.MethodPost,
			uri:                "/api/volume",
			responseStatusCode: http.StatusInternalServerError,
			responseBody:       `{"error": "internal error"}`,
			expectError:        false,
			expectObservation:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.responseStatusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			// Create client with observer
			c, err := New(ts.URL, ClientOptions{Timeout: 10 * time.Second}, false)
			assert.NoError(t, err)

			observer := &mockRequestObserver{}
			c.SetRequestObserver(observer)

			// Make request
			_, err = c.DoAndGetResponseBody(context.Background(), tt.method, tt.uri, nil, nil)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify observation
			if tt.expectObservation {
				observations := observer.getObservations()
				assert.Len(t, observations, 1)
				obs := observations[0]
				assert.Equal(t, tt.method, obs.Method)
				assert.Contains(t, obs.Endpoint, tt.uri)
				assert.Equal(t, tt.responseStatusCode, obs.StatusCode)
				assert.Greater(t, obs.Duration, time.Duration(0))
			}
		})
	}
}

func TestDoAndGetResponseBodyWithObserverError(t *testing.T) {
	// Create a client that will fail to connect
	c, err := New("http://localhost:1", ClientOptions{Timeout: 1 * time.Millisecond}, false)
	assert.NoError(t, err)

	observer := &mockRequestObserver{}
	c.SetRequestObserver(observer)

	// Make request that will fail
	_, err = c.DoAndGetResponseBody(context.Background(), http.MethodGet, "/test", nil, nil)
	assert.Error(t, err)

	// Verify observation captured the error
	observations := observer.getObservations()
	assert.Len(t, observations, 1)
	obs := observations[0]
	assert.Equal(t, http.MethodGet, obs.Method)
	assert.Equal(t, 0, obs.StatusCode)
	assert.NotNil(t, obs.Err)
	assert.Greater(t, obs.Duration, time.Duration(0))
}

func TestConcurrentObservations(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, ClientOptions{Timeout: 10 * time.Second}, false)
	assert.NoError(t, err)

	observer := &mockRequestObserver{}
	c.SetRequestObserver(observer)

	// Make concurrent requests
	const numRequests = 10
	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			_, _ = c.DoAndGetResponseBody(context.Background(), http.MethodGet, "/test", nil, nil)
		}()
	}

	wg.Wait()

	// Verify all observations were captured
	observations := observer.getObservations()
	assert.Len(t, observations, numRequests)
}
