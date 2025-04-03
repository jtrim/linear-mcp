package linear

import (
	"fmt"
)

// Team represents a Linear team
type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

// GetTeams returns all teams in the Linear workspace
func (c *Client) GetTeams() ([]Team, error) {
	query := `query {
		teams {
			nodes {
				id
				name
				key
			}
		}
	}`

	resp, err := c.ExecuteGraphQL(query, nil)
	if err != nil {
		return nil, err
	}

	// Extract the teams data
	teamsData, ok := resp.Data["teams"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid teams data format")
	}

	nodesData, ok := teamsData["nodes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid teams nodes format")
	}

	teams := make([]Team, 0, len(nodesData))
	for _, node := range nodesData {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid team node format")
		}

		team := Team{
			ID:   nodeMap["id"].(string),
			Name: nodeMap["name"].(string),
			Key:  nodeMap["key"].(string),
		}
		teams = append(teams, team)
	}

	return teams, nil
}
