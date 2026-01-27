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
