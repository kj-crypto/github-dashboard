package contribution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const query = `
query($username: String!) {
        user(login: $username) {
            contributionsCollection {
                contributionCalendar {
                    totalContributions
                    weeks {
                        contributionDays {
                            contributionCount
                            date
                            weekday
                        }
                    }
                }
            }
        }
    }
`

type ContributionDay struct {
	ContributionCount uint64 `json:"contributionCount"`
	Month             uint8  `json:"date"`
	Weekday           uint8  `json:"weekday"`
}

var MonthAbreviations = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

func (c *ContributionDay) GetMonthAbreviation() string {
	return MonthAbreviations[c.Month]
}

func (c *ContributionDay) GetPreviousMonthAbreviation() string {
	if c.Month == 0 {
		return MonthAbreviations[11]
	}
	return MonthAbreviations[c.Month-1]
}

func parseContributions(raw map[string]interface{}) ([]ContributionDay, error) {
	raw, ok := raw["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected 'data' keyword, got %T", raw["data"])
	}
	raw, ok = raw["user"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected 'user', got %T", raw["user"])
	}
	raw, ok = raw["contributionsCollection"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected 'contributionsCollection', got %T", raw["contributionsCollection"])
	}
	raw, ok = raw["contributionCalendar"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected 'contributionsCalendar', got %T", raw["contributionCalendar"])
	}
	var contributions []ContributionDay
	for _, week := range raw["weeks"].([]interface{}) {
		for _, day := range week.(map[string]interface{})["contributionDays"].([]interface{}) {
			date := day.(map[string]interface{})["date"].(string)
			month, err := time.Parse("2006-01-02", date)
			if err != nil {
				return nil, err
			}
			contributions = append(contributions, ContributionDay{
				ContributionCount: uint64(day.(map[string]interface{})["contributionCount"].(float64)),
				Month:             uint8(month.Month()) - 1,
				Weekday:           uint8(day.(map[string]interface{})["weekday"].(float64)),
			})
		}
	}
	return contributions, nil
}

func MakeContributionMatrix(contributions []ContributionDay) [][]ContributionDay {
	matrix := make([][]ContributionDay, 7)
	for _, contribution := range contributions {

		row := matrix[contribution.Weekday]
		if row == nil {
			row = make([]ContributionDay, 0)
		}
		row = append(row, contribution)
		matrix[contribution.Weekday] = row
	}
	return matrix
}

func GetContributionsFromApi(token, username string) ([][]ContributionDay, error) {
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
		return nil, fmt.Errorf("non 200; status: %d body: %s", resp.StatusCode, string(bodyBytes))
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	contributions, err := parseContributions(raw)
	if err != nil {
		return nil, err
	}

	return MakeContributionMatrix(contributions), nil
}
