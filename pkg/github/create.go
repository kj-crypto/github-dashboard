package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CreatedRepository struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type GraphQlError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func CreateRepository(token, repoName, repoDescription string, isPublic bool) (*CreatedRepository, error) {
	const mutation = `
	mutation($input: CreateRepositoryInput!) {
		createRepository(input: $input) {
			repository {
				name
				url
			}
		}
	}
	`

	visibility := "PRIVATE"
	if isPublic {
		visibility = "PUBLIC"
	}

	requestBody := map[string]interface{}{
		"query": mutation,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"name":        repoName,
				"description": repoDescription,
				"visibility":  visibility,
			},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Error code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Errors []GraphQlError `json:"errors"`
		Data   struct {
			CreateRepository struct {
				Repository struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"repository"`
			} `json:"createRepository"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL Error: %s (Type: %s)", response.Errors[0].Message, response.Errors[0].Type)
	}

	if response.Data.CreateRepository.Repository.Name == "" {
		return nil, fmt.Errorf("API returned 200 OK, but no repository data was found in response")
	}

	return &CreatedRepository{
		Name: response.Data.CreateRepository.Repository.Name,
		URL:  response.Data.CreateRepository.Repository.URL,
	}, nil
}
