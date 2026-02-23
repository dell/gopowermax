package pmax

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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
