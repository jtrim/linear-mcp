package linear

import (
	"fmt"
)

// WorkflowState represents a Linear workflow state
type WorkflowState struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Issue represents a Linear issue
type Issue struct {
	ID          string         `json:"id"`
	Identifier  string         `json:"identifier"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	State       *WorkflowState `json:"state,omitempty"`
	Assignee    *User          `json:"assignee,omitempty"`
	Project     *Project       `json:"project,omitempty"`
	Parent      *Issue         `json:"parent,omitempty"`
	Children    []Issue        `json:"children,omitempty"`
	Priority    int            `json:"priority"`
	CreatedAt   string         `json:"createdAt"`
	UpdatedAt   string         `json:"updatedAt,omitempty"`
	URL         string         `json:"url,omitempty"`
	BranchName  string         `json:"branchName,omitempty"`
}

// GetTeamIssuesOptions contains optional parameters for getting team issues
type GetTeamIssuesOptions struct {
	First int // Number of issues to fetch (max 100)
}

// GetTeamIssues returns issues for a specific team
func (c *Client) GetTeamIssues(teamID string, opts *GetTeamIssuesOptions) ([]Issue, error) {
	variables := map[string]interface{}{
		"teamId": teamID,
	}

	first := 50
	if opts != nil && opts.First > 0 && opts.First <= 100 {
		first = opts.First
	}
	variables["first"] = first

	query, err := getGraphQLQuery("get_team_issues.graphql")
	if err != nil {
		return nil, fmt.Errorf("failed to load GetTeamIssues query: %w", err)
	}

	resp, err := c.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	teamData, ok := resp.Data["team"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid team data format")
	}

	issuesData, ok := teamData["issues"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid issues data format")
	}

	nodesData, ok := issuesData["nodes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid issues nodes format")
	}

	issues := make([]Issue, 0, len(nodesData))
	for _, node := range nodesData {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid issue node format")
		}

		issue := Issue{
			ID:          safeGetString(nodeMap, "id"),
			Identifier:  safeGetString(nodeMap, "identifier"),
			Title:       safeGetString(nodeMap, "title"),
			Description: safeGetString(nodeMap, "description"),
			Priority:    safeGetInt(nodeMap, "priority"),
			CreatedAt:   safeGetString(nodeMap, "createdAt"),
			UpdatedAt:   safeGetString(nodeMap, "updatedAt"),
			BranchName:  safeGetString(nodeMap, "branchName"),
		}

		if stateMap, ok := nodeMap["state"].(map[string]interface{}); ok {
			issue.State = &WorkflowState{
				ID:   safeGetString(stateMap, "id"),
				Name: safeGetString(stateMap, "name"),
			}
		}

		if assigneeMap, ok := nodeMap["assignee"].(map[string]interface{}); ok {
			issue.Assignee = &User{
				ID:    safeGetString(assigneeMap, "id"),
				Name:  safeGetString(assigneeMap, "name"),
				Email: safeGetString(assigneeMap, "email"),
			}
		}

		issues = append(issues, issue)
	}

	return issues, nil
}

// GetIssueOptions contains optional parameters for getting issue details
type GetIssueOptions struct {
	IncludeChildren bool // Whether to include children (sub-issues) in the response
	ChildrenFirst   int  // Number of children to fetch (max 100)
}

// GetIssue returns details of a specific issue by ID
func (c *Client) GetIssue(issueID string, opts *GetIssueOptions) (*Issue, error) {
	variables := map[string]interface{}{
		"id": issueID,
	}

	query, err := getGraphQLQuery("get_issue.graphql")
	if err != nil {
		return nil, fmt.Errorf("failed to load GetIssue query: %w", err)
	}

	resp, err := c.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	issueData, ok := resp.Data["issue"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid issue data format")
	}

	issue := &Issue{
		ID:          safeGetString(issueData, "id"),
		Identifier:  safeGetString(issueData, "identifier"),
		Title:       safeGetString(issueData, "title"),
		Description: safeGetString(issueData, "description"),
		Priority:    safeGetInt(issueData, "priority"),
		CreatedAt:   safeGetString(issueData, "createdAt"),
		UpdatedAt:   safeGetString(issueData, "updatedAt"),
		URL:         safeGetString(issueData, "url"),
		BranchName:  safeGetString(issueData, "branchName"),
	}

	if stateMap, ok := issueData["state"].(map[string]interface{}); ok {
		issue.State = &WorkflowState{
			ID:   safeGetString(stateMap, "id"),
			Name: safeGetString(stateMap, "name"),
		}
	}

	if assigneeMap, ok := issueData["assignee"].(map[string]interface{}); ok {
		issue.Assignee = &User{
			ID:    safeGetString(assigneeMap, "id"),
			Name:  safeGetString(assigneeMap, "name"),
			Email: safeGetString(assigneeMap, "email"),
		}
	}

	if parentMap, ok := issueData["parent"].(map[string]interface{}); ok {
		issue.Parent = &Issue{
			ID:         safeGetString(parentMap, "id"),
			Identifier: safeGetString(parentMap, "identifier"),
			Title:      safeGetString(parentMap, "title"),
		}
	}

	// If IncludeChildren is true, fetch and populate the children
	if opts != nil && opts.IncludeChildren {
		childrenOpts := &GetIssueChildrenOptions{
			First: 50, // Default to 50 children
		}
		
		if opts.ChildrenFirst > 0 && opts.ChildrenFirst <= 100 {
			childrenOpts.First = opts.ChildrenFirst
		}
		
		children, err := c.GetIssueChildren(issueID, childrenOpts)
		if err != nil {
			return issue, fmt.Errorf("failed to load children: %w", err)
		}
		
		issue.Children = children
	}

	return issue, nil
}

// CreateIssueInput represents input for creating a new issue
type CreateIssueInput struct {
	TeamID      string `json:"teamId"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Priority    int    `json:"priority,omitempty"`
	StateID     string `json:"stateId,omitempty"`
	AssigneeID  string `json:"assigneeId,omitempty"`
	ProjectID   string `json:"projectId,omitempty"` // Optional project ID to associate the issue with
	ParentID    string `json:"parentId,omitempty"`  // Optional parent issue ID to create a sub-issue
}

// CreateIssue creates a new issue in Linear
func (c *Client) CreateIssue(input CreateIssueInput) (*Issue, error) {
	// Build the input object
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"teamId":      input.TeamID,
			"title":       input.Title,
			"description": input.Description,
		},
	}

	// Add optional fields to the input object
	inputObj := variables["input"].(map[string]interface{})

	if input.Priority > 0 {
		inputObj["priority"] = input.Priority
	}

	if input.StateID != "" {
		inputObj["stateId"] = input.StateID
	}

	if input.AssigneeID != "" {
		inputObj["assigneeId"] = input.AssigneeID
	}

	if input.ProjectID != "" {
		inputObj["projectId"] = input.ProjectID
	}

	if input.ParentID != "" {
		inputObj["parentId"] = input.ParentID
	}

	query, err := getGraphQLQuery("create_issue.graphql")
	if err != nil {
		return nil, fmt.Errorf("failed to load CreateIssue query: %w", err)
	}

	resp, err := c.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	issueCreateData, ok := resp.Data["issueCreate"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid issueCreate data format")
	}

	success, ok := issueCreateData["success"].(bool)
	if !ok || !success {
		return nil, fmt.Errorf("issue creation was not successful")
	}

	issueData, ok := issueCreateData["issue"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid issue data format")
	}

	issue := &Issue{
		ID:          safeGetString(issueData, "id"),
		Identifier:  safeGetString(issueData, "identifier"),
		Title:       safeGetString(issueData, "title"),
		Description: safeGetString(issueData, "description"),
		Priority:    safeGetInt(issueData, "priority"),
		URL:         safeGetString(issueData, "url"),
		BranchName:  safeGetString(issueData, "branchName"),
	}

	if stateMap, ok := issueData["state"].(map[string]interface{}); ok {
		issue.State = &WorkflowState{
			ID:   safeGetString(stateMap, "id"),
			Name: safeGetString(stateMap, "name"),
		}
	}

	if assigneeMap, ok := issueData["assignee"].(map[string]interface{}); ok {
		issue.Assignee = &User{
			ID:   safeGetString(assigneeMap, "id"),
			Name: safeGetString(assigneeMap, "name"),
		}
	}

	if projectMap, ok := issueData["project"].(map[string]interface{}); ok {
		issue.Project = &Project{
			ID:   safeGetString(projectMap, "id"),
			Name: safeGetString(projectMap, "name"),
		}
	}
	
	if parentMap, ok := issueData["parent"].(map[string]interface{}); ok {
		issue.Parent = &Issue{
			ID:         safeGetString(parentMap, "id"),
			Identifier: safeGetString(parentMap, "identifier"),
			Title:      safeGetString(parentMap, "title"),
		}
	}

	return issue, nil
}

// UpdateIssueInput represents input for updating an issue
type UpdateIssueInput struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Priority    *int    `json:"priority,omitempty"`
	StateID     *string `json:"stateId,omitempty"`
	AssigneeID  *string `json:"assigneeId,omitempty"`
	ProjectID   *string `json:"projectId,omitempty"` // Optional project ID to associate the issue with
	ParentID    *string `json:"parentId,omitempty"`  // Optional parent issue ID to update parent-child relationship
}

// GetIssueChildrenOptions contains optional parameters for getting issue children
type GetIssueChildrenOptions struct {
	First int // Number of sub-issues to fetch (max 100)
}

// GetIssueChildren returns child issues (sub-issues) for a specific issue
func (c *Client) GetIssueChildren(issueID string, opts *GetIssueChildrenOptions) ([]Issue, error) {
	variables := map[string]interface{}{
		"id": issueID,
	}

	first := 50
	if opts != nil && opts.First > 0 && opts.First <= 100 {
		first = opts.First
	}
	variables["first"] = first

	query, err := getGraphQLQuery("get_issue_children.graphql")
	if err != nil {
		return nil, fmt.Errorf("failed to load GetIssueChildren query: %w", err)
	}

	resp, err := c.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	issueData, ok := resp.Data["issue"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid issue data format")
	}

	childrenData, ok := issueData["children"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid children data format")
	}

	nodesData, ok := childrenData["nodes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid children nodes format")
	}

	children := make([]Issue, 0, len(nodesData))
	for _, node := range nodesData {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid child node format")
		}

		child := Issue{
			ID:          safeGetString(nodeMap, "id"),
			Identifier:  safeGetString(nodeMap, "identifier"),
			Title:       safeGetString(nodeMap, "title"),
			Description: safeGetString(nodeMap, "description"),
			Priority:    safeGetInt(nodeMap, "priority"),
			CreatedAt:   safeGetString(nodeMap, "createdAt"),
			UpdatedAt:   safeGetString(nodeMap, "updatedAt"),
			BranchName:  safeGetString(nodeMap, "branchName"),
		}

		if stateMap, ok := nodeMap["state"].(map[string]interface{}); ok {
			child.State = &WorkflowState{
				ID:   safeGetString(stateMap, "id"),
				Name: safeGetString(stateMap, "name"),
			}
		}

		if assigneeMap, ok := nodeMap["assignee"].(map[string]interface{}); ok {
			child.Assignee = &User{
				ID:    safeGetString(assigneeMap, "id"),
				Name:  safeGetString(assigneeMap, "name"),
				Email: safeGetString(assigneeMap, "email"),
			}
		}

		children = append(children, child)
	}

	return children, nil
}

// UpdateIssue updates an existing issue in Linear
func (c *Client) UpdateIssue(issueID string, input UpdateIssueInput) (*Issue, error) {
	// Build the input object
	variables := map[string]interface{}{
		"id":    issueID,
		"input": map[string]interface{}{},
	}

	// Add optional fields to the input object
	inputObj := variables["input"].(map[string]interface{})

	if input.Title != nil {
		inputObj["title"] = *input.Title
	}

	if input.Description != nil {
		inputObj["description"] = *input.Description
	}

	if input.Priority != nil {
		inputObj["priority"] = *input.Priority
	}

	if input.StateID != nil {
		inputObj["stateId"] = *input.StateID
	}

	if input.AssigneeID != nil {
		inputObj["assigneeId"] = *input.AssigneeID
	}

	if input.ProjectID != nil {
		inputObj["projectId"] = *input.ProjectID
	}

	if input.ParentID != nil {
		inputObj["parentId"] = *input.ParentID
	}

	query, err := getGraphQLQuery("update_issue.graphql")
	if err != nil {
		return nil, fmt.Errorf("failed to load UpdateIssue query: %w", err)
	}

	resp, err := c.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	issueUpdateData, ok := resp.Data["issueUpdate"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid issueUpdate data format")
	}

	success, ok := issueUpdateData["success"].(bool)
	if !ok || !success {
		return nil, fmt.Errorf("issue update was not successful")
	}

	issueData, ok := issueUpdateData["issue"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid issue data format")
	}

	issue := &Issue{
		ID:          safeGetString(issueData, "id"),
		Identifier:  safeGetString(issueData, "identifier"),
		Title:       safeGetString(issueData, "title"),
		Description: safeGetString(issueData, "description"),
		Priority:    safeGetInt(issueData, "priority"),
		URL:         safeGetString(issueData, "url"),
		BranchName:  safeGetString(issueData, "branchName"),
	}

	if stateMap, ok := issueData["state"].(map[string]interface{}); ok {
		issue.State = &WorkflowState{
			ID:   safeGetString(stateMap, "id"),
			Name: safeGetString(stateMap, "name"),
		}
	}

	if assigneeMap, ok := issueData["assignee"].(map[string]interface{}); ok {
		issue.Assignee = &User{
			ID:   safeGetString(assigneeMap, "id"),
			Name: safeGetString(assigneeMap, "name"),
		}
	}

	if projectMap, ok := issueData["project"].(map[string]interface{}); ok {
		issue.Project = &Project{
			ID:   safeGetString(projectMap, "id"),
			Name: safeGetString(projectMap, "name"),
		}
	}
	
	if parentMap, ok := issueData["parent"].(map[string]interface{}); ok {
		issue.Parent = &Issue{
			ID:         safeGetString(parentMap, "id"),
			Identifier: safeGetString(parentMap, "identifier"),
			Title:      safeGetString(parentMap, "title"),
		}
	}

	return issue, nil
}
