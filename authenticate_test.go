package pmax

import (
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
