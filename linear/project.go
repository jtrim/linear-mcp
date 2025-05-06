package linear

import (
	"fmt"
)

// Project represents a Linear project
type Project struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Icon        string        `json:"icon,omitempty"`
	Color       string        `json:"color,omitempty"`
	State       string        `json:"state,omitempty"`
	Status      *ProjectStatus `json:"status,omitempty"`
	Lead        *User         `json:"lead,omitempty"`
	Teams       []Team        `json:"teams,omitempty"`
	Issues      []Issue       `json:"issues,omitempty"`
	CreatedAt   string        `json:"createdAt"`
	UpdatedAt   string        `json:"updatedAt,omitempty"`
	StartedAt   string        `json:"startedAt,omitempty"`
	TargetDate  string        `json:"targetDate,omitempty"`
	SortOrder   float64       `json:"sortOrder,omitempty"`
	URL         string        `json:"url,omitempty"`
}

// GetProjectsOptions contains optional parameters for listing projects
type GetProjectsOptions struct {
	First int    // Number of projects to fetch (max 100)
	State string // Filter by project state (started, planned, paused, completed, canceled)
}

// GetProjects returns all projects in the Linear workspace with optional filtering
func (c *Client) GetProjects(opts *GetProjectsOptions) ([]Project, error) {
	variables := map[string]interface{}{}

	first := 50
	if opts != nil && opts.First > 0 && opts.First <= 100 {
		first = opts.First
	}
	variables["first"] = first

	var filterClause string
	if opts != nil && opts.State != "" {
		filterClause = ", filter: { state: { eq: \"" + opts.State + "\" } }"
	}

	query := fmt.Sprintf(`query GetProjects($first: Int!) {
		projects(first: $first%s) {
			nodes {
				id
				name
				description
				icon
				color
				state
				createdAt
				updatedAt
				startedAt
				targetDate
				sortOrder
				url
				lead {
					id
					name
					email
				}
				teams {
					nodes {
						id
						name
						key
					}
				}
			}
		}
	}`, filterClause)

	resp, err := c.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	projectsData, ok := resp.Data["projects"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid projects data format")
	}

	nodesData, ok := projectsData["nodes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid projects nodes format")
	}

	projects := make([]Project, 0, len(nodesData))
	for _, node := range nodesData {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid project node format")
		}

		project := Project{
			ID:          safeGetString(nodeMap, "id"),
			Name:        safeGetString(nodeMap, "name"),
			Description: safeGetString(nodeMap, "description"),
			Icon:        safeGetString(nodeMap, "icon"),
			Color:       safeGetString(nodeMap, "color"),
			State:       safeGetString(nodeMap, "state"),
			CreatedAt:   safeGetString(nodeMap, "createdAt"),
			UpdatedAt:   safeGetString(nodeMap, "updatedAt"),
			StartedAt:   safeGetString(nodeMap, "startedAt"),
			TargetDate:  safeGetString(nodeMap, "targetDate"),
			SortOrder:   safeGetFloat64(nodeMap, "sortOrder"),
			URL:         safeGetString(nodeMap, "url"),
		}

		if leadMap, ok := nodeMap["lead"].(map[string]interface{}); ok {
			project.Lead = &User{
				ID:    safeGetString(leadMap, "id"),
				Name:  safeGetString(leadMap, "name"),
				Email: safeGetString(leadMap, "email"),
			}
		}

		if teamsMap, ok := nodeMap["teams"].(map[string]interface{}); ok {
			if teamsNodes, ok := teamsMap["nodes"].([]interface{}); ok {
				teams := make([]Team, 0, len(teamsNodes))
				for _, teamNode := range teamsNodes {
					teamMap, ok := teamNode.(map[string]interface{})
					if !ok {
						continue
					}
					team := Team{
						ID:   safeGetString(teamMap, "id"),
						Name: safeGetString(teamMap, "name"),
						Key:  safeGetString(teamMap, "key"),
					}
					teams = append(teams, team)
				}
				project.Teams = teams
			}
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// GetProject returns details of a specific project by ID
func (c *Client) GetProject(projectID string) (*Project, error) {
	variables := map[string]interface{}{
		"id": projectID,
	}

	query := `query GetProject($id: String!) {
		project(id: $id) {
			id
			name
			description
			icon
			color
			state
			createdAt
			updatedAt
			startedAt
			targetDate
			sortOrder
			url
			lead {
				id
				name
				email
			}
			teams {
				nodes {
					id
					name
					key
				}
			}
			issues {
				nodes {
					id
					identifier
					title
				}
			}
		}
	}`

	resp, err := c.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	projectData, ok := resp.Data["project"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid project data format")
	}

	project := &Project{
		ID:          safeGetString(projectData, "id"),
		Name:        safeGetString(projectData, "name"),
		Description: safeGetString(projectData, "description"),
		Icon:        safeGetString(projectData, "icon"),
		Color:       safeGetString(projectData, "color"),
		State:       safeGetString(projectData, "state"),
		CreatedAt:   safeGetString(projectData, "createdAt"),
		UpdatedAt:   safeGetString(projectData, "updatedAt"),
		StartedAt:   safeGetString(projectData, "startedAt"),
		TargetDate:  safeGetString(projectData, "targetDate"),
		SortOrder:   safeGetFloat64(projectData, "sortOrder"),
		URL:         safeGetString(projectData, "url"),
	}

	if leadMap, ok := projectData["lead"].(map[string]interface{}); ok {
		project.Lead = &User{
			ID:    safeGetString(leadMap, "id"),
			Name:  safeGetString(leadMap, "name"),
			Email: safeGetString(leadMap, "email"),
		}
	}

	if teamsMap, ok := projectData["teams"].(map[string]interface{}); ok {
		if teamsNodes, ok := teamsMap["nodes"].([]interface{}); ok {
			teams := make([]Team, 0, len(teamsNodes))
			for _, teamNode := range teamsNodes {
				teamMap, ok := teamNode.(map[string]interface{})
				if !ok {
					continue
				}
				team := Team{
					ID:   safeGetString(teamMap, "id"),
					Name: safeGetString(teamMap, "name"),
					Key:  safeGetString(teamMap, "key"),
				}
				teams = append(teams, team)
			}
			project.Teams = teams
		}
	}

	return project, nil
}

// CreateProjectInput represents input for creating a new project
type CreateProjectInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Color       string   `json:"color,omitempty"`
	State       string   `json:"state,omitempty"` // planned, started, paused, completed, canceled
	TeamIDs     []string `json:"teamIds,omitempty"`
	LeadID      string   `json:"leadId,omitempty"`
	StartDate   string   `json:"startDate,omitempty"`  // ISO date format
	TargetDate  string   `json:"targetDate,omitempty"` // ISO date format
}

// CreateProject creates a new project in Linear
func (c *Client) CreateProject(input CreateProjectInput) (*Project, error) {
	// Build the input object
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"name":        input.Name,
			"description": input.Description,
		},
	}

	// Add optional fields to the input object
	inputObj := variables["input"].(map[string]interface{})

	if input.Icon != "" {
		inputObj["icon"] = input.Icon
	}

	if input.Color != "" {
		inputObj["color"] = input.Color
	}

	if input.State != "" {
		inputObj["state"] = input.State
	}

	if len(input.TeamIDs) > 0 {
		inputObj["teamIds"] = input.TeamIDs
	}

	if input.LeadID != "" {
		inputObj["leadId"] = input.LeadID
	}

	if input.StartDate != "" {
		inputObj["startDate"] = input.StartDate
	}

	if input.TargetDate != "" {
		inputObj["targetDate"] = input.TargetDate
	}

	query := `mutation CreateProject($input: ProjectCreateInput!) {
		projectCreate(input: $input) {
			success
			project {
				id
				name
				description
				icon
				color
				state
				createdAt
				updatedAt
				startedAt
				targetDate
				sortOrder
				url
				lead {
					id
					name
					email
				}
				teams {
					nodes {
						id
						name
						key
					}
				}
			}
		}
	}`

	resp, err := c.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	projectCreateData, ok := resp.Data["projectCreate"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid projectCreate data format")
	}

	success, ok := projectCreateData["success"].(bool)
	if !ok || !success {
		return nil, fmt.Errorf("project creation was not successful")
	}

	projectData, ok := projectCreateData["project"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid project data format")
	}

	project := &Project{
		ID:          safeGetString(projectData, "id"),
		Name:        safeGetString(projectData, "name"),
		Description: safeGetString(projectData, "description"),
		Icon:        safeGetString(projectData, "icon"),
		Color:       safeGetString(projectData, "color"),
		State:       safeGetString(projectData, "state"),
		CreatedAt:   safeGetString(projectData, "createdAt"),
		UpdatedAt:   safeGetString(projectData, "updatedAt"),
		StartedAt:   safeGetString(projectData, "startedAt"),
		TargetDate:  safeGetString(projectData, "targetDate"),
		SortOrder:   safeGetFloat64(projectData, "sortOrder"),
		URL:         safeGetString(projectData, "url"),
	}

	if leadMap, ok := projectData["lead"].(map[string]interface{}); ok {
		project.Lead = &User{
			ID:    safeGetString(leadMap, "id"),
			Name:  safeGetString(leadMap, "name"),
			Email: safeGetString(leadMap, "email"),
		}
	}

	if teamsMap, ok := projectData["teams"].(map[string]interface{}); ok {
		if teamsNodes, ok := teamsMap["nodes"].([]interface{}); ok {
			teams := make([]Team, 0, len(teamsNodes))
			for _, teamNode := range teamsNodes {
				teamMap, ok := teamNode.(map[string]interface{})
				if !ok {
					continue
				}
				team := Team{
					ID:   safeGetString(teamMap, "id"),
					Name: safeGetString(teamMap, "name"),
					Key:  safeGetString(teamMap, "key"),
				}
				teams = append(teams, team)
			}
			project.Teams = teams
		}
	}

	return project, nil
}

// ProjectWithIssues represents a Linear project with its issues
type ProjectWithIssues struct {
	ID     string        `json:"id"`
	Status *ProjectStatus `json:"status,omitempty"`
	Issues []Issue       `json:"issues"`
}

// GetProjectIssuesOptions contains optional parameters for fetching project issues
type GetProjectIssuesOptions struct {
	First int // Number of issues to fetch (max 100)
}

// GetProjectIssues returns issues for a specific project
func (c *Client) GetProjectIssues(projectID string, opts *GetProjectIssuesOptions) (*ProjectWithIssues, error) {
	query, err := getGraphQLQuery("get_project_issues.graphql")
	if err != nil {
		return nil, fmt.Errorf("failed to load GraphQL query: %w", err)
	}

	variables := map[string]interface{}{
		"projectId": projectID,
	}

	resp, err := c.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	// Extract the project data
	projectData, ok := resp.Data["project"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid project data format")
	}

	project := &ProjectWithIssues{
		ID: safeGetString(projectData, "id"),
	}

	// Extract status if present
	if statusMap, ok := projectData["status"].(map[string]interface{}); ok {
		project.Status = &ProjectStatus{
			ID:   safeGetString(statusMap, "id"),
			Name: safeGetString(statusMap, "name"),
		}
	}

	// Extract issues data
	if issuesData, ok := projectData["issues"].(map[string]interface{}); ok {
		if nodesData, ok := issuesData["nodes"].([]interface{}); ok {
			issues := make([]Issue, 0, len(nodesData))
			
			for _, node := range nodesData {
				nodeMap, ok := node.(map[string]interface{})
				if !ok {
					continue
				}

				issue := Issue{
					ID:          safeGetString(nodeMap, "id"),
					Title:       safeGetString(nodeMap, "title"),
					Description: safeGetString(nodeMap, "description"),
				}

				// Extract assignee if present
				if assigneeMap, ok := nodeMap["assignee"].(map[string]interface{}); ok {
					issue.Assignee = &User{
						ID:    safeGetString(assigneeMap, "id"),
						Name:  safeGetString(assigneeMap, "name"),
						Email: safeGetString(assigneeMap, "email"),
					}
				}

				issues = append(issues, issue)
			}
			
			project.Issues = issues
		}
	}

	return project, nil
}

// UpdateProjectInput represents input for updating an existing project
type UpdateProjectInput struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Icon        *string  `json:"icon,omitempty"`
	Color       *string  `json:"color,omitempty"`
	State       *string  `json:"state,omitempty"` // planned, started, paused, completed, canceled
	TeamIDs     []string `json:"teamIds,omitempty"`
	LeadID      *string  `json:"leadId,omitempty"`
	StartDate   *string  `json:"startDate,omitempty"`  // ISO date format
	TargetDate  *string  `json:"targetDate,omitempty"` // ISO date format
}

// UpdateProject updates an existing project in Linear
func (c *Client) UpdateProject(projectID string, input UpdateProjectInput) (*Project, error) {
	// Build the input object
	variables := map[string]interface{}{
		"id":    projectID,
		"input": map[string]interface{}{},
	}

	// Add fields to the input object only if they are provided
	inputObj := variables["input"].(map[string]interface{})

	if input.Name != nil {
		inputObj["name"] = *input.Name
	}

	if input.Description != nil {
		inputObj["description"] = *input.Description
	}

	if input.Icon != nil {
		inputObj["icon"] = *input.Icon
	}

	if input.Color != nil {
		inputObj["color"] = *input.Color
	}

	if input.State != nil {
		inputObj["state"] = *input.State
	}

	if len(input.TeamIDs) > 0 {
		inputObj["teamIds"] = input.TeamIDs
	}

	if input.LeadID != nil {
		inputObj["leadId"] = *input.LeadID
	}

	if input.StartDate != nil {
		inputObj["startDate"] = *input.StartDate
	}

	if input.TargetDate != nil {
		inputObj["targetDate"] = *input.TargetDate
	}

	query := `mutation UpdateProject($id: String!, $input: ProjectUpdateInput!) {
		projectUpdate(id: $id, input: $input) {
			success
			project {
				id
				name
				description
				icon
				color
				state
				createdAt
				updatedAt
				startedAt
				targetDate
				sortOrder
				url
				lead {
					id
					name
					email
				}
				teams {
					nodes {
						id
						name
						key
					}
				}
			}
		}
	}`

	resp, err := c.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	projectUpdateData, ok := resp.Data["projectUpdate"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid projectUpdate data format")
	}

	success, ok := projectUpdateData["success"].(bool)
	if !ok || !success {
		return nil, fmt.Errorf("project update was not successful")
	}

	projectData, ok := projectUpdateData["project"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid project data format")
	}

	project := &Project{
		ID:          safeGetString(projectData, "id"),
		Name:        safeGetString(projectData, "name"),
		Description: safeGetString(projectData, "description"),
		Icon:        safeGetString(projectData, "icon"),
		Color:       safeGetString(projectData, "color"),
		State:       safeGetString(projectData, "state"),
		CreatedAt:   safeGetString(projectData, "createdAt"),
		UpdatedAt:   safeGetString(projectData, "updatedAt"),
		StartedAt:   safeGetString(projectData, "startedAt"),
		TargetDate:  safeGetString(projectData, "targetDate"),
		SortOrder:   safeGetFloat64(projectData, "sortOrder"),
		URL:         safeGetString(projectData, "url"),
	}

	if leadMap, ok := projectData["lead"].(map[string]interface{}); ok {
		project.Lead = &User{
			ID:    safeGetString(leadMap, "id"),
			Name:  safeGetString(leadMap, "name"),
			Email: safeGetString(leadMap, "email"),
		}
	}

	if teamsMap, ok := projectData["teams"].(map[string]interface{}); ok {
		if teamsNodes, ok := teamsMap["nodes"].([]interface{}); ok {
			teams := make([]Team, 0, len(teamsNodes))
			for _, teamNode := range teamsNodes {
				teamMap, ok := teamNode.(map[string]interface{})
				if !ok {
					continue
				}
				team := Team{
					ID:   safeGetString(teamMap, "id"),
					Name: safeGetString(teamMap, "name"),
					Key:  safeGetString(teamMap, "key"),
				}
				teams = append(teams, team)
			}
			project.Teams = teams
		}
	}

	return project, nil
}
