package pmax

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetVersionDetails(t *testing.T) {
	// Create a mock server with error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	// Use the mock server's URL in the client
	client, err := NewClientWithArgs(srv.URL, "", true, false, "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetVersionDetails(context.Background())
	if err == nil {
		t.Errorf("Expected error but got none")
	} else {
		fmt.Printf("Expected error received: %s\n", err.Error())
	}

	// when body decode error occurs.
	srv2 := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
	}))
	defer srv2.Close()
	// Use the mock server's URL in the client
	client, err = NewClientWithArgs(srv2.URL, "", true, false, "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetVersionDetails(context.Background())
	if err == nil {
		t.Errorf("Expected error but got none")
	} else {
		fmt.Printf("Expected error received: %s\n", err.Error())
	}
}

func TestGetSymmetrixByID(t *testing.T) {
	// Test success case with new microcode fields
	mockJSON := `{
		"symmetrixId": "000197900046",
		"dell_service_tag": "service-tag-46",
		"device_count": 1045,
		"ucode": "5978.221.221",
		"model": "PowerMax_2500",
		"local": true,
		"all_flash": true,
		"disk_count": 8,
		"cache_size_mb": 203776,
		"data_encryption": "Disabled",
		"microcode": "6079.325.0",
		"microcode_date": "09-30-2025",
		"microcode_registered_build": 84,
		"microcode_package_version": "10.3.0.0 (Release 01, Build 6079_325/0084, 2025-09-30 14:44:22)"
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockJSON))
	}))
	defer srv.Close()

	client, err := NewClientWithArgs(srv.URL, "", true, false, "")
	if err != nil {
		t.Fatal(err)
	}

	sym, err := client.GetSymmetrixByID(context.Background(), "000197900046")
	if err != nil {
		t.Fatalf("Expected no error, got: %s", err.Error())
	}
	if sym.SymmetrixID != "000197900046" {
		t.Errorf("Expected SymmetrixID 000197900046, got %s", sym.SymmetrixID)
	}
	if sym.DellServiceTag != "service-tag-46" {
		t.Errorf("Expected DellServiceTag service-tag-46, got %s", sym.DellServiceTag)
	}
	if sym.Model != "PowerMax_2500" {
		t.Errorf("Expected Model PowerMax_2500, got %s", sym.Model)
	}
	if !sym.Local {
		t.Errorf("Expected Local true, got false")
	}
	if sym.Microcode != "6079.325.0" {
		t.Errorf("Expected Microcode 6079.325.0, got %s", sym.Microcode)
	}
	if sym.MicrocodeDate != "09-30-2025" {
		t.Errorf("Expected MicrocodeDate 09-30-2025, got %s", sym.MicrocodeDate)
	}
	if sym.MicrocodeRegisteredBuild != 84 {
		t.Errorf("Expected MicrocodeRegisteredBuild 84, got %d", sym.MicrocodeRegisteredBuild)
	}
	if sym.MicrocodePackageVersion != "10.3.0.0 (Release 01, Build 6079_325/0084, 2025-09-30 14:44:22)" {
		t.Errorf("Expected MicrocodePackageVersion, got %s", sym.MicrocodePackageVersion)
	}

	// Test HTTP error case
	srvErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}))
	defer srvErr.Close()

	client, err = NewClientWithArgs(srvErr.URL, "", true, false, "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetSymmetrixByID(context.Background(), "000197900046")
	if err == nil {
		t.Errorf("Expected error but got none")
	} else {
		fmt.Printf("Expected error received: %s\n", err.Error())
	}

	// Test decode error case (empty body)
	srvEmpty := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
	}))
	defer srvEmpty.Close()

	client, err = NewClientWithArgs(srvEmpty.URL, "", true, false, "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetSymmetrixByID(context.Background(), "000197900046")
	if err == nil {
		t.Errorf("Expected error but got none")
	} else {
		fmt.Printf("Expected error received: %s\n", err.Error())
	}
}

func TestGetPorts(t *testing.T) {
	// Create a mock server with error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	// Use the mock server's URL in the client
	client, err := NewClientWithArgs(srv.URL, "", true, false, "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetPorts(context.Background(), "000197900046")
	if err == nil {
		t.Errorf("Expected error but got none")
	} else {
		fmt.Printf("Expected error received: %s\n", err.Error())
	}
}

func TestGetISCSIEndpoints(t *testing.T) {
	tests := []struct {
		name          string
		directorsResp string
		endpointsResp string
		portDetails   string
		expectedError bool
		expectedCount int
	}{
		{
			name:          "Successful retrieval with valid endpoints",
			directorsResp: `["dir-1", "dir-2"]`,
			endpointsResp: `{
				"symmetrixPortKey": [
					{"directorId": "dir-1", "portId": "port-1"},
					{"directorId": "dir-1", "portId": "port-2"}
				]
			}`,
			portDetails: `{
				"symmetrixPort": {
					"port_status": "ON",
					"identifier": "iqn.2020-01.com.example:test",
					"ip_addresses": ["10.0.0.1", "10.0.0.2"]
				}
			}`,
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:          "No directors found",
			directorsResp: `[]`,
			endpointsResp: `{"symmetrixPortKey": []}`,
			portDetails:   `{}`,
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:          "Directors error",
			directorsResp: `[]`,
			endpointsResp: `{"symmetrixPortKey": []}`,
			portDetails:   `{}`,
			expectedError: true,
			expectedCount: 0,
		},
		{
			name:          "No endpoints on directors",
			directorsResp: `["dir-1"]`,
			endpointsResp: `{"symmetrixPortKey": []}`,
			portDetails:   `{}`,
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:          "Endpoint without IP addresses filtered out",
			directorsResp: `["dir-1"]`,
			endpointsResp: `{
				"symmetrixPortKey": [
					{"directorId": "dir-1", "portId": "port-1"}
				]
			}`,
			portDetails: `{
				"symmetrixPort": {
					"port_status": "ON",
					"identifier": "iqn.2020-01.com.example:test",
					"ip_addresses": []
				}
			}`,
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:          "Endpoint without identifier filtered out",
			directorsResp: `["dir-1"]`,
			endpointsResp: `{
				"symmetrixPortKey": [
					{"directorId": "dir-1", "portId": "port-1"}
				]
			}`,
			portDetails: `{
				"symmetrixPort": {
					"port_status": "ON",
					"identifier": "",
					"ip_addresses": ["10.0.0.1"]
				}
			}`,
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:          "Port details error - function continues",
			directorsResp: `["dir-1"]`,
			endpointsResp: `{
				"symmetrixPortKey": [
					{"directorId": "dir-1", "portId": "port-1"}
				]
			}`,
			portDetails:   `{"error": "Port not found"}`,
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:          "Mixed success and failures",
			directorsResp: `["dir-1", "dir-2"]`,
			endpointsResp: `{
				"symmetrixPortKey": [
					{"directorId": "dir-1", "portId": "port-1"}
				]
			}`,
			portDetails: `{
				"symmetrixPort": {
					"port_status": "ON",
					"identifier": "iqn.2020-01.com.example:test",
					"ip_addresses": ["10.0.0.1"]
				}
			}`,
			expectedError: false,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Directors list request
				if strings.Contains(r.URL.Path, "/symmetrix/") && strings.Contains(r.URL.Path, "/director") && r.Method == "GET" && !strings.Contains(r.URL.Path, "/director/") {
					if tt.name == "Directors error" {
						w.WriteHeader(500)
					}
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprint(w, `{"directorId": `+tt.directorsResp+`}`)
				}

				// Port list request (any port request)
				if strings.Contains(r.URL.Path, "/director/") && strings.Contains(r.URL.Path, "/port") && !strings.Contains(r.URL.Path, "/port/") {
					// Check if this is an iSCSI endpoint query
					if strings.Contains(r.URL.RawQuery, "iscsi_endpoint=true") {
						// Only return endpoints for dir-1
						if strings.Contains(r.URL.Path, "/director/dir-1/port") {
							w.Header().Set("Content-Type", "application/json")
							fmt.Fprint(w, tt.endpointsResp)
						} else {
							// Empty response for other directors
							w.Header().Set("Content-Type", "application/json")
							fmt.Fprint(w, `{"symmetrixPortKey": []}`)
						}
					} else {
						// Default empty response for other port queries
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprint(w, `{"symmetrixPortKey": []}`)
					}
					return
				}

				// Port details request
				if strings.Contains(r.URL.Path, "/director/") && strings.Contains(r.URL.Path, "/port/") {
					// Extract director and port from URL
					parts := strings.Split(r.URL.Path, "/")
					var directorID, portID string
					for i, part := range parts {
						if part == "director" && i+1 < len(parts) {
							directorID = parts[i+1]
						}
						if part == "port" && i+1 < len(parts) {
							portID = parts[i+1]
						}
					}

					// Only respond to valid port requests
					if directorID != "" && portID != "" {
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprint(w, tt.portDetails)
					} else {
						w.WriteHeader(404)
					}
					return
				}

				// Default response
				w.WriteHeader(200)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{}`)
			}))
			defer srv.Close()

			// Create client
			client, err := NewClientWithArgs(srv.URL, "", true, false, "")
			if err != nil {
				t.Fatal(err)
			}

			// Call GetISCSIEndpoints
			targets, err := client.(*Client).GetISCSIEndpoints(context.Background(), "000197900046")

			// Check error expectation
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check target count
			if len(targets) != tt.expectedCount {
				t.Errorf("Expected %d targets, got %d", tt.expectedCount, len(targets))
			}

			// Verify target details if we expect any
			if len(targets) > 0 && tt.expectedCount > 0 {
				for _, target := range targets {
					if target.IQN == "" {
						t.Errorf("Expected non-empty IQN")
					}
					if len(target.PortalIPs) == 0 {
						t.Errorf("Expected non-empty portal IPs")
					}
					if target.PortStatus == "" {
						t.Errorf("Expected non-empty PortStatus")
					}
				}
			}
		})
	}
}
