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
                    pushedAt
                    primaryLanguage {
                        name
                    }
                    readme1: object(expression: "HEAD:README.md") { ...BlobText }
					readme2: object(expression: "HEAD:Readme.md") { ...BlobText }
					readme3: object(expression: "HEAD:readme.md") { ...BlobText }
					readme4: object(expression: "HEAD:README") { ...BlobText }
					readme5: object(expression: "HEAD:Readme") { ...BlobText }
					readme6: object(expression: "HEAD:readme") { ...BlobText }
					readme7: object(expression: "HEAD:README.rst") { ...BlobText }
					readme8: object(expression: "HEAD:Readme.rst") { ...BlobText }
					readme9: object(expression: "HEAD:readme.rst") { ...BlobText }
					readme10: object(expression: "HEAD:README.adoc") { ...BlobText }
					readme11: object(expression: "HEAD:Readme.adoc") { ...BlobText }
					readme12: object(expression: "HEAD:readme.adoc") { ...BlobText }
				}
            }
        }
    }
	fragment BlobText on Blob {
		text
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
						UpdatedAt   time.Time `json:"pushedAt"`
						Language    struct {
							Name string `json:"name"`
						} `json:"primaryLanguage"`
						Readme1 struct {
							Text string `json:"text"`
						} `json:"readme1"`
						Readme2 struct {
							Text string `json:"text"`
						} `json:"readme2"`
						Readme3 struct {
							Text string `json:"text"`
						} `json:"readme3"`
						Readme4 struct {
							Text string `json:"text"`
						} `json:"readme4"`
						Readme5 struct {
							Text string `json:"text"`
						} `json:"readme5"`
						Readme6 struct {
							Text string `json:"text"`
						} `json:"readme6"`
						Readme7 struct {
							Text string `json:"text"`
						} `json:"readme7"`
						Readme8 struct {
							Text string `json:"text"`
						} `json:"readme8"`
						Readme9 struct {
							Text string `json:"text"`
						} `json:"readme9"`
						Readme10 struct {
							Text string `json:"text"`
						} `json:"readme10"`
						Readme11 struct {
							Text string `json:"text"`
						} `json:"readme11"`
						Readme12 struct {
							Text string `json:"text"`
						} `json:"readme12"`
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
		readmeContent := ""
		for _, readme := range []struct {
			Text string `json:"text"`
		}{
			node.Readme1,
			node.Readme2,
			node.Readme3,
			node.Readme4,
			node.Readme5,
			node.Readme6,
			node.Readme7,
			node.Readme8,
			node.Readme9,
			node.Readme10,
			node.Readme11,
			node.Readme12,
		} {
			if readme.Text != "" {
				readmeContent = readme.Text
				break
			}
		}
		repo := Repository{
			Name:        node.Name,
			Description: node.Description,
			URL:         node.URL,
			Stars:       node.Stars,
			Forks:       node.Forks,
			Language:    node.Language.Name,
			Readme:      readmeContent,
			UpdatedAt:   node.UpdatedAt,
		}
		repos = append(repos, repo)
	}

	return sortRepositories(repos), nil
}

func sortRepositories(repos []Repository) []Repository {
	for i := range len(repos) {
		for j := i + 1; j < len(repos); j++ {
			if repos[i].UpdatedAt.Before(repos[j].UpdatedAt) {
				repos[i], repos[j] = repos[j], repos[i]
			}
		}
	}
	return repos
}
