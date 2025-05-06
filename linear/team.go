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

// ProjectStatus represents a project status in Linear
type ProjectStatus struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TeamProject represents a project within a team in Linear
type TeamProject struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	SlugID string        `json:"slugId"`
	Status *ProjectStatus `json:"status,omitempty"`
}

// GetTeamProjectsOptions contains optional parameters for listing team projects
type GetTeamProjectsOptions struct {
	First int // Number of projects to fetch (max 100)
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

// GetTeamProjects returns all projects for a specific team
func (c *Client) GetTeamProjects(teamID string, opts *GetTeamProjectsOptions) ([]TeamProject, error) {
	query, err := getGraphQLQuery("get_team_projects.graphql")
	if err != nil {
		return nil, fmt.Errorf("failed to load GraphQL query: %w", err)
	}

	variables := map[string]interface{}{
		"teamId": teamID,
	}

	resp, err := c.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	// Extract the team data
	teamData, ok := resp.Data["team"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid team data format")
	}

	// Extract the projects data
	projectsData, ok := teamData["projects"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid projects data format")
	}

	// Extract the nodes data
	nodesData, ok := projectsData["nodes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid projects nodes format")
	}

	projects := make([]TeamProject, 0, len(nodesData))
	for _, node := range nodesData {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid project node format")
		}

		project := TeamProject{
			ID:     safeGetString(nodeMap, "id"),
			Name:   safeGetString(nodeMap, "name"),
			SlugID: safeGetString(nodeMap, "slugId"),
		}

		// Extract status if present
		if statusMap, ok := nodeMap["status"].(map[string]interface{}); ok {
			project.Status = &ProjectStatus{
				ID:   safeGetString(statusMap, "id"),
				Name: safeGetString(statusMap, "name"),
			}
		}

		projects = append(projects, project)
	}

	return projects, nil
}
