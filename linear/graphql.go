package linear

import (
	"embed"
	"path/filepath"
)

//go:embed graphql/*.graphql
var graphqlFS embed.FS

// getGraphQLQuery loads a GraphQL query by filename
func getGraphQLQuery(filename string) (string, error) {
	filename = filepath.Base(filename)
	data, err := graphqlFS.ReadFile("graphql/" + filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
