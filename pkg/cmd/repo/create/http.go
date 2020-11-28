package create

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cli/cli/api"
)

// repoCreateInput represents input parameters for repoCreate
type repoCreateInput struct {
	Name        string `json:"name"`
	Visibility  string `json:"visibility"`
	HomepageURL string `json:"homepageUrl,omitempty"`
	Description string `json:"description,omitempty"`

	OwnerID string `json:"ownerId,omitempty"`
	TeamID  string `json:"teamId,omitempty"`

	HasIssuesEnabled bool `json:"hasIssuesEnabled"`
	HasWikiEnabled   bool `json:"hasWikiEnabled"`
}

type repoTemplateInput struct {
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
	OwnerID    string `json:"ownerId,omitempty"`

	RepositoryID string `json:"repositoryId,omitempty"`
	Description  string `json:"description,omitempty"`
}

// repoCreate creates a new GitHub repository
func repoCreate(client *http.Client, hostname string, input repoCreateInput, templateRepositoryID string) (*api.Repository, error) {
	apiClient := api.NewClientFromHTTP(client)

	if input.TeamID != "" {
		orgID, teamID, err := resolveOrganizationTeam(apiClient, hostname, input.OwnerID, input.TeamID)
		if err != nil {
			return nil, err
		}
		input.TeamID = teamID
		input.OwnerID = orgID
	} else if input.OwnerID != "" {
		orgID, err := resolveOrganization(apiClient, hostname, input.OwnerID)
		if err != nil {
			return nil, err
		}
		input.OwnerID = orgID
	}

	if templateRepositoryID != "" {
		var response struct {
			CloneTemplateRepository struct {
				Repository api.Repository
			}
		}

		if input.OwnerID == "" {
			var err error
			input.OwnerID, err = api.CurrentUserID(apiClient, hostname)
			if err != nil {
				return nil, err
			}
		}

		//cdl--templateInput := repoTemplateInput{
		//cdl--	Name:         input.Name,
		//cdl--	Visibility:   input.Visibility,
		//cdl--	OwnerID:      input.OwnerID,
		//cdl--	RepositoryID: templateRepositoryID,
		//cdl--}

		//cdl--variables := map[string]interface{}{
		//cdl--	"input": templateInput,
		//cdl--}

		//cdl--		err := apiClient.GraphQL(hostname, `
		//cdl--		mutation CloneTemplateRepository($input: CloneTemplateRepositoryInput!) {
		//cdl--			cloneTemplateRepository(input: $input) {
		//cdl--				repository {
		//cdl--					id
		//cdl--					name
		//cdl--					owner { login }
		//cdl--					url
		//cdl--				}
		//cdl--			}
		//cdl--		}
		//cdl--		`, variables, &response)
		//cdl--if err != nil {
		//cdl--	return nil, err
		//cdl--}

		return api.InitRepoHostname(&response.CloneTemplateRepository.Repository, hostname), nil
	}

	//cdl--var response struct {
	//cdl--	CreateRepository struct {
	//cdl--		Repository api.Repository
	//cdl--	}
	//cdl--}

	//cdl--variables := map[string]interface{}{
	//cdl--	"input": input,
	//cdl--}

	//cdl--	err := apiClient.GraphQL(hostname, `
	//cdl--	mutation RepositoryCreate($input: CreateRepositoryInput!) {
	//cdl--		createRepository(input: $input) {
	//cdl--			repository {
	//cdl--				id
	//cdl--				name
	//cdl--				owner { login }
	//cdl--				url
	//cdl--			}
	//cdl--		}
	//cdl--	}
	//cdl--	`, variables, &response)
	// dilandry : Not caring about response for now
	info := repoTemplateInput{
		Name:       input.Name,
		Visibility: input.Visibility,
	}

	var response struct {
		NodeID       string `json:"node_id"`
		Organization struct {
			NodeID string `json:"node_id"`
		}
	}

	requestByte, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	requestBody := bytes.NewReader(requestByte)

	//cdl--err = apiClient.REST(hostname, "POST", fmt.Sprintf("users/%s", "dilandry"), requestBody, nil)
	err = apiClient.REST(hostname, "POST", "user/repos", requestBody, &response)
	if err != nil {
		return nil, err
	}

	return api.InitRepoHostname(&response.CreateRepository.Repository, hostname), nil
	//return nil, nil
}

// using API v3 here because the equivalent in GraphQL needs `read:org` scope
func resolveOrganization(client *api.Client, hostname, orgName string) (string, error) {
	var response struct {
		NodeID string `json:"node_id"`
	}
	err := client.REST(hostname, "GET", fmt.Sprintf("users/%s", orgName), nil, &response)
	return response.NodeID, err
}

// using API v3 here because the equivalent in GraphQL needs `read:org` scope
func resolveOrganizationTeam(client *api.Client, hostname, orgName, teamSlug string) (string, string, error) {
	var response struct {
		NodeID       string `json:"node_id"`
		Organization struct {
			NodeID string `json:"node_id"`
		}
	}
	err := client.REST(hostname, "GET", fmt.Sprintf("orgs/%s/teams/%s", orgName, teamSlug), nil, &response)
	return response.Organization.NodeID, response.NodeID, err
}
