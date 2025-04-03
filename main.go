package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"

	"github.com/jtrim/linear-mcp/linear"
)

// Get Issue Arguments
type GetIssueArguments struct {
	ID string `json:"id" jsonschema:"required,description=The Linear issue ID to fetch"`
	IncludeChildren bool `json:"include_children" jsonschema:"description=Whether to include children (sub-issues) in the response"`
}

// Get Team Issues Arguments
type GetTeamIssuesArguments struct {
	TeamID string `json:"team_id" jsonschema:"required,description=The Linear team ID to fetch issues for"`
	First  int    `json:"first" jsonschema:"description=Number of issues to fetch (max 100)"`
}

// Create Issue Arguments
type CreateIssueArguments struct {
	TeamID      string `json:"team_id" jsonschema:"required,description=The Linear team ID to create the issue in"`
	Title       string `json:"title" jsonschema:"required,description=The title of the issue"`
	Description string `json:"description" jsonschema:"description=The description of the issue"`
	Priority    int    `json:"priority" jsonschema:"description=The priority of the issue (1-4)"`
	StateID     string `json:"state_id" jsonschema:"description=The state ID for the issue"`
	AssigneeID  string `json:"assignee_id" jsonschema:"description=The user ID to assign the issue to"`
	ProjectID   string `json:"project_id" jsonschema:"description=The project ID to associate the issue with"`
	ParentID    string `json:"parent_id" jsonschema:"description=The parent issue ID to create this as a sub-issue of"`
}

// Update Issue Arguments
type UpdateIssueArguments struct {
	IssueID     string  `json:"issue_id" jsonschema:"required,description=The Linear issue ID to update"`
	Title       *string `json:"title" jsonschema:"description=The new title for the issue"`
	Description *string `json:"description" jsonschema:"description=The new description for the issue"`
	Priority    *int    `json:"priority" jsonschema:"description=The new priority for the issue (1-4)"`
	StateID     *string `json:"state_id" jsonschema:"description=The new state ID for the issue"`
	AssigneeID  *string `json:"assignee_id" jsonschema:"description=The new assignee user ID"`
	ProjectID   *string `json:"project_id" jsonschema:"description=The new project ID"`
	ParentID    *string `json:"parent_id" jsonschema:"description=The new parent issue ID"`
}

// Get Issue Children Arguments
type GetIssueChildrenArguments struct {
	IssueID string `json:"issue_id" jsonschema:"required,description=The Linear parent issue ID to fetch children for"`
	First   int    `json:"first" jsonschema:"description=Number of children to fetch (max 100)"`
}

// Create Project Arguments
type CreateProjectArguments struct {
	Name        string   `json:"name" jsonschema:"required,description=The name of the project"`
	Description string   `json:"description" jsonschema:"description=The description of the project"`
	Icon        string   `json:"icon" jsonschema:"description=The icon for the project"`
	Color       string   `json:"color" jsonschema:"description=The color for the project"`
	State       string   `json:"state" jsonschema:"description=The state of the project (planned, started, paused, completed, canceled)"`
	TeamIDs     []string `json:"team_ids" jsonschema:"description=The team IDs to associate with the project"`
	LeadID      string   `json:"lead_id" jsonschema:"description=The user ID of the project lead"`
}

// Get Teams Arguments
type GetTeamsArguments struct{}

func main() {
	// Load API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatalf("LINEAR_API_KEY environment variable is required")
	}
	
	// Create Linear client
	client := linear.NewClient(apiKey)

	// Set up MCP server
	server := mcp_golang.NewServer(stdio.NewStdioServerTransport())

	// Register getIssue tool
	err := server.RegisterTool("get_issue", "Get a Linear issue by ID", func(args GetIssueArguments) (*mcp_golang.ToolResponse, error) {
		opts := &linear.GetIssueOptions{
			IncludeChildren: args.IncludeChildren,
		}
		
		issue, err := client.GetIssue(args.ID, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to get issue: %w", err)
		}
		
		jsonData, err := json.MarshalIndent(issue, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal issue to JSON: %w", err)
		}
		
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(jsonData))), nil
	})
	if err != nil {
		log.Fatalf("Failed to register get_issue tool: %v", err)
	}

	// Register getTeamIssues tool
	err = server.RegisterTool("get_team_issues", "Get issues for a Linear team", func(args GetTeamIssuesArguments) (*mcp_golang.ToolResponse, error) {
		opts := &linear.GetTeamIssuesOptions{
			First: args.First,
		}
		
		issues, err := client.GetTeamIssues(args.TeamID, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to get team issues: %w", err)
		}
		
		jsonData, err := json.MarshalIndent(issues, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal issues to JSON: %w", err)
		}
		
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(jsonData))), nil
	})
	if err != nil {
		log.Fatalf("Failed to register get_team_issues tool: %v", err)
	}

	// Register createIssue tool
	err = server.RegisterTool("create_issue", "Create a new Linear issue", func(args CreateIssueArguments) (*mcp_golang.ToolResponse, error) {
		input := linear.CreateIssueInput{
			TeamID:      args.TeamID,
			Title:       args.Title,
			Description: args.Description,
			Priority:    args.Priority,
			StateID:     args.StateID,
			AssigneeID:  args.AssigneeID,
			ProjectID:   args.ProjectID,
			ParentID:    args.ParentID,
		}
		
		issue, err := client.CreateIssue(input)
		if err != nil {
			return nil, fmt.Errorf("failed to create issue: %w", err)
		}
		
		jsonData, err := json.MarshalIndent(issue, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal issue to JSON: %w", err)
		}
		
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(jsonData))), nil
	})
	if err != nil {
		log.Fatalf("Failed to register create_issue tool: %v", err)
	}

	// Register updateIssue tool
	err = server.RegisterTool("update_issue", "Update an existing Linear issue", func(args UpdateIssueArguments) (*mcp_golang.ToolResponse, error) {
		input := linear.UpdateIssueInput{
			Title:       args.Title,
			Description: args.Description,
			Priority:    args.Priority,
			StateID:     args.StateID,
			AssigneeID:  args.AssigneeID,
			ProjectID:   args.ProjectID,
			ParentID:    args.ParentID,
		}
		
		issue, err := client.UpdateIssue(args.IssueID, input)
		if err != nil {
			return nil, fmt.Errorf("failed to update issue: %w", err)
		}
		
		jsonData, err := json.MarshalIndent(issue, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal issue to JSON: %w", err)
		}
		
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(jsonData))), nil
	})
	if err != nil {
		log.Fatalf("Failed to register update_issue tool: %v", err)
	}

	// Register getIssueChildren tool
	err = server.RegisterTool("get_issue_children", "Get sub-issues for a Linear issue", func(args GetIssueChildrenArguments) (*mcp_golang.ToolResponse, error) {
		opts := &linear.GetIssueChildrenOptions{
			First: args.First,
		}
		
		children, err := client.GetIssueChildren(args.IssueID, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to get issue children: %w", err)
		}
		
		jsonData, err := json.MarshalIndent(children, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal children to JSON: %w", err)
		}
		
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(jsonData))), nil
	})
	if err != nil {
		log.Fatalf("Failed to register get_issue_children tool: %v", err)
	}

	// Register createProject tool
	err = server.RegisterTool("create_project", "Create a new Linear project", func(args CreateProjectArguments) (*mcp_golang.ToolResponse, error) {
		input := linear.CreateProjectInput{
			Name:        args.Name,
			Description: args.Description,
			Icon:        args.Icon,
			Color:       args.Color,
			State:       args.State,
			TeamIDs:     args.TeamIDs,
			LeadID:      args.LeadID,
		}
		
		project, err := client.CreateProject(input)
		if err != nil {
			return nil, fmt.Errorf("failed to create project: %w", err)
		}
		
		jsonData, err := json.MarshalIndent(project, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal project to JSON: %w", err)
		}
		
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(jsonData))), nil
	})
	if err != nil {
		log.Fatalf("Failed to register create_project tool: %v", err)
	}

	// Register getTeams tool
	err = server.RegisterTool("get_teams", "Get all Linear teams", func(args GetTeamsArguments) (*mcp_golang.ToolResponse, error) {
		teams, err := client.GetTeams()
		if err != nil {
			return nil, fmt.Errorf("failed to get teams: %w", err)
		}
		
		jsonData, err := json.MarshalIndent(teams, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal teams to JSON: %w", err)
		}
		
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(jsonData))), nil
	})
	if err != nil {
		log.Fatalf("Failed to register get_teams tool: %v", err)
	}

	// Start the server
	log.Println("Starting Linear MCP server...")
	err = server.Serve()
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}

	// Keep the server running
	for {
		time.Sleep(1 * time.Hour)
	}
}