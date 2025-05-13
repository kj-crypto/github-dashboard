package main

import (
	contribution "github-dashboard/pkg"
	"github-dashboard/pkg/github"
	"github-dashboard/pkg/tui"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN not set")
	}

	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("No username provided")
	}
	username := args[0]
	contributions, err := contribution.GetContributionsFromApi(token, username)
	if err != nil {
		log.Fatal(err)
	}
	calendar := contribution.FormatCalendar(contributions, true)
	repos, err := github.GetRepositories(token, username)
	if err != nil {
		log.Fatal(err)
	}

	m := tui.NewModel(repos, calendar)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

}
