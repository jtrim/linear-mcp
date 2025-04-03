package linear

import (
	"fmt"
)

// User represents a Linear user
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetViewer returns information about the authenticated user
func (c *Client) GetViewer() (*User, error) {
	query := `query {
		viewer {
			id
			name
			email
		}
	}`

	resp, err := c.ExecuteGraphQL(query, nil)
	if err != nil {
		return nil, err
	}

	// Extract the viewer data
	viewerData, ok := resp.Data["viewer"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid viewer data format")
	}

	user := &User{
		ID:    viewerData["id"].(string),
		Name:  viewerData["name"].(string),
		Email: viewerData["email"].(string),
	}

	return user, nil
}
