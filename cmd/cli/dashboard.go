package main

import (
	"fmt"
	contribution "github-dashboard/pkg"
	"github-dashboard/pkg/tui"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func setupFileLogger() {
	debug := os.Getenv("GITHUB_DASHBOARD_DEBUG")
	debugEnabled := debug == "on" || debug == "true" || debug == "1"

	if !debugEnabled {
		log.SetOutput(io.Discard)
		return
	}

	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
		return
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logsDir, fmt.Sprintf("ui_debug_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}

	log.SetOutput(file)
	log.Printf("UI logging initialized - %s", time.Now().Format(time.RFC3339))
}

func main() {
	setupFileLogger()

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
