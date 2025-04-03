package linear

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	// Test default client
	apiKey := "test_api_key"
	client := NewClient(apiKey)

	if client.apiKey != apiKey {
		t.Errorf("Expected apiKey to be %s, got %s", apiKey, client.apiKey)
	}

	if client.apiURL != DefaultAPIURL {
		t.Errorf("Expected apiURL to be %s, got %s", DefaultAPIURL, client.apiURL)
	}

	if client.httpCli == nil {
		t.Error("Expected httpCli to be set")
	}

	// Test with custom options
	customURL := "https://custom.linear.app/graphql"
	customHTTPClient := &http.Client{}

	client = NewClient(apiKey, WithURL(customURL), WithHTTPClient(customHTTPClient))

	if client.apiURL != customURL {
		t.Errorf("Expected apiURL to be %s, got %s", customURL, client.apiURL)
	}

	if client.httpCli != customHTTPClient {
		t.Error("Expected httpCli to be the custom client")
	}
}

func TestMockResponses(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		if r.Header.Get("Authorization") != "test_api_key" {
			t.Errorf("Expected Authorization header to be test_api_key, got %s", r.Header.Get("Authorization"))
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Mock a successful GraphQL response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"viewer": {
					"id": "user123",
					"name": "Test User",
					"email": "test@example.com"
				}
			}
		}`))
	}))
	defer server.Close()

	// Create a client that uses the test server
	client := NewClient("test_api_key", WithURL(server.URL))

	// Test GetViewer
	user, err := client.GetViewer()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if user.ID != "user123" {
		t.Errorf("Expected user ID to be user123, got %s", user.ID)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected user name to be Test User, got %s", user.Name)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected user email to be test@example.com, got %s", user.Email)
	}
}

func TestExecuteGraphQLWithErrors(t *testing.T) {
	// Create a test server that returns GraphQL errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"errors": [
				{
					"message": "Test GraphQL error"
				}
			]
		}`))
	}))
	defer server.Close()

	// Create a client that uses the test server
	client := NewClient("test_api_key", WithURL(server.URL))

	// Test ExecuteGraphQL with errors
	resp, err := client.ExecuteGraphQL("query {}", nil)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	if resp == nil {
		t.Fatal("Expected a response, but got nil")
	}

	if len(resp.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(resp.Errors))
	}

	if resp.Errors[0].Message != "Test GraphQL error" {
		t.Errorf("Expected error message to be 'Test GraphQL error', got '%s'", resp.Errors[0].Message)
	}
}
