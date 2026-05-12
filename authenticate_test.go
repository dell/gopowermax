package pmax

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithSymmetrixID(t *testing.T) {
	originalClient := &Client{}

	symID := "000123456789"
	newClient := originalClient.WithSymmetrixID(symID)

	// Type assertion to get the underlying *Client
	updatedClient, ok := newClient.(*Client)
	if !ok {
		t.Fatalf("Expected *Client, got %T", newClient)
	}

	if updatedClient.symmetrixID != symID {
		t.Errorf("Expected symmetrixID to be %s, got %s", symID, updatedClient.symmetrixID)
	}

	// Ensure original client is not modified
	if originalClient.symmetrixID != "" {
		t.Errorf("Expected original client symmetrixID to remain empty, got %s", originalClient.symmetrixID)
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name             string
		endpoint         string
		appName          string
		insecure         string
		useCerts         string
		expectedInsecure bool
		expectedUseCerts bool
		expectError      bool
	}{
		{
			name:             "Valid environment variables",
			endpoint:         "https://powermax.example.com",
			appName:          "CSIApp",
			insecure:         "true",
			useCerts:         "false",
			expectedInsecure: true,
			expectedUseCerts: false,
			expectError:      false,
		},
		{
			name:             "UseCerts enabled",
			endpoint:         "https://powermax.example.com",
			appName:          "CSIApp",
			insecure:         "false",
			useCerts:         "true",
			expectedInsecure: false,
			expectedUseCerts: true,
			expectError:      false,
		},
		{
			name:        "Missing endpoint",
			endpoint:    "",
			appName:     "CSIApp",
			insecure:    "true",
			useCerts:    "true",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("CSI_POWERMAX_ENDPOINT", tc.endpoint)
			os.Setenv("CSI_APPLICATION_NAME", tc.appName)
			os.Setenv("CSI_POWERMAX_INSECURE", tc.insecure)
			os.Setenv("CSI_POWERMAX_USECERTS", tc.useCerts)

			client, err := NewClient()

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				// Optionally, assert internal fields if accessible
			}

			// Clean up environment variables
			os.Unsetenv("CSI_POWERMAX_ENDPOINT")
			os.Unsetenv("CSI_APPLICATION_NAME")
			os.Unsetenv("CSI_POWERMAX_INSECURE")
			os.Unsetenv("CSI_POWERMAX_USECERTS")
		})
	}
}

func TestAuthenticate(t *testing.T) {
	tests := []struct {
		name            string
		serverBody      string
		serverStatus    int
		explicitVersion string
		wantErr         bool
		expectedVersion string
	}{
		{
			name:            "No explicit version - DefaultAPIVersion used",
			serverBody:      `{"version":"V10.4","api_version":"104"}`,
			serverStatus:    http.StatusOK,
			explicitVersion: "",
			wantErr:         false,
			expectedVersion: DefaultAPIVersion,
		},
		{
			name:            "Explicit version preserved",
			serverBody:      `{"version":"V10.4","api_version":"104"}`,
			serverStatus:    http.StatusOK,
			explicitVersion: "104",
			wantErr:         false,
			expectedVersion: "104",
		},
		{
			name:            "No APIVersion in response - version unchanged",
			serverBody:      `{"version":"V9.1"}`,
			serverStatus:    http.StatusOK,
			explicitVersion: "",
			wantErr:         false,
			expectedVersion: DefaultAPIVersion,
		},
		{
			name:         "HTTP error from server",
			serverBody:   `{"message":"Internal Server Error"}`,
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
		{
			name:         "Invalid JSON response causes decode error",
			serverBody:   `not-json{{{`,
			serverStatus: http.StatusOK,
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.serverStatus)
				w.Write([]byte(tc.serverBody))
			}))
			defer srv.Close()

			c, err := NewClientWithArgs(srv.URL, "", true, false, "")
			assert.NoError(t, err)

			client := c.(*Client)
			err = client.Authenticate(context.Background(), &ConfigConnect{
				Endpoint: srv.URL,
				Username: "testuser",
				Password: "testpass",
				Version:  tc.explicitVersion,
			})

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedVersion, client.version)
			}
		})
	}
}
