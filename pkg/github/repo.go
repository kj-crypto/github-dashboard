package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Repository struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Stars       int       `json:"stargazerCount"`
	Forks       int       `json:"forkCount"`
	Language    string    `json:"primaryLanguage"`
	Readme      string    `json:"readme"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func GetRepositories(token, username string) ([]Repository, error) {
	const query = `
    query($username: String!) {
        user(login: $username) {
            repositories(first: 100) {
                nodes {
                    name
                    description
                    url
                    stargazerCount
                    forkCount
					updatedAt
                    primaryLanguage {
                        name
                    }
                    object(expression: "HEAD:README.md") {
                        ... on Blob {
                            text
                        }
                    }
                }
            }
        }
    }
    `

	requestBody := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"username": username,
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
		return nil, fmt.Errorf("expected status code %d, got %d body: %s", http.StatusOK, resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Data struct {
			User struct {
				Repositories struct {
					Nodes []struct {
						Name        string    `json:"name"`
						Description string    `json:"description"`
						URL         string    `json:"url"`
						Stars       int       `json:"stargazerCount"`
						Forks       int       `json:"forkCount"`
						UpdatedAt   time.Time `json:"updatedAt"`
						Language    struct {
							Name string `json:"name"`
						} `json:"primaryLanguage"`
						Object struct {
							Text string `json:"text"`
						} `json:"object"`
					} `json:"nodes"`
				} `json:"repositories"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var repos []Repository
	for _, node := range response.Data.User.Repositories.Nodes {
		repo := Repository{
			Name:        node.Name,
			Description: node.Description,
			URL:         node.URL,
			Stars:       node.Stars,
			Forks:       node.Forks,
			Language:    node.Language.Name,
			Readme:      node.Object.Text,
			UpdatedAt:   node.UpdatedAt,
		}
		repos = append(repos, repo)
	}

	return sortRepositories(repos), nil
}

func sortRepositories(repos []Repository) []Repository {
	// sort by updated at
	for i := 0; i < len(repos); i++ {
		for j := i + 1; j < len(repos); j++ {
			if repos[i].UpdatedAt.Before(repos[j].UpdatedAt) {
				repos[i], repos[j] = repos[j], repos[i]
			}
		}
	}
	return repos
}
