package linear

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// DefaultAPIURL is the default URL for Linear's GraphQL API
const DefaultAPIURL = "https://api.linear.app/graphql"

// Client is a client for interacting with the Linear API
type Client struct {
	apiKey  string
	apiURL  string
	httpCli *http.Client
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithURL sets the API URL for the client
func WithURL(url string) ClientOption {
	return func(c *Client) {
		c.apiURL = url
	}
}

// WithHTTPClient sets the HTTP client for the client
func WithHTTPClient(httpCli *http.Client) ClientOption {
	return func(c *Client) {
		c.httpCli = httpCli
	}
}

// NewClient creates a new Linear API client
func NewClient(apiKey string, opts ...ClientOption) *Client {
	client := &Client{
		apiKey:  apiKey,
		apiURL:  DefaultAPIURL,
		httpCli: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data,omitempty"`
	Errors []GraphQLError         `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
}

// ExecuteGraphQL makes a GraphQL request to the Linear API
func (c *Client) ExecuteGraphQL(query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var result GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		if resp.StatusCode != http.StatusOK {
			// If we can't decode the response body and status is not OK,
			// return the HTTP error instead
			return nil, fmt.Errorf("received non-OK response: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// If we have GraphQL errors, format them nicely
	if len(result.Errors) > 0 {
		errorMsgs := make([]string, 0, len(result.Errors))
		for _, err := range result.Errors {
			errorMsgs = append(errorMsgs, err.Message)
		}
		return &result, fmt.Errorf("GraphQL errors: %s", strings.Join(errorMsgs, "; "))
	}

	return &result, nil
}

// Helper functions for safely extracting values from maps
func safeGetString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

func safeGetInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}

func safeGetFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return 0
}
