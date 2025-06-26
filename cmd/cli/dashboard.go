package main

import (
	contribution "github-dashboard/pkg"
	"github-dashboard/pkg/tui"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	token := contribution.GetToken()
	if token == "" {
		log.Fatal("GITHUB_TOKEN not set")
	}

	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("No username provided")
	}
	username := args[0]

	m := tui.InitModel(username)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

}
