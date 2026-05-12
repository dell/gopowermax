/*
Copyright © 2025 Dell Inc. or its subsidiaries. All Rights Reserved.

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

package v100

import (
	"errors"
	"net/http"
	"testing"
)

func TestGetJobResource(t *testing.T) {
	tests := []struct {
		name                 string
		job                  Job
		expectedSymmetrixID  string
		expectedResourceType string
		expectedResourceID   string
	}{
		{
			name: "valid resource link",
			job: Job{
				ResourceLink: "provisioning/system/SYMMETRIX-1234/volume/1234",
			},
			expectedSymmetrixID:  "SYMMETRIX-1234",
			expectedResourceType: "volume",
			expectedResourceID:   "1234",
		},
		{
			name: "Empty resource link",
			job: Job{
				ResourceLink: "",
			},
			expectedSymmetrixID:  "",
			expectedResourceType: "",
			expectedResourceID:   "",
		},
		{
			name: "invalid resource link",
			job: Job{
				ResourceLink: "system/SYMMETRIX-1234",
			},
			expectedSymmetrixID:  "",
			expectedResourceType: "",
			expectedResourceID:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symID, resourceType, id := tt.job.GetJobResource()

			if symID != tt.expectedSymmetrixID {
				t.Errorf("expected %s, got %s", tt.expectedSymmetrixID, symID)
			}

			if resourceType != tt.expectedResourceType {
				t.Errorf("expected %s, got %s", tt.expectedResourceType, resourceType)
			}

			if id != tt.expectedResourceID {
				t.Errorf("expected %s, got %s", tt.expectedResourceID, id)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
	}{
		{
			name: "valid error",
			err: &Error{
				Message: "test-error",
			},
		},
		{
			name: "empty error",
			err: &Error{
				Message: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.err.Message {
				t.Errorf("expected %s, got %s", tt.err.Message, tt.err.Error())
			}
		})
	}
}

func TestError_HasHTTPStatus(t *testing.T) {
	e := Error{HTTPStatusCode: http.StatusNotFound}
	if !e.HasHTTPStatus(http.StatusNotFound) {
		t.Error("expected HasHTTPStatus(404) to be true")
	}
	if e.HasHTTPStatus(http.StatusOK) {
		t.Error("expected HasHTTPStatus(200) to be false")
	}
}

func TestError_IsNotFound(t *testing.T) {
	nf := Error{HTTPStatusCode: http.StatusNotFound}
	if !nf.IsNotFound() {
		t.Error("expected IsNotFound to be true for 404")
	}
	other := Error{HTTPStatusCode: http.StatusInternalServerError}
	if other.IsNotFound() {
		t.Error("expected IsNotFound to be false for 500")
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: 0,
		},
		{
			name:     "non-API error",
			err:      errors.New("plain error"),
			expected: 0,
		},
		{
			name:     "404 API error",
			err:      &Error{HTTPStatusCode: http.StatusNotFound, Message: "not found"},
			expected: http.StatusNotFound,
		},
		{
			name:     "500 API error",
			err:      &Error{HTTPStatusCode: http.StatusInternalServerError, Message: "server error"},
			expected: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHTTPStatus(tt.err); got != tt.expected {
				t.Errorf("GetHTTPStatus() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestHasHTTPStatus_Package(t *testing.T) {
	apiErr := &Error{HTTPStatusCode: http.StatusConflict}
	if !HasHTTPStatus(apiErr, http.StatusConflict) {
		t.Error("expected HasHTTPStatus to be true for matching code")
	}
	if HasHTTPStatus(apiErr, http.StatusOK) {
		t.Error("expected HasHTTPStatus to be false for non-matching code")
	}
	if HasHTTPStatus(errors.New("plain"), http.StatusConflict) {
		t.Error("expected HasHTTPStatus to be false for non-API error")
	}
	if HasHTTPStatus(nil, http.StatusConflict) {
		t.Error("expected HasHTTPStatus to be false for nil error")
	}
}

func TestIsNotFoundError(t *testing.T) {
	if !IsNotFoundError(&Error{HTTPStatusCode: http.StatusNotFound}) {
		t.Error("expected true for 404 API error")
	}
	if IsNotFoundError(&Error{HTTPStatusCode: http.StatusBadRequest}) {
		t.Error("expected false for 400 API error")
	}
	if IsNotFoundError(errors.New("not found")) {
		t.Error("expected false for plain error")
	}
	if IsNotFoundError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestError_IsBadRequest(t *testing.T) {
	if !(&Error{HTTPStatusCode: http.StatusBadRequest}).IsBadRequest() {
		t.Error("expected IsBadRequest to be true for 400")
	}
	if (&Error{HTTPStatusCode: http.StatusOK}).IsBadRequest() {
		t.Error("expected IsBadRequest to be false for 200")
	}
}

func TestError_IsUnauthorized(t *testing.T) {
	if !(&Error{HTTPStatusCode: http.StatusUnauthorized}).IsUnauthorized() {
		t.Error("expected IsUnauthorized to be true for 401")
	}
	if (&Error{HTTPStatusCode: http.StatusForbidden}).IsUnauthorized() {
		t.Error("expected IsUnauthorized to be false for 403")
	}
}

func TestError_IsForbidden(t *testing.T) {
	if !(&Error{HTTPStatusCode: http.StatusForbidden}).IsForbidden() {
		t.Error("expected IsForbidden to be true for 403")
	}
	if (&Error{HTTPStatusCode: http.StatusNotFound}).IsForbidden() {
		t.Error("expected IsForbidden to be false for 404")
	}
}

func TestError_IsConflict(t *testing.T) {
	if !(&Error{HTTPStatusCode: http.StatusConflict}).IsConflict() {
		t.Error("expected IsConflict to be true for 409")
	}
	if (&Error{HTTPStatusCode: http.StatusBadRequest}).IsConflict() {
		t.Error("expected IsConflict to be false for 400")
	}
}

func TestError_IsUnprocessableEntity(t *testing.T) {
	if !(&Error{HTTPStatusCode: http.StatusUnprocessableEntity}).IsUnprocessableEntity() {
		t.Error("expected IsUnprocessableEntity to be true for 422")
	}
	if (&Error{HTTPStatusCode: http.StatusBadRequest}).IsUnprocessableEntity() {
		t.Error("expected IsUnprocessableEntity to be false for 400")
	}
}

func TestError_IsInternalServerError(t *testing.T) {
	if !(&Error{HTTPStatusCode: http.StatusInternalServerError}).IsInternalServerError() {
		t.Error("expected IsInternalServerError to be true for 500")
	}
	if (&Error{HTTPStatusCode: http.StatusBadRequest}).IsInternalServerError() {
		t.Error("expected IsInternalServerError to be false for 400")
	}
}

func TestError_IsServiceUnavailable(t *testing.T) {
	if !(&Error{HTTPStatusCode: http.StatusServiceUnavailable}).IsServiceUnavailable() {
		t.Error("expected IsServiceUnavailable to be true for 503")
	}
	if (&Error{HTTPStatusCode: http.StatusInternalServerError}).IsServiceUnavailable() {
		t.Error("expected IsServiceUnavailable to be false for 500")
	}
}

func TestError_IsClientError(t *testing.T) {
	if !(&Error{HTTPStatusCode: http.StatusBadRequest}).IsClientError() {
		t.Error("expected IsClientError to be true for 400")
	}
	if !(&Error{HTTPStatusCode: http.StatusNotFound}).IsClientError() {
		t.Error("expected IsClientError to be true for 404")
	}
	if !(&Error{HTTPStatusCode: http.StatusConflict}).IsClientError() {
		t.Error("expected IsClientError to be true for 409")
	}
	if !(&Error{HTTPStatusCode: http.StatusUnprocessableEntity}).IsClientError() {
		t.Error("expected IsClientError to be true for 422")
	}
	if (&Error{HTTPStatusCode: http.StatusOK}).IsClientError() {
		t.Error("expected IsClientError to be false for 200")
	}
	if (&Error{HTTPStatusCode: http.StatusInternalServerError}).IsClientError() {
		t.Error("expected IsClientError to be false for 500")
	}
}

func TestError_IsServerError(t *testing.T) {
	if !(&Error{HTTPStatusCode: http.StatusInternalServerError}).IsServerError() {
		t.Error("expected IsServerError to be true for 500")
	}
	if !(&Error{HTTPStatusCode: http.StatusServiceUnavailable}).IsServerError() {
		t.Error("expected IsServerError to be true for 503")
	}
	if !(&Error{HTTPStatusCode: http.StatusBadGateway}).IsServerError() {
		t.Error("expected IsServerError to be true for 502")
	}
	if !(&Error{HTTPStatusCode: http.StatusGatewayTimeout}).IsServerError() {
		t.Error("expected IsServerError to be true for 504")
	}
	if (&Error{HTTPStatusCode: http.StatusBadRequest}).IsServerError() {
		t.Error("expected IsServerError to be false for 400")
	}
	if (&Error{HTTPStatusCode: http.StatusOK}).IsServerError() {
		t.Error("expected IsServerError to be false for 200")
	}
}

func TestIsBadRequestError(t *testing.T) {
	if !IsBadRequestError(&Error{HTTPStatusCode: http.StatusBadRequest}) {
		t.Error("expected true for 400 API error")
	}
	if IsBadRequestError(&Error{HTTPStatusCode: http.StatusNotFound}) {
		t.Error("expected false for 404 API error")
	}
	if IsBadRequestError(errors.New("bad request")) {
		t.Error("expected false for plain error")
	}
	if IsBadRequestError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsUnauthorizedError(t *testing.T) {
	if !IsUnauthorizedError(&Error{HTTPStatusCode: http.StatusUnauthorized}) {
		t.Error("expected true for 401 API error")
	}
	if IsUnauthorizedError(&Error{HTTPStatusCode: http.StatusForbidden}) {
		t.Error("expected false for 403 API error")
	}
}

func TestIsForbiddenError(t *testing.T) {
	if !IsForbiddenError(&Error{HTTPStatusCode: http.StatusForbidden}) {
		t.Error("expected true for 403 API error")
	}
	if IsForbiddenError(&Error{HTTPStatusCode: http.StatusUnauthorized}) {
		t.Error("expected false for 401 API error")
	}
}

func TestIsConflictError(t *testing.T) {
	if !IsConflictError(&Error{HTTPStatusCode: http.StatusConflict}) {
		t.Error("expected true for 409 API error")
	}
	if IsConflictError(&Error{HTTPStatusCode: http.StatusBadRequest}) {
		t.Error("expected false for 400 API error")
	}
}

func TestIsUnprocessableEntityError(t *testing.T) {
	if !IsUnprocessableEntityError(&Error{HTTPStatusCode: http.StatusUnprocessableEntity}) {
		t.Error("expected true for 422 API error")
	}
	if IsUnprocessableEntityError(&Error{HTTPStatusCode: http.StatusBadRequest}) {
		t.Error("expected false for 400 API error")
	}
}

func TestIsInternalServerError(t *testing.T) {
	if !IsInternalServerError(&Error{HTTPStatusCode: http.StatusInternalServerError}) {
		t.Error("expected true for 500 API error")
	}
	if IsInternalServerError(&Error{HTTPStatusCode: http.StatusBadRequest}) {
		t.Error("expected false for 400 API error")
	}
}

func TestIsServiceUnavailableError(t *testing.T) {
	if !IsServiceUnavailableError(&Error{HTTPStatusCode: http.StatusServiceUnavailable}) {
		t.Error("expected true for 503 API error")
	}
	if IsServiceUnavailableError(&Error{HTTPStatusCode: http.StatusInternalServerError}) {
		t.Error("expected false for 500 API error")
	}
}

func TestIsClientError_Package(t *testing.T) {
	if !IsClientError(&Error{HTTPStatusCode: http.StatusBadRequest}) {
		t.Error("expected true for 400 API error")
	}
	if !IsClientError(&Error{HTTPStatusCode: http.StatusNotFound}) {
		t.Error("expected true for 404 API error")
	}
	if !IsClientError(&Error{HTTPStatusCode: http.StatusConflict}) {
		t.Error("expected true for 409 API error")
	}
	if IsClientError(&Error{HTTPStatusCode: http.StatusOK}) {
		t.Error("expected false for 200 API error")
	}
	if IsClientError(&Error{HTTPStatusCode: http.StatusInternalServerError}) {
		t.Error("expected false for 500 API error")
	}
	if IsClientError(errors.New("plain")) {
		t.Error("expected false for plain error")
	}
	if IsClientError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsServerError_Package(t *testing.T) {
	if !IsServerError(&Error{HTTPStatusCode: http.StatusInternalServerError}) {
		t.Error("expected true for 500 API error")
	}
	if !IsServerError(&Error{HTTPStatusCode: http.StatusServiceUnavailable}) {
		t.Error("expected true for 503 API error")
	}
	if IsServerError(&Error{HTTPStatusCode: http.StatusBadRequest}) {
		t.Error("expected false for 400 API error")
	}
	if IsServerError(&Error{HTTPStatusCode: http.StatusOK}) {
		t.Error("expected false for 200 API error")
	}
	if IsServerError(errors.New("plain")) {
		t.Error("expected false for plain error")
	}
	if IsServerError(nil) {
		t.Error("expected false for nil error")
	}
}
